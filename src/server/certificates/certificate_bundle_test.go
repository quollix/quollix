package certificates

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestNewCertificateBundleFromPemBytes_Rsa_Success(t *testing.T) {
	certificatePemBytes, keyPemBytes := createSelfSignedRsaCertificateAndKeyPemBytes(t)
	pemBundleBytes := bytes.Join([][]byte{certificatePemBytes, keyPemBytes}, []byte("\n"))

	bundle, err := NewCertificateBundleFromPemBytes(pemBundleBytes)
	assert.Nil(t, err)
	assert.NotNil(t, bundle)
	assert.Equal(t, pemBundleBytes, bundle.GetBytes())
	assert.Equal(t, string(pemBundleBytes), bundle.GetString())
	assert.NotNil(t, bundle.GetTlsCertificate())
}

func TestNewCertificateBundleFromString_Rsa_Success(t *testing.T) {
	certificatePemBytes, keyPemBytes := createSelfSignedRsaCertificateAndKeyPemBytes(t)
	pemBundleString := string(bytes.Join([][]byte{certificatePemBytes, keyPemBytes}, []byte("\n")))

	bundle, err := NewCertificateBundleFromString(pemBundleString)
	assert.Nil(t, err)
	assert.NotNil(t, bundle)
	assert.Equal(t, pemBundleString, bundle.GetString())
	assert.NotNil(t, bundle.GetTlsCertificate())
}

func TestNewCertificateBundleFromPemBytes_InvalidPem_ReturnsInvalidPemBundleError(t *testing.T) {
	invalidPemBytes := []byte("not a pem bundle")
	bundle, err := NewCertificateBundleFromPemBytes(invalidPemBytes)

	assert.Nil(t, bundle)
	assert.Equal(t, "pem bundle contains no certificate blocks", u.ExtractError(err))
}

func TestSplitPemBundle_Rsa_Success(t *testing.T) {
	expectedCertificatePemBytes, expectedKeyPemBytes := createSelfSignedRsaCertificateAndKeyPemBytes(t)
	expectedPemBundleBytes := bytes.Join([][]byte{expectedCertificatePemBytes, expectedKeyPemBytes}, []byte("\n"))

	actualCertificatePem, actualKeyPem, err := SplitPemBundle(expectedPemBundleBytes)
	assert.Nil(t, err)
	assert.Equal(t, joinCertificatesFromBundle(expectedPemBundleBytes), actualCertificatePem)
	assert.Equal(t, expectedKeyPemBytes, actualKeyPem)
}

func TestSplitPemBundle_Ecdsa_Success(t *testing.T) {
	certificatePemBytes, keyPemBytes := createSelfSignedEcdsaCertificateAndKeyPemBytes(t)
	pemBundleBytes := bytes.Join([][]byte{certificatePemBytes, keyPemBytes}, []byte("\n"))

	certificatePem, keyPem, err := SplitPemBundle(pemBundleBytes)
	assert.Nil(t, err)
	assert.Equal(t, joinCertificatesFromBundle(pemBundleBytes), certificatePem)
	assert.Equal(t, keyPemBytes, keyPem)
}

func TestSplitPemBundle_Ed25519_Success(t *testing.T) {
	certificatePemBytes, keyPemBytes := createSelfSignedEd25519CertificateAndKeyPemBytes(t)
	pemBundleBytes := bytes.Join([][]byte{certificatePemBytes, keyPemBytes}, []byte("\n"))

	certificatePem, keyPem, err := SplitPemBundle(pemBundleBytes)
	assert.Nil(t, err)
	assert.Equal(t, joinCertificatesFromBundle(pemBundleBytes), certificatePem)
	assert.Equal(t, keyPemBytes, keyPem)
}

func TestSplitPemBundle_MultipleCertificateBlocks_Success(t *testing.T) {
	certificatePemBytes, keyPemBytes := createSelfSignedRsaCertificateAndKeyPemBytes(t)
	pemBundleBytes := bytes.Join([][]byte{certificatePemBytes, certificatePemBytes, keyPemBytes}, []byte("\n"))

	certificatePem, keyPem, err := SplitPemBundle(pemBundleBytes)
	assert.Nil(t, err)
	assert.Equal(t, joinCertificatesFromBundle(pemBundleBytes), certificatePem)
	assert.Equal(t, keyPemBytes, keyPem)
}

func TestSplitPemBundle_MissingCertificateBlocks_Error(t *testing.T) {
	_, keyPemBytes := createSelfSignedRsaCertificateAndKeyPemBytes(t)
	pemBundleBytes := keyPemBytes

	certificatePem, keyPem, err := SplitPemBundle(pemBundleBytes)
	assert.Nil(t, certificatePem)
	assert.Nil(t, keyPem)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "pem bundle contains no certificate blocks"))
}

func TestSplitPemBundle_MissingPrivateKeyBlock_Error(t *testing.T) {
	certificatePemBytes, _ := createSelfSignedRsaCertificateAndKeyPemBytes(t)
	pemBundleBytes := certificatePemBytes

	certificatePem, keyPem, err := SplitPemBundle(pemBundleBytes)
	assert.Nil(t, certificatePem)
	assert.Nil(t, keyPem)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "pem bundle contains no private key block"))
}

func TestSplitPemBundle_MultiplePrivateKeys_Error(t *testing.T) {
	certificatePemBytes, firstKeyPemBytes := createSelfSignedRsaCertificateAndKeyPemBytes(t)
	_, secondKeyPemBytes := createSelfSignedRsaCertificateAndKeyPemBytes(t)
	pemBundleBytes := bytes.Join([][]byte{certificatePemBytes, firstKeyPemBytes, secondKeyPemBytes}, []byte("\n"))

	certificatePem, keyPem, err := SplitPemBundle(pemBundleBytes)
	assert.Nil(t, certificatePem)
	assert.Nil(t, keyPem)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "pem bundle contains multiple private keys"))
}

func TestSplitPemBundle_MismatchedKeyAndCertificate_Error(t *testing.T) {
	certificatePemBytes, _ := createSelfSignedRsaCertificateAndKeyPemBytes(t)
	_, mismatchedKeyPemBytes := createSelfSignedRsaCertificateAndKeyPemBytes(t)
	pemBundleBytes := bytes.Join([][]byte{certificatePemBytes, mismatchedKeyPemBytes}, []byte("\n"))

	certificatePem, keyPem, err := SplitPemBundle(pemBundleBytes)
	assert.Nil(t, certificatePem)
	assert.Nil(t, keyPem)
	assert.NotNil(t, err)
	assert.Equal(t, "private key does not match certificate", u.ExtractError(err))
}

func TestNewCertificateBundleFromPemBytes_MismatchedKeyAndCertificate_ReturnsInvalidPemBundleError(t *testing.T) {
	certificatePemBytes, _ := createSelfSignedRsaCertificateAndKeyPemBytes(t)
	_, mismatchedKeyPemBytes := createSelfSignedRsaCertificateAndKeyPemBytes(t)
	pemBundleBytes := bytes.Join([][]byte{certificatePemBytes, mismatchedKeyPemBytes}, []byte("\n"))

	bundle, err := NewCertificateBundleFromPemBytes(pemBundleBytes)
	assert.Nil(t, bundle)
	assert.NotNil(t, err)
	assert.Equal(t, "private key does not match certificate", u.ExtractError(err))
}

func createSelfSignedRsaCertificateAndKeyPemBytes(t *testing.T) ([]byte, []byte) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.Nil(t, err)

	certificateDerBytes := createSelfSignedCertificateDerBytes(t, privateKey.Public(), privateKey)

	certificatePemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateDerBytes})
	keyPemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return certificatePemBytes, keyPemBytes
}

func createSelfSignedEcdsaCertificateAndKeyPemBytes(t *testing.T) ([]byte, []byte) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.Nil(t, err)

	privateKeyDerBytes, err := x509.MarshalECPrivateKey(privateKey)
	assert.Nil(t, err)

	certificateDerBytes := createSelfSignedCertificateDerBytes(t, privateKey.Public(), privateKey)

	certificatePemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateDerBytes})
	keyPemBytes := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKeyDerBytes})

	return certificatePemBytes, keyPemBytes
}

func createSelfSignedEd25519CertificateAndKeyPemBytes(t *testing.T) ([]byte, []byte) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	assert.Nil(t, err)

	privateKeyDerBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	assert.Nil(t, err)

	certificateDerBytes := createSelfSignedCertificateDerBytes(t, publicKey, privateKey)

	certificatePemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateDerBytes})
	keyPemBytes := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDerBytes})

	return certificatePemBytes, keyPemBytes
}

func createSelfSignedCertificateDerBytes(t *testing.T, publicKey any, privateKey any) []byte {
	return createCertificateDerBytes(
		t,
		publicKey,
		privateKey,
		false,
		x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	)
}

func joinCertificatesFromBundle(pemBundleBytes []byte) []byte {
	var certificateBlocks [][]byte
	rest := pemBundleBytes

	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			certificateBlocks = append(certificateBlocks, pem.EncodeToMemory(block))
		}
	}

	return bytes.Join(certificateBlocks, []byte("\n"))
}

func TestSplitPemBundle_ContainsJunkBlock_Error(t *testing.T) {
	certificatePemBytes, keyPemBytes := createSelfSignedRsaCertificateAndKeyPemBytes(t)

	junkBlockBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "JUNK",
		Bytes: []byte("junk"),
	})

	pemBundleBytes := bytes.Join([][]byte{certificatePemBytes, junkBlockBytes, keyPemBytes}, []byte("\n"))

	certificatePem, keyPem, err := SplitPemBundle(pemBundleBytes)
	assert.Nil(t, certificatePem)
	assert.Nil(t, keyPem)
	assert.NotNil(t, err)
	assert.Equal(t, "pem bundle contains unsupported block type", u.ExtractError(err))
}

func TestSplitPemBundle_CertificateChainWrongOrder_StillMatchesKey_Success(t *testing.T) {
	leafCertificatePemBytes, leafKeyPemBytes := createSelfSignedRsaCertificateAndKeyPemBytes(t)

	caCertificatePemBytes, err := createCaCertificatePemBytes(t)
	assert.Nil(t, err)

	pemBundleBytes := bytes.Join([][]byte{caCertificatePemBytes, leafCertificatePemBytes, leafKeyPemBytes}, []byte("\n"))

	expectedCertificatePemBytes := bytes.Join([][]byte{leafCertificatePemBytes, caCertificatePemBytes}, []byte("\n"))

	certificatePem, keyPem, err := SplitPemBundle(pemBundleBytes)
	assert.Nil(t, err)
	assert.Equal(t, expectedCertificatePemBytes, certificatePem)
	assert.Equal(t, leafKeyPemBytes, keyPem)
}

func createCaCertificatePemBytes(t *testing.T) ([]byte, error) {
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.Nil(t, err)

	certificateDerBytes := createCertificateDerBytes(
		t,
		caPrivateKey.Public(),
		caPrivateKey,
		true,
		x509.KeyUsageCertSign|x509.KeyUsageCRLSign|x509.KeyUsageDigitalSignature,
		[]x509.ExtKeyUsage{},
	)

	certificatePemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateDerBytes})
	return certificatePemBytes, nil
}

func TestNewCertificateBundleFromPemBytes_ReordersAndReturnsNormalizedBytes(t *testing.T) {
	leafCertificatePemBytes, leafKeyPemBytes := createSelfSignedRsaCertificateAndKeyPemBytes(t)
	caCertificatePemBytes, err := createCaCertificatePemBytes(t)
	assert.Nil(t, err)

	pemBundleBytes := bytes.Join([][]byte{caCertificatePemBytes, leafCertificatePemBytes, leafKeyPemBytes}, []byte("\n"))

	bundle, err := NewCertificateBundleFromPemBytes(pemBundleBytes)
	assert.Nil(t, err)

	expectedCertificatePemBytes := bytes.Join([][]byte{leafCertificatePemBytes, caCertificatePemBytes}, []byte("\n"))
	expectedNormalizedPemBundleBytes := bytes.Join([][]byte{expectedCertificatePemBytes, leafKeyPemBytes}, []byte("\n"))

	assert.Equal(t, expectedNormalizedPemBundleBytes, bundle.GetBytes())
}

func createCertificateDerBytes(
	t *testing.T,
	publicKey any,
	privateKey any,
	isCa bool,
	keyUsage x509.KeyUsage,
	extKeyUsage []x509.ExtKeyUsage,
) []byte {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	assert.Nil(t, err)

	certificateTemplate := &x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              keyUsage,
		ExtKeyUsage:           extKeyUsage,
		IsCA:                  isCa,
		BasicConstraintsValid: isCa,
	}

	certificateDerBytes, err := x509.CreateCertificate(rand.Reader, certificateTemplate, certificateTemplate, publicKey, privateKey)
	assert.Nil(t, err)

	return certificateDerBytes
}
