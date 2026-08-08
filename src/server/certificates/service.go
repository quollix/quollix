package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"server/configs"

	u "github.com/quollix/common/utils"
)

const (
	AcmeAccountPrivateKeyFileName = "acme_account_private_key.pem"

	/*
		- By default, we use the production ACME directory to issue publicly trusted certificates. However, if you test this feature frequently, you may hit the Let's Encrypt rate limit, which could block your work. To avoid this, you can temporarily point the variable below to the staging directory, which issues untrusted test certificates with much more lenient limits.
		- Staging/Test Certificate URL: https://acme-staging-v02.api.letsencrypt.org/directory.
		- Production Certificate URL (should be default): https://acme-v02.api.letsencrypt.org/directory
		- Remember to switch back to production before merging the code.
	*/
	certificateOrderEndpoint = "https://acme-v02.api.letsencrypt.org/directory"
)

type CertificateService interface {
	ReplaceCertificate(certBundle *CertificateBundle) error
	GetCurrentCertificate() (*CertificateBundle, error)
	GenerateUniversalSelfSignedCert() (*CertificateBundle, error)
	GetAcmeAccountPrivateKey() (*rsa.PrivateKey, error)
	GetAcmeAccountPrivateKeyPemBytes() ([]byte, error)
	ReplaceAcmeAccountPrivateKeyPemBytes(privateKeyPemBytes []byte) error
}

type CertificateServiceImpl struct {
	CertificatePersister CertificatePersister
	ConfigsRepository    configs.ConfigsRepository
	CertificateCache     CertificateCache
}

func (s *CertificateServiceImpl) ReplaceCertificate(certBundle *CertificateBundle) error {
	if err := s.ConfigsRepository.SetConfig(configs.ConfigKeys.CertificatePemBundle, certBundle.GetString()); err != nil {
		return err
	}

	s.CertificateCache.SetCertificate(certBundle.GetTlsCertificate())
	return nil
}

func (s *CertificateServiceImpl) GetCurrentCertificate() (*CertificateBundle, error) {
	certString, err := s.ConfigsRepository.GetConfig(configs.ConfigKeys.CertificatePemBundle)
	if err != nil {
		return nil, err
	}

	return NewCertificateBundleFromString(certString)
}

func (c *CertificateServiceImpl) GenerateUniversalSelfSignedCert() (*CertificateBundle, error) {
	privateKey, err := GenerateRsaKey()
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		DNSNames:     []string{"*", "*.*"},
		IPAddresses:  []net.IP{net.IPv4(0, 0, 0, 0), net.IPv6zero},
		IsCA:         false,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDerBytes, err := x509.CreateCertificate(rand.Reader, template, template, privateKey.Public(), privateKey)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}

	certPemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDerBytes})

	privateKeyDerBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	keyPemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDerBytes})

	_, err = tls.X509KeyPair(certPemBytes, keyPemBytes)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}

	pemBundle := append(certPemBytes, keyPemBytes...)
	return NewCertificateBundleFromPemBytes(pemBundle)
}

func (c *CertificateServiceImpl) GetAcmeAccountPrivateKey() (*rsa.PrivateKey, error) {
	pemString, err := c.ConfigsRepository.GetConfig(configs.ConfigKeys.AcmeAccountPrivateKey)
	if err != nil {
		return nil, err
	}

	return parseAcmeAccountPrivateKeyPemBytes([]byte(pemString))
}

func (c *CertificateServiceImpl) GetAcmeAccountPrivateKeyPemBytes() ([]byte, error) {
	pemString, err := c.ConfigsRepository.GetConfig(configs.ConfigKeys.AcmeAccountPrivateKey)
	if err != nil {
		return nil, err
	}
	return []byte(pemString), nil
}

func (c *CertificateServiceImpl) ReplaceAcmeAccountPrivateKeyPemBytes(privateKeyPemBytes []byte) error {
	privateKey, err := parseAcmeAccountPrivateKeyPemBytes(privateKeyPemBytes)
	if err != nil {
		return err
	}
	normalizedPrivateKeyPemBytes, err := marshalAcmeAccountPrivateKeyPemBytes(privateKey)
	if err != nil {
		return err
	}
	return c.ConfigsRepository.SetConfig(configs.ConfigKeys.AcmeAccountPrivateKey, string(normalizedPrivateKeyPemBytes))
}

func GenerateAcmeAccountPrivateKeyPemBytes() ([]byte, error) {
	privateKey, err := GenerateRsaKey()
	if err != nil {
		return nil, err
	}
	return marshalAcmeAccountPrivateKeyPemBytes(privateKey)
}

func parseAcmeAccountPrivateKeyPemBytes(privateKeyPemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyPemBytes)
	if block == nil {
		return nil, u.Logger.NewError("invalid ACME account key PEM")
	}

	if block.Type != "PRIVATE KEY" {
		return nil, u.Logger.NewError("unsupported ACME account key PEM type", "type", block.Type)
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	privateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, u.Logger.NewError("ACME key is not a *rsa.PrivateKey")
	}
	return privateKey, nil
}

func marshalAcmeAccountPrivateKeyPemBytes(privateKey *rsa.PrivateKey) ([]byte, error) {
	privateKeyDerBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDerBytes}), nil
}

func GenerateRsaKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 3072)
}
