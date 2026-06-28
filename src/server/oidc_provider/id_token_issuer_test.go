package oidc_provider

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"server/certificates"
	"server/configs"
	"strings"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

type idTokenIssuerTestObjects struct {
	Issuer      *IdTokenIssuerImpl
	ConfigsRepo *configs.ConfigsRepositoryMock
}

func newIdTokenIssuerTestObjects(t *testing.T) idTokenIssuerTestObjects {
	configsRepo := configs.NewConfigsRepositoryMock(t)
	issuer := &IdTokenIssuerImpl{
		ConfigsRepo: configsRepo,
	}
	return idTokenIssuerTestObjects{
		Issuer:      issuer,
		ConfigsRepo: configsRepo,
	}
}

func generateRsaPrivateKeyPemForIssuerTests(t *testing.T) (*rsa.PrivateKey, string) {
	privateKey, err := certificates.GenerateRsaKey()
	assert.Nil(t, err)

	privateKeyDer := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyDer,
	})
	return privateKey, string(privateKeyPem)
}

func TestIdTokenIssuerImpl_Initialize_WhenKeyExists_LoadsKeyAndDoesNotPersist(t *testing.T) {
	testObjects := newIdTokenIssuerTestObjects(t)

	privateKey, privateKeyPem := generateRsaPrivateKeyPemForIssuerTests(t)
	testObjects.ConfigsRepo.EXPECT().GetConfig(configs.ConfigKeys.OidcPrivateKey).Return(privateKeyPem, nil)

	err := testObjects.Issuer.PrepareJwkToken()
	assert.Nil(t, err)

	jwks := testObjects.Issuer.GetJwks()
	assert.Equal(t, 1, len(jwks.Keys))
	assert.Equal(t, "RSA", jwks.Keys[0].Kty)
	assert.Equal(t, "RS256", jwks.Keys[0].Alg)
	assert.Equal(t, "sig", jwks.Keys[0].Use)
	assert.Equal(t, "key-1", jwks.Keys[0].Kid)

	assert.NotEqual(t, (*rsa.PrivateKey)(nil), testObjects.Issuer.privateKey)
	assert.Equal(t, privateKey.N.String(), testObjects.Issuer.privateKey.N.String())
}

func TestIdTokenIssuerImpl_Initialize_WhenPemInvalid_ReturnsError(t *testing.T) {
	testObjects := newIdTokenIssuerTestObjects(t)

	testObjects.ConfigsRepo.EXPECT().GetConfig(configs.ConfigKeys.OidcPrivateKey).Return("not-a-pem", nil)

	err := testObjects.Issuer.PrepareJwkToken()
	assert.Equal(t, "invalid OIDC RSA private key format", u.ExtractError(err))
}

func TestIdTokenIssuerImpl_Initialize_WhenPemTypeWrong_ReturnsError(t *testing.T) {
	testObjects := newIdTokenIssuerTestObjects(t)

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "NOT RSA PRIVATE KEY",
		Bytes: []byte("whatever"),
	})

	testObjects.ConfigsRepo.EXPECT().GetConfig(configs.ConfigKeys.OidcPrivateKey).Return(string(pemBytes), nil)

	err := testObjects.Issuer.PrepareJwkToken()
	assert.Equal(t, "invalid OIDC RSA private key format", u.ExtractError(err))
}

func TestIdTokenIssuerImpl_Sign_ReturnsJwtWithValidSignatureAndClaims(t *testing.T) {
	testObjects := newIdTokenIssuerTestObjects(t)

	privateKey, err := certificates.GenerateRsaKey()
	assert.Nil(t, err)
	testObjects.Issuer.privateKey = privateKey
	testObjects.Issuer.jwk = jwkFromRSAPrivateKey(privateKey)

	claims := &IDTokenClaims{
		Sub:               "7",
		Iss:               "https://quollix.localhost",
		Aud:               "client-1",
		Nonce:             "nonce-1",
		Role:              "admin",
		Groups:            []string{"admins"},
		Exp:               1234567890,
		Name:              "Test User",
		PreferredUsername: "testuser",
		Email:             "test@example.com",
	}

	token, err := testObjects.Issuer.Sign(claims)
	assert.Nil(t, err)

	parts := strings.Split(token, ".")
	assert.Equal(t, 3, len(parts))

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	assert.Nil(t, err)
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	assert.Nil(t, err)
	signatureBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	assert.Nil(t, err)

	var header map[string]interface{}
	assert.Nil(t, json.Unmarshal(headerBytes, &header))
	assert.Equal(t, "RS256", header["alg"])
	assert.Equal(t, "JWT", header["typ"])
	assert.Equal(t, "key-1", header["kid"])

	var payload map[string]interface{}
	assert.Nil(t, json.Unmarshal(payloadBytes, &payload))

	assert.Equal(t, "7", payload["sub"])
	assert.Equal(t, "https://quollix.localhost", payload["iss"])
	assert.Equal(t, "client-1", payload["aud"])
	assert.Equal(t, "nonce-1", payload["nonce"])
	assert.Equal(t, "admin", payload["role"])
	assert.Equal(t, "Test User", payload["name"])
	assert.Equal(t, "testuser", payload["preferred_username"])
	assert.Equal(t, "test@example.com", payload["email"])

	jwks := testObjects.Issuer.GetJwks()
	pubKey := publicKeyFromJwk(t, jwks.Keys[0])

	unsigned := parts[0] + "." + parts[1]
	hash := sha256.Sum256([]byte(unsigned))
	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], signatureBytes)
	assert.Nil(t, err)
}

func publicKeyFromJwk(t *testing.T, jwk JWK) *rsa.PublicKey {
	nb, err := base64.RawURLEncoding.DecodeString(jwk.N)
	assert.Nil(t, err)
	eb, err := base64.RawURLEncoding.DecodeString(jwk.E)
	assert.Nil(t, err)

	n := new(big.Int).SetBytes(nb)
	eBig := new(big.Int).SetBytes(eb)

	return &rsa.PublicKey{
		N: n,
		E: int(eBig.Int64()),
	}
}
