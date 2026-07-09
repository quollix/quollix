package certificates

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"

	u "github.com/quollix/common/utils"
)

type CertificateBundle struct {
	pemBundleBytes []byte
	tlsCertificate *tls.Certificate
}

func NewCertificateBundleFromString(pemBundleString string) (*CertificateBundle, error) {
	return NewCertificateBundleFromPemBytes([]byte(pemBundleString))
}

func NewCertificateBundleFromPemBytes(pemBundleBytes []byte) (*CertificateBundle, error) {
	certPemBytes, keyPemBytes, err := SplitPemBundle(pemBundleBytes)
	if err != nil {
		return nil, err
	}

	tlsCertificate, err := tls.X509KeyPair(certPemBytes, keyPemBytes)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}

	normalizedPemBundleBytes := bytes.Join([][]byte{certPemBytes, keyPemBytes}, []byte("\n"))

	return &CertificateBundle{
		pemBundleBytes: normalizedPemBundleBytes,
		tlsCertificate: &tlsCertificate,
	}, nil
}

func (c *CertificateBundle) GetBytes() []byte {
	return c.pemBundleBytes
}

func (c *CertificateBundle) GetString() string {
	return string(c.pemBundleBytes)
}

func (c *CertificateBundle) GetTlsCertificate() *tls.Certificate {
	return c.tlsCertificate
}

func SplitPemBundle(pemBundle []byte) (certPem []byte, keyPem []byte, err error) {
	var certificateBlocks [][]byte
	var keyBlockBytes []byte

	rest := pemBundle
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}

		switch block.Type {
		case "CERTIFICATE":
			certificateBlocks = append(certificateBlocks, pem.EncodeToMemory(block))
		case "PRIVATE KEY", "RSA PRIVATE KEY", "EC PRIVATE KEY":
			if len(keyBlockBytes) != 0 {
				return nil, nil, u.Logger.NewError("pem bundle contains multiple private keys")
			}
			keyBlockBytes = pem.EncodeToMemory(block)
		default:
			return nil, nil, u.Logger.NewError("pem bundle contains unsupported block type", "blockType", block.Type)
		}
	}

	if len(certificateBlocks) == 0 {
		return nil, nil, u.Logger.NewError("pem bundle contains no certificate blocks")
	}
	if len(keyBlockBytes) == 0 {
		return nil, nil, u.Logger.NewError("pem bundle contains no private key block")
	}

	reorderedCertificateBlocks, err := reorderCertificateBlocksSoMatchingCertificateIsFirst(certificateBlocks, keyBlockBytes)
	if err != nil {
		return nil, nil, err
	}

	certPemCombined := bytes.Join(reorderedCertificateBlocks, []byte("\n"))
	if len(certPemCombined) > 0 && certPemCombined[len(certPemCombined)-1] != '\n' {
		certPemCombined = append(certPemCombined, '\n')
	}

	if _, err := tls.X509KeyPair(certPemCombined, keyBlockBytes); err != nil {
		return nil, nil, u.Logger.NewError(err.Error())
	}

	return certPemCombined, keyBlockBytes, nil
}

func reorderCertificateBlocksSoMatchingCertificateIsFirst(certificateBlocks [][]byte, keyPem []byte) ([][]byte, error) {
	privateKey, err := parsePrivateKeyFromPemBytes(keyPem)
	if err != nil {
		return nil, err
	}

	matchingIndex := -1
	for certificateIndex, certificatePemBytes := range certificateBlocks {
		certificate, err := parseSingleCertificateFromPemBytes(certificatePemBytes)
		if err != nil {
			return nil, err
		}
		if doesPrivateKeyMatchPublicKey(privateKey, certificate.PublicKey) {
			matchingIndex = certificateIndex
			break
		}
	}

	if matchingIndex == -1 {
		return nil, u.Logger.NewError("private key does not match certificate")
	}

	if matchingIndex == 0 {
		return certificateBlocks, nil
	}

	reordered := make([][]byte, 0, len(certificateBlocks))
	reordered = append(reordered, certificateBlocks[matchingIndex])
	reordered = append(reordered, certificateBlocks[:matchingIndex]...)
	reordered = append(reordered, certificateBlocks[matchingIndex+1:]...)
	return reordered, nil
}

func parseSingleCertificateFromPemBytes(certificatePemBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(certificatePemBytes)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, u.Logger.NewError("failed to parse certificate pem block")
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	return certificate, nil
}

func parsePrivateKeyFromPemBytes(keyPemBytes []byte) (any, error) {
	block, _ := pem.Decode(keyPemBytes)
	if block == nil {
		return nil, u.Logger.NewError("failed to parse private key pem block")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		return privateKey, nil
	case "EC PRIVATE KEY":
		privateKey, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		return privateKey, nil
	case "PRIVATE KEY":
		privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		return privateKey, nil
	default:
		return nil, u.Logger.NewError("unsupported private key type")
	}
}

func doesPrivateKeyMatchPublicKey(privateKey any, publicKey any) bool {
	switch typedPrivateKey := privateKey.(type) {
	case *rsa.PrivateKey:
		typedPublicKey, ok := publicKey.(*rsa.PublicKey)
		return ok && typedPublicKey.N.Cmp(typedPrivateKey.N) == 0 && typedPublicKey.E == typedPrivateKey.E
	case *ecdsa.PrivateKey:
		typedPublicKey, ok := publicKey.(*ecdsa.PublicKey)
		return ok && typedPrivateKey.PublicKey.Equal(typedPublicKey)
	case ed25519.PrivateKey:
		typedPublicKey, ok := publicKey.(ed25519.PublicKey)
		if !ok {
			return false
		}
		derivedPublicKey := typedPrivateKey.Public().(ed25519.PublicKey)
		return bytes.Equal(typedPublicKey, derivedPublicKey)
	default:
		return false
	}
}
