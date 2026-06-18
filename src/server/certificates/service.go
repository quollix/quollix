package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"server/configs"
	"time"

	u "github.com/quollix/common/utils"
)

const (
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

	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		return nil, u.Logger.NewError("invalid ACME account key PEM")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	signer, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, u.Logger.NewError("ACME key is not a *rsa.PrivateKey")
	}

	return signer, nil
}

func GenerateRsaKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 3072)
}
