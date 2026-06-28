package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"server/configs"

	u "github.com/quollix/common/utils"
	"golang.org/x/crypto/acme"
)

const CertificatChallengeNotPossibleForLocalhost = "certificate challenge is not possible for base domain 'localhost'"

type WildcardCertificateService interface {
	StartDns01Session() (*Dns01Session, *Dns01ChallengeInfo, error)
	FinishDns01Session(session *Dns01Session, wildcardKeyAuth string)
}

type WildcardCertificateServiceImpl struct {
	ConfigsRepository    configs.ConfigsRepository
	ConfigsService       configs.ConfigsService
	CertificatePersister CertificatePersister
	NetworkWaiter        NetworkWaiter
	AcmeClient           AcmeClient
	CertificateService   CertificateService
	OperationMonitor     OperationMonitor
}

func (w *WildcardCertificateServiceImpl) StartDns01Session() (*Dns01Session, *Dns01ChallengeInfo, error) {
	host, err := w.ConfigsService.GetBaseDomain()
	if err != nil {
		return nil, nil, err
	}

	if host == "localhost" {
		return nil, nil, u.Logger.NewError(CertificatChallengeNotPossibleForLocalhost)
	}

	sessionId := w.OperationMonitor.BeginRun("starting wildcard certificate generation")

	if err := w.registerAcmeAccountIfNotExist(sessionId); err != nil {
		return nil, nil, err
	}

	certificatePrivateKey, err := GenerateRsaKey()
	if err != nil {
		return nil, nil, err
	}

	w.OperationMonitor.SetOperation(sessionId, "creating order")
	order, err := w.createOrder(host)
	if err != nil {
		return nil, nil, err
	}

	w.OperationMonitor.SetOperation(sessionId, "fetching DNS-01 challenge")
	wildcardChallenge, wildcardAuthzUrl, err := w.fetchWildcardChallenge(order)
	if err != nil {
		return nil, nil, err
	}

	wildcardKeyAuth, err := w.AcmeClient.DNS01ChallengeRecord(wildcardChallenge.Token)
	if err != nil {
		return nil, nil, u.Logger.NewError(err.Error())
	}

	session := &Dns01Session{
		Host:           host,
		Order:          order,
		AuthzUrl:       wildcardAuthzUrl,
		Challenge:      wildcardChallenge,
		CertificateKey: certificatePrivateKey,
		Id:             sessionId,
	}

	info := &Dns01ChallengeInfo{
		RecordName:      acmeChallengePrefix + host,
		WildcardKeyAuth: wildcardKeyAuth,
	}

	return session, info, nil
}

func (w *WildcardCertificateServiceImpl) FinishDns01Session(session *Dns01Session, wildcardKeyAuth string) {
	err := w.finishDns01Session(session, wildcardKeyAuth)
	if err != nil {
		w.OperationMonitor.EndRun(session.Id, false, "DNS challenge failed: "+err.Error())
		w.NetworkWaiter.Sleep(5 * time.Minute)
		w.OperationMonitor.Clear(session.Id)
		return
	}

	w.OperationMonitor.EndRun(session.Id, true, "wildcard certificate generation successful")
	w.NetworkWaiter.Sleep(5 * time.Minute)
	w.OperationMonitor.Clear(session.Id)
}

func (w *WildcardCertificateServiceImpl) finishDns01Session(session *Dns01Session, wildcardKeyAuth string) error {
	err := w.NetworkWaiter.WaitForDnsTxtRecord(session, wildcardKeyAuth)
	if err != nil {
		return err
	}

	w.OperationMonitor.SetOperation(session.Id, "accepting challenge")
	_, err = w.AcmeClient.Accept(session.Challenge)
	if err != nil {
		return err
	}

	w.OperationMonitor.SetOperation(session.Id, "waiting for authorization")
	err = w.NetworkWaiter.WaitForAuthorization(session.AuthzUrl)
	if err != nil {
		return err
	}

	csr, err := generateAndValidateCsr(session.Host, session.CertificateKey)
	if err != nil {
		return err
	}

	return w.finalizeAndSaveCertificate(session, csr)
}

func (w *WildcardCertificateServiceImpl) registerAcmeAccountIfNotExist(sessionId int) error {
	isRegistered, err := w.ConfigsRepository.GetConfig(configs.ConfigKeys.AcmeAccountRegistered)
	if err != nil {
		return err
	}
	if isRegistered == "true" {
		return nil
	}
	w.OperationMonitor.SetOperation(sessionId, "registering ACME account")

	account := &acme.Account{}
	_, registrationError := w.AcmeClient.Register(account, func(string) bool { return true })
	if registrationError != nil {
		if acmeError, ok := registrationError.(*acme.Error); ok {
			_ = acmeError
		} else if registrationError.Error() != "acme: account already exists" {
			return u.Logger.NewError(registrationError.Error())
		}
	}

	return w.ConfigsRepository.SetConfig(configs.ConfigKeys.AcmeAccountRegistered, "true")
}

func (w *WildcardCertificateServiceImpl) createOrder(host string) (*acme.Order, error) {
	order, err := w.AcmeClient.AuthorizeOrder([]acme.AuthzID{
		{Type: "dns", Value: "*." + host},
	})
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	if len(order.AuthzURLs) < 1 {
		return nil, u.Logger.NewError("expected at least 1 authorizations for base and wildcard domains")
	}
	return order, nil
}

func (w *WildcardCertificateServiceImpl) fetchWildcardChallenge(order *acme.Order) (*acme.Challenge, string, error) {
	wildcardAuthzUrl := order.AuthzURLs[0]
	wildcardAuthorization, err := w.AcmeClient.GetAuthorization(wildcardAuthzUrl)
	if err != nil {
		return nil, "", u.Logger.NewError(err.Error())
	}

	var wildcardChallenge *acme.Challenge
	for _, challenge := range wildcardAuthorization.Challenges {
		if challenge.Type == "dns-01" {
			wildcardChallenge = challenge
			break
		}
	}
	if wildcardChallenge == nil {
		return nil, "", u.Logger.NewError("DNS-01 challenges not found for base or wildcard domains")
	}

	return wildcardChallenge, wildcardAuthzUrl, nil
}

func generateAndValidateCsr(host string, certificateKey *rsa.PrivateKey) ([]byte, error) {
	wildcardDomain := "*." + host
	csr := &x509.CertificateRequest{
		DNSNames: []string{wildcardDomain},
	}
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, csr, certificateKey)
	if err != nil {
		return nil, u.Logger.NewError("failed to generate CSR")
	}
	return csrBytes, nil
}

func (w *WildcardCertificateServiceImpl) finalizeAndSaveCertificate(session *Dns01Session, csr []byte) error {
	attempt := 0
	var certDerChain [][]byte
	err := w.NetworkWaiter.Retry("Finalize certificate issuance with Let's Encrypt", 1*time.Minute, 10*time.Minute, func() error {
		attempt++
		operation := fmt.Sprintf("finalizing certificate issuance, attempt %d", attempt)
		w.OperationMonitor.SetOperation(session.Id, operation)

		var createError error
		certDerChain, _, createError = w.AcmeClient.CreateOrderCert(session.Order.FinalizeURL, csr, true)
		return createError
	})
	if err != nil {
		return err
	}

	privateKeyDerBytes, err := x509.MarshalPKCS8PrivateKey(session.CertificateKey)
	if err != nil {
		return err
	}

	var pemBundle []byte
	for _, certDer := range certDerChain {
		pemBundle = append(pemBundle, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDer})...)
	}
	pemBundle = append(pemBundle, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDerBytes})...)

	certificateBundle, err := NewCertificateBundleFromPemBytes(pemBundle)
	if err != nil {
		return err
	}

	if err := w.CertificateService.ReplaceCertificate(certificateBundle); err != nil {
		return err
	}

	u.Logger.Info("new certificate was saved and loaded successfully")
	w.OperationMonitor.SetOperation(session.Id, "new certificate was saved and loaded successfully")
	return nil
}
