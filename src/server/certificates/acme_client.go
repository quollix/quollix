package certificates

import (
	"context"

	"golang.org/x/crypto/acme"
)

type AcmeClient interface {
	Register(account *acme.Account, prompt func(string) bool) (*acme.Account, error)
	AuthorizeOrder(ids []acme.AuthzID) (*acme.Order, error)
	GetAuthorization(url string) (*acme.Authorization, error)
	DNS01ChallengeRecord(token string) (string, error)
	Accept(challenge *acme.Challenge) (*acme.Challenge, error)
	CreateOrderCert(finalizeUrl string, csr []byte, bundle bool) (certDerChain [][]byte, certUrl string, err error)
	LoadAcmeAccountClient() error
}

type AcmeClientImpl struct {
	CertificateService CertificateService
	acmeClient         *acme.Client    `wire:"-"`
	ctx                context.Context `wire:"-"`
}

func (a *AcmeClientImpl) Register(account *acme.Account, prompt func(string) bool) (*acme.Account, error) {
	return a.acmeClient.Register(a.ctx, account, prompt)
}

func (a *AcmeClientImpl) AuthorizeOrder(ids []acme.AuthzID) (*acme.Order, error) {
	return a.acmeClient.AuthorizeOrder(a.ctx, ids)
}

func (a *AcmeClientImpl) GetAuthorization(url string) (*acme.Authorization, error) {
	return a.acmeClient.GetAuthorization(a.ctx, url)
}

func (a *AcmeClientImpl) DNS01ChallengeRecord(token string) (string, error) {
	return a.acmeClient.DNS01ChallengeRecord(token)
}

func (a *AcmeClientImpl) Accept(challenge *acme.Challenge) (*acme.Challenge, error) {
	return a.acmeClient.Accept(a.ctx, challenge)
}

func (a *AcmeClientImpl) CreateOrderCert(finalizeUrl string, csr []byte, bundle bool) ([][]byte, string, error) {
	return a.acmeClient.CreateOrderCert(a.ctx, finalizeUrl, csr, bundle)
}

func (a *AcmeClientImpl) LoadAcmeAccountClient() error {
	key, err := a.CertificateService.GetAcmeAccountPrivateKey()
	if err != nil {
		return err
	}
	a.acmeClient = &acme.Client{
		Key:          key,
		DirectoryURL: certificateOrderEndpoint,
	}
	a.ctx = context.Background()
	return nil
}
