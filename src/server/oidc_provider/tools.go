package oidc_provider

import "time"

const (
	authCodeTTL             = 5 * time.Minute
	accessTokenTTL          = 10 * time.Minute
	refreshTokenTTL         = 24 * time.Hour
	accessTokenExpiry int64 = 600 // seconds, matches accessTokenTTL

)

type AuthCodeResult struct {
	Code  string
	State string
	Nonce string
}

type AccessToken struct {
	Token    string
	UserID   string
	ClientID string
	Expiry   time.Time
}

type Client struct {
	ID          string
	Secret      string
	RedirectURI string
}

type AuthCode struct {
	Code                string
	ClientID            string
	UserID              string
	RedirectURI         string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
	Expiry              time.Time
}

// Quollix intentionally returns a fixed opinionated claim set for app integration simplicity.
// Claims are not filtered dynamically based on requested scopes or the optional OIDC claims parameter.

type IDTokenClaims struct {
	Sub    string   `json:"sub"`
	Iss    string   `json:"iss"`
	Aud    string   `json:"aud"`
	Role   string   `json:"role"`
	Groups []string `json:"groups"`
	Exp    int64    `json:"exp"`

	Nonce             string `json:"nonce,omitempty"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	Iat               int64  `json:"iat"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UserinfoResponse struct {
	Sub               string   `json:"sub"`
	Role              string   `json:"role"`
	Groups            []string `json:"groups"`
	Name              string   `json:"name"`
	PreferredUsername string   `json:"preferred_username"`
	Email             string   `json:"email"`
}

type JWK struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type DiscoveryConfig struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
}

type RefreshToken struct {
	Token    string
	UserID   string
	ClientID string
	Expiry   time.Time

	FamilyId string
}

type AuthorizeInput struct {
	ResponseType        string
	ClientID            string
	RedirectURI         string
	State               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
	UserID              int
}

type AuthCodeGrantInput struct {
	Code         string
	RedirectURI  string
	CodeVerifier string
	ClientID     string
	ClientSecret string
}

type RefreshTokenGrantInput struct {
	RefreshToken string
	ClientID     string
	ClientSecret string
}
