package oidc_provider

import (
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"strconv"
	"strings"

	"server/configs"
	"server/tools"
	"server/users"

	u "github.com/quollix/common/utils"
)

type OidcService interface {
	InitializeOidcService() error
	Authorize(in AuthorizeInput) (string, error)
	TokenWithAuthCode(in AuthCodeGrantInput) (*TokenResponse, error)
	TokenWithRefresh(in RefreshTokenGrantInput) (*TokenResponse, error)
	Userinfo(accessToken string) (*UserinfoResponse, error)
	Discovery() (*DiscoveryConfig, error)
	GetJwks() JWKS
}

type OidcServiceImpl struct {
	Cache          OidcCache
	ConfigsService configs.ConfigsService
	UserRepository users.UserRepository
	AuthHelper     u.AuthHelper
	Clock          Clock
	ClientService  OidcClientService
	TokenIssuer    TokenIssuer
	IdTokenIssuer  IdTokenIssuer
}

func (s *OidcServiceImpl) InitializeOidcService() error {
	return s.IdTokenIssuer.PrepareJwkToken()
}

func (s *OidcServiceImpl) Authorize(in AuthorizeInput) (string, error) {
	if err := validateAuthorizeInput(in); err != nil {
		return "", err
	}
	if in.ResponseType != "code" {
		return "", u.Logger.NewError("unsupported response_type, expected 'code'", "actual", in.ResponseType)
	}
	if in.CodeChallengeMethod != "" && in.CodeChallengeMethod != "S256" {
		return "", u.Logger.NewError("unsupported code_challenge_method, expected 'S256' or empty", "actual", in.CodeChallengeMethod)
	}

	if err := s.ClientService.CheckClientIdAndRedirectUri(in.ClientID, in.RedirectURI); err != nil {
		return "", err
	}

	_, err := s.UserRepository.GetUserById(in.UserID)
	if err != nil {
		return "", err
	}

	code, err := s.AuthHelper.GenerateSecret()
	if err != nil {
		return "", err
	}

	normalizedRedirectUri, err := normalizeRedirectUri(in.RedirectURI)
	if err != nil {
		return "", err
	}

	s.Cache.StoreAuthCode(AuthCode{
		Code:                code,
		ClientID:            in.ClientID,
		UserID:              strconv.Itoa(in.UserID),
		RedirectURI:         normalizedRedirectUri,
		Nonce:               in.Nonce,
		CodeChallenge:       in.CodeChallenge,
		CodeChallengeMethod: in.CodeChallengeMethod,
		Expiry:              s.Clock.Now().Add(authCodeTTL),
	})

	redirect, err := url.Parse(in.RedirectURI)
	if err != nil {
		return "", u.Logger.NewError("invalid redirect_uri format")
	}
	query := redirect.Query()
	query.Set("code", code)
	if in.State != "" {
		query.Set("state", in.State)
	}
	redirect.RawQuery = query.Encode()
	return redirect.String(), nil
}

// normalizeRedirectUri canonicalizes redirect_uri (scheme/host case + default port 443) so authorize and token requests compare equal even if clients format the URL differently.
func normalizeRedirectUri(redirectUri string) (string, error) {
	parsedUrl, err := url.Parse(redirectUri)
	if err != nil {
		return "", err
	}

	scheme := strings.ToLower(parsedUrl.Scheme)
	hostname := strings.ToLower(parsedUrl.Hostname())
	port := parsedUrl.Port()

	if port == "443" || port == "" {
		parsedUrl.Host = hostname
	} else {
		parsedUrl.Host = hostname + ":" + port
	}

	parsedUrl.Scheme = scheme

	return parsedUrl.String(), nil
}

func (s *OidcServiceImpl) TokenWithAuthCode(in AuthCodeGrantInput) (*TokenResponse, error) {
	if err := validateAuthCodeGrantInput(in); err != nil {
		return nil, err
	}

	if err := s.ClientService.CheckClientIdAndRedirectUri(in.ClientID, in.RedirectURI); err != nil {
		return nil, err
	}
	if err := s.ClientService.AuthenticateClient(in.ClientID, in.ClientSecret); err != nil {
		return nil, err
	}

	authCode, ok := s.Cache.ConsumeAuthCode(in.Code)
	if !ok {
		return nil, u.Logger.NewError("grant not found")
	}
	if s.Clock.Now().After(authCode.Expiry) {
		return nil, u.Logger.NewError("grant expired")
	}
	if authCode.ClientID != in.ClientID {
		return nil, u.Logger.NewError("grant client_id mismatch")
	}

	normalizedRedirectUri, err := normalizeRedirectUri(in.RedirectURI)
	if err != nil {
		return nil, err
	}
	if authCode.RedirectURI != normalizedRedirectUri {
		return nil, u.Logger.NewError("grant redirect_uri mismatch")
	}

	if authCode.CodeChallengeMethod != "" {
		if authCode.CodeChallengeMethod != "S256" {
			return nil, u.Logger.NewError("unsupported code_challenge_method in stored auth code, expected 'S256'", "actual", authCode.CodeChallengeMethod)
		}
		if in.CodeVerifier == "" {
			return nil, u.Logger.NewError("grant missing code_verifier")
		}
		sum := sha256.Sum256([]byte(in.CodeVerifier))
		expectedChallenge := base64.RawURLEncoding.EncodeToString(sum[:])
		if expectedChallenge != authCode.CodeChallenge {
			return nil, u.Logger.NewError("invalid code_verifier")
		}
	}

	claims, err := s.ClientService.BuildIdTokenClaims(authCode.UserID, authCode.ClientID, authCode.Nonce)
	if err != nil {
		return nil, err
	}
	idToken, err := s.IdTokenIssuer.Sign(claims)
	if err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := s.TokenIssuer.IssueForUser(authCode.UserID, authCode.ClientID)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		IDToken:      idToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    accessTokenExpiry,
	}, nil
}

func (s *OidcServiceImpl) TokenWithRefresh(in RefreshTokenGrantInput) (*TokenResponse, error) {
	if err := validateRefreshTokenGrantInput(in); err != nil {
		return nil, err
	}

	if err := s.ClientService.AuthenticateClient(in.ClientID, in.ClientSecret); err != nil {
		return nil, err
	}

	storedRefreshToken, err := s.Cache.ConsumeRefreshToken(in.RefreshToken)
	if err != nil {
		return nil, err
	}

	if storedRefreshToken.ClientID != in.ClientID {
		return nil, u.Logger.NewError("refresh token client_id mismatch")
	}
	if s.Clock.Now().After(storedRefreshToken.Expiry) {
		return nil, u.Logger.NewError("refresh token expired")
	}

	claims, err := s.ClientService.BuildIdTokenClaims(storedRefreshToken.UserID, in.ClientID, "")
	if err != nil {
		return nil, err
	}
	idToken, err := s.IdTokenIssuer.Sign(claims)
	if err != nil {
		return nil, err
	}

	newAccessToken, err := s.TokenIssuer.IssueAccessToken(storedRefreshToken.UserID, in.ClientID)
	if err != nil {
		return nil, err
	}

	newRefreshTokenValue, err := s.AuthHelper.GenerateSecret()
	if err != nil {
		return nil, err
	}

	candidateExpiry := s.Clock.Now().Add(refreshTokenTTL)
	newExpiry := storedRefreshToken.Expiry
	if candidateExpiry.Before(newExpiry) {
		// Rotation must not extend the refresh token lifetime beyond the configured TTL window.
		newExpiry = candidateExpiry
	}

	s.Cache.StoreRefreshToken(RefreshToken{
		Token:    newRefreshTokenValue,
		UserID:   storedRefreshToken.UserID,
		ClientID: storedRefreshToken.ClientID,
		Expiry:   newExpiry,
		FamilyId: storedRefreshToken.FamilyId,
	})

	return &TokenResponse{
		AccessToken:  newAccessToken,
		IDToken:      idToken,
		RefreshToken: newRefreshTokenValue,
		TokenType:    "Bearer",
		ExpiresIn:    accessTokenExpiry,
	}, nil
}

func (s *OidcServiceImpl) Userinfo(accessToken string) (*UserinfoResponse, error) {
	if accessToken == "" {
		return nil, u.Logger.NewError("access token missing")
	}

	at, ok := s.Cache.GetAccessToken(accessToken)
	if !ok {
		return nil, u.Logger.NewError("access token not found")
	}
	if s.Clock.Now().After(at.Expiry) {
		return nil, u.Logger.NewError("access token expired")
	}

	// Returning role/groups/name/preferredUsername/email from UserInfo is optional in OIDC, but many clients rely on it. We mirror the same claims here as in the ID Token for compatibility.
	return s.ClientService.GetOidcUserInfo(at.UserID)
}

func (s *OidcServiceImpl) Discovery() (*DiscoveryConfig, error) {
	host, err := s.ConfigsService.GetBaseDomain()
	if err != nil {
		return nil, err
	}
	issuer := "https://quollix." + host

	return &DiscoveryConfig{
		Issuer:                            issuer,
		AuthorizationEndpoint:             issuer + tools.Paths.BackendOidcAuthorize,
		TokenEndpoint:                     issuer + tools.Paths.BackendOidcToken,
		JwksURI:                           issuer + tools.Paths.BackendOidcJwks,
		UserinfoEndpoint:                  issuer + tools.Paths.BackendOidcUserinfo,
		ResponseTypesSupported:            []string{"code"},
		SubjectTypesSupported:             []string{"public"},
		IDTokenSigningAlgValuesSupported:  []string{"RS256"},
		TokenEndpointAuthMethodsSupported: []string{"client_secret_basic", "client_secret_post"},
		CodeChallengeMethodsSupported:     []string{"S256"},
		// NOTE: offline_access is advertised here but currently not required to receive a refresh token.
		ScopesSupported: []string{
			"openid",
			"profile",
			"email",
			"groups",
			"offline_access",
		},
	}, nil
}

func (s *OidcServiceImpl) GetJwks() JWKS {
	return s.IdTokenIssuer.GetJwks()
}
