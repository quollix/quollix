package certificates

import (
	"crypto/tls"
	"sync"
)

type CertificateCache interface {
	GetCertificate() *tls.Certificate
	SetCertificate(cert *tls.Certificate)
}

type CertificateCacheImpl struct {
	CurrentCertificate        *tls.Certificate
	ReadWriteCertificateMutex sync.RWMutex
}

func (c *CertificateCacheImpl) GetCertificate() *tls.Certificate {
	c.ReadWriteCertificateMutex.RLock()
	defer c.ReadWriteCertificateMutex.RUnlock()
	return c.CurrentCertificate
}

func (c *CertificateCacheImpl) SetCertificate(cert *tls.Certificate) {
	c.ReadWriteCertificateMutex.Lock()
	defer c.ReadWriteCertificateMutex.Unlock()
	c.CurrentCertificate = cert
}
