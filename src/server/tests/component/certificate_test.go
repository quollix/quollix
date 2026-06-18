//go:build component

package component

import (
	"encoding/pem"
	"server/certificates"
	"testing"

	"github.com/quollix/common/assert"
)

func TestCertificateUploadDownloadRoundTrip(t *testing.T) {
	client := GetClientAndLogin(t)
	defer client.Test.ResetTestState()

	certificateService := certificates.CertificateServiceImpl{}
	cert, err := certificateService.GenerateUniversalSelfSignedCert()
	assert.Nil(t, err)

	client.Certificates.Upload(cert.GetBytes())

	downloadedBytes := client.Certificates.DownloadBundleBytes()

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
	client.Certificates.Reset()
	defer client.Test.ResetTestState()

	beforeResetBundleBytes := client.Certificates.DownloadBundleBytes()
	beforeResetDownloadedLeafDerBytes := ExtractLeafCertificateDerBytesFromBundle(t, beforeResetBundleBytes)
	beforeResetServerLeafDerBytes := GetServerLeafCertificateDerBytes(t)

	assert.Equal(t, beforeResetDownloadedLeafDerBytes, beforeResetServerLeafDerBytes)

	client.Certificates.Reset()

	afterResetBundleBytes := client.Certificates.DownloadBundleBytes()
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
