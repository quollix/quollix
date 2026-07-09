package certificates

import (
	"crypto/tls"
	"testing"

	"github.com/quollix/common/assert"
)

var cache = &CertificateCacheImpl{}

func TestGetCertificate_Empty(t *testing.T) {
	cert := cache.GetCertificate()
	assert.Nil(t, cert)
}

func TestSetCertificate_ThenGetCertificate(t *testing.T) {
	expectedCertificate := &tls.Certificate{}

	cache.SetCertificate(expectedCertificate)
	actualCertificate := cache.GetCertificate()

	assert.Equal(t, expectedCertificate, actualCertificate)
}

func TestSetCertificate_Overwrite(t *testing.T) {
	firstCertificate := &tls.Certificate{}
	secondCertificate := &tls.Certificate{}

	cache.SetCertificate(firstCertificate)
	cache.SetCertificate(secondCertificate)
	actualCertificate := cache.GetCertificate()

	assert.Equal(t, secondCertificate, actualCertificate)
}
