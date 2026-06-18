package oidc_provider

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"server/configs"

	u "github.com/quollix/common/utils"
)

type IdTokenIssuer interface {
	PrepareJwkToken() error
	Sign(claims *IDTokenClaims) (string, error)
	GetJwks() JWKS
}

type IdTokenIssuerImpl struct {
	ConfigsRepo configs.ConfigsRepository
	privateKey  *rsa.PrivateKey `wire:"-"`
	jwk         *JWK            `wire:"-"`
}

func (i *IdTokenIssuerImpl) PrepareJwkToken() error {
	var key *rsa.PrivateKey
	keyString, err := i.ConfigsRepo.GetConfig(configs.ConfigKeys.OidcPrivateKey)
	if err != nil {
		return err
	}
	block, _ := pem.Decode([]byte(keyString))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return u.Logger.NewError("invalid OIDC RSA private key format")
	}

	parsedKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	key = parsedKey

	i.privateKey = key
	i.jwk = jwkFromRSAPrivateKey(key)
	return nil
}

func (i *IdTokenIssuerImpl) GetJwks() JWKS {
	return JWKS{
		Keys: []JWK{*i.jwk},
	}
}

func (i *IdTokenIssuerImpl) Sign(claims *IDTokenClaims) (string, error) {
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
		"kid": i.jwk.Kid,
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsBytes)
	unsigned := headerEncoded + "." + claimsEncoded

	hash := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, i.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}

	signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)
	return unsigned + "." + signatureEncoded, nil
}

func jwkFromRSAPrivateKey(key *rsa.PrivateKey) *JWK {
	pub := key.PublicKey
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())

	eBytes := big.NewInt(int64(pub.E)).Bytes()
	eBytes = trimLeadingZeros(eBytes)
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	return &JWK{
		Kty: "RSA",
		Alg: "RS256",
		Use: "sig",
		Kid: "key-1",
		N:   n,
		E:   e,
	}
}
