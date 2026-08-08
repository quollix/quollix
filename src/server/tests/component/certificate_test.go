//go:build component

package component

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"server/certificates"

	"github.com/quollix/common/assert"
)

func TestCertificateUploadDownloadRoundTrip(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	certificateService := certificates.CertificateServiceImpl{}
	cert, err := certificateService.GenerateUniversalSelfSignedCert()
	assert.Nil(t, err)

	assert.Nil(t, client.Certificates.UploadCertificateBundle(cert.GetBytes()))

	downloadedBytes, err := client.Certificates.DownloadCertificateBundleBytes()
	assert.Nil(t, err)
	AssertBundleParsesAsTlsKeyPair(t, downloadedBytes)

	uploadedLeafDerBytes := ExtractLeafCertificateDerBytesFromBundle(t, cert.GetBytes())
	downloadedLeafDerBytes := ExtractLeafCertificateDerBytesFromBundle(t, downloadedBytes)
	assert.Equal(t, uploadedLeafDerBytes, downloadedLeafDerBytes)

	uploadedKeyDerBytes := extractPrivateKeyDerBytesFromBundle(t, cert.GetBytes())
	downloadedKeyDerBytes := extractPrivateKeyDerBytesFromBundle(t, downloadedBytes)
	assert.Equal(t, uploadedKeyDerBytes, downloadedKeyDerBytes)

	serverLeafDerBytes := GetServerLeafCertificateDerBytes(t)
	assert.Equal(t, downloadedLeafDerBytes, serverLeafDerBytes)
}

func TestCertificateReset(t *testing.T) {
	client := GetClientAndLogin(t)
	// The database reset restores an older certificate state. However, other tests may have loaded different certificates into the in-memory cache beforehand. This can cause a mismatch between:
	//   1) the certificate served by the HTTP server (from cache), and
	//   2) the certificate bundle retrieved from the database.
	// Resetting the certificate ensures both sources are synchronized again.
	assert.Nil(t, client.Certificates.Reset())
	defer client.Test.ResetTestState()

	beforeResetBundleBytes, err := client.Certificates.DownloadCertificateBundleBytes()
	assert.Nil(t, err)
	AssertBundleParsesAsTlsKeyPair(t, beforeResetBundleBytes)
	beforeResetDownloadedLeafDerBytes := ExtractLeafCertificateDerBytesFromBundle(t, beforeResetBundleBytes)
	beforeResetServerLeafDerBytes := GetServerLeafCertificateDerBytes(t)

	assert.Equal(t, beforeResetDownloadedLeafDerBytes, beforeResetServerLeafDerBytes)

	assert.Nil(t, client.Certificates.Reset())

	afterResetBundleBytes, err := client.Certificates.DownloadCertificateBundleBytes()
	assert.Nil(t, err)
	AssertBundleParsesAsTlsKeyPair(t, afterResetBundleBytes)
	afterResetDownloadedLeafDerBytes := ExtractLeafCertificateDerBytesFromBundle(t, afterResetBundleBytes)
	afterResetServerLeafDerBytes := GetServerLeafCertificateDerBytes(t)

	assert.Equal(t, afterResetDownloadedLeafDerBytes, afterResetServerLeafDerBytes)
	assert.NotEqual(t, beforeResetBundleBytes, afterResetBundleBytes)
	assert.NotEqual(t, beforeResetDownloadedLeafDerBytes, afterResetDownloadedLeafDerBytes)
}

func extractPrivateKeyDerBytesFromBundle(t *testing.T, pemBundleBytes []byte) []byte {
	_, keyPem, err := certificates.SplitPemBundle(pemBundleBytes)
	assert.Nil(t, err)

	block, _ := pem.Decode(keyPem)
	assert.NotNil(t, block)

	return block.Bytes
}

func TestStubWildcardCertificateGeneration(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	dnsChallengeInfo, err := client.Certificates.TryDns01Challenge()
	assert.Nil(t, err)
	assert.Equal(t, "_acme-challenge.localhost", dnsChallengeInfo.RecordName)
	assert.Equal(t, certificates.SampleWildcardKeyAuth, dnsChallengeInfo.WildcardKeyAuth)
}

func TestAcmeAccountPrivateKeyUploadDownloadRoundTrip(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	initialFile, err := client.Certificates.DownloadAcmeAccountPrivateKey()
	assert.Nil(t, err)
	assert.Equal(t, certificates.AcmeAccountPrivateKeyFileName, initialFile.FileName)
	initialKey := parseAcmeAccountPrivateKey(t, initialFile.Content)

	pkcs8Key, pkcs8PemBytes := generatePkcs8RsaPrivateKeyPem(t)
	assert.NotEqual(t, initialKey.N.Bytes(), pkcs8Key.N.Bytes())
	assert.Nil(t, client.Certificates.UploadAcmeAccountPrivateKey(pkcs8PemBytes))

	pkcs8DownloadedFile, err := client.Certificates.DownloadAcmeAccountPrivateKey()
	assert.Nil(t, err)
	assert.Equal(t, certificates.AcmeAccountPrivateKeyFileName, pkcs8DownloadedFile.FileName)
	assertAcmeAccountPrivateKeyMatches(t, pkcs8Key, pkcs8DownloadedFile.Content)
	assert.NotEqual(t, initialFile.Content, pkcs8DownloadedFile.Content)
}

func generatePkcs8RsaPrivateKeyPem(t *testing.T) (*rsa.PrivateKey, []byte) {
	privateKey, err := certificates.GenerateRsaKey()
	assert.Nil(t, err)

	privateKeyDerBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	assert.Nil(t, err)
	return privateKey, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDerBytes})
}

func assertAcmeAccountPrivateKeyMatches(t *testing.T, expectedKey *rsa.PrivateKey, actualPemBytes []byte) {
	actualKey := parseAcmeAccountPrivateKey(t, actualPemBytes)

	assert.Equal(t, expectedKey.N.Bytes(), actualKey.N.Bytes())
	assert.Equal(t, expectedKey.E, actualKey.E)
}

func parseAcmeAccountPrivateKey(t *testing.T, pemBytes []byte) *rsa.PrivateKey {
	block, _ := pem.Decode(pemBytes)
	assert.NotNil(t, block)
	assert.Equal(t, "PRIVATE KEY", block.Type)

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	assert.Nil(t, err)
	privateKey, ok := key.(*rsa.PrivateKey)
	assert.True(t, ok)
	return privateKey
}
