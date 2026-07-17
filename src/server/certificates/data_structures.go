package certificates

import (
	"crypto/rsa"

	"golang.org/x/crypto/acme"
)

type Dns01Session struct {
	Host           string
	Order          *acme.Order
	AuthzUrl       string
	Challenge      *acme.Challenge
	CertificateKey *rsa.PrivateKey
	Id             int
}
