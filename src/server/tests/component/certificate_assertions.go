package component

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"server/certificates"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func ExtractLeafCertificateDerBytesFromBundle(t *testing.T, pemBundleBytes []byte) []byte {
	certPem, _, err := certificates.SplitPemBundle(pemBundleBytes)
	assert.Nil(t, err)

	rest := certPem
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			certificate, err := x509.ParseCertificate(block.Bytes)
			assert.Nil(t, err)
			return certificate.Raw
		}
	}

	assert.True(t, false)
	return nil
}

func GetServerLeafCertificateDerBytes(t *testing.T) []byte {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // #nosec G402: certificate tests intentionally inspect the local test certificate
		ServerName:         "localhost",
	}

	tlsConnection, err := tls.Dial("tcp", "localhost:443", tlsConfig)
	assert.Nil(t, err)
	defer u.Close(tlsConnection)

	connectionState := tlsConnection.ConnectionState()
	assert.True(t, len(connectionState.PeerCertificates) > 0)

	return connectionState.PeerCertificates[0].Raw
}

func AssertServerUsesCertificateBundle(t *testing.T, pemBundleBytes []byte) {
	assert.Equal(t, ExtractLeafCertificateDerBytesFromBundle(t, pemBundleBytes), GetServerLeafCertificateDerBytes(t))
}
