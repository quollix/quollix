package oidc_provider

import (
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"reflect"
	"server/configs"
	"server/tools"
	"server/users"
	"strconv"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
	"github.com/stretchr/testify/mock"
)

var sampleTime = time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)

type oidcServiceTestObjects struct {
	Service       *OidcServiceImpl
	ConfigsRepo   *configs.ConfigsRepositoryMock
	UserRepo      *users.UserRepositoryMock
	Cache         *OidcCacheMock
	AuthHelper    *tools.AuthHelperMock
	ClientService *OidcClientServiceMock
	TokenIssuer   *TokenIssuerMock
	IdTokenIssuer *IdTokenIssuerMock
	Clock         *ClockMock
}

func newOidcServiceTestObjects(t *testing.T) oidcServiceTestObjects {
	configsRepo := configs.NewConfigsRepositoryMock(t)
	userRepo := users.NewUserRepositoryMock(t)
	cache := NewOidcCacheMock(t)
	authHelper := tools.NewAuthHelperMock(t)

	clientService := NewOidcClientServiceMock(t)
	tokenIssuer := NewTokenIssuerMock(t)
	idTokenIssuer := NewIdTokenIssuerMock(t)
	clock := NewClockMock(t)

	service := &OidcServiceImpl{
		Cache:          cache,
		ConfigsRepo:    configsRepo,
		UserRepository: userRepo,
		AuthHelper:     authHelper,
		Clock:          clock,
		ClientService:  clientService,
		TokenIssuer:    tokenIssuer,
		IdTokenIssuer:  idTokenIssuer,
	}

	return oidcServiceTestObjects{
		Service:       service,
		ConfigsRepo:   configsRepo,
		UserRepo:      userRepo,
		Cache:         cache,
		AuthHelper:    authHelper,
		ClientService: clientService,
		TokenIssuer:   tokenIssuer,
		IdTokenIssuer: idTokenIssuer,
		Clock:         clock,
	}
}

func TestInitializeOidcService_DelegatesToIdTokenIssuer(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	testObjects.IdTokenIssuer.EXPECT().PrepareJwkToken().Return(nil)

	err := testObjects.Service.InitializeOidcService()
	assert.Nil(t, err)
}

func authorizeInput() AuthorizeInput {
	return AuthorizeInput{
		ResponseType:        "code",
		ClientID:            "client-1",
		RedirectURI:         "https://client.localhost/callback",
		State:               "state-1",
		Nonce:               "nonce-1",
		CodeChallenge:       "challenge-1",
		CodeChallengeMethod: "S256",
		UserID:              7,
	}
}

func assertAuthorizeRedirect(t *testing.T, redirectString string, expectedCode string, expectedState string) {
	redirectUrl, err := url.Parse(redirectString)
	assert.Nil(t, err)
	assert.Equal(t, "https", redirectUrl.Scheme)
	assert.Equal(t, "client.localhost", redirectUrl.Host)
	assert.Equal(t, "/callback", redirectUrl.Path)

	query := redirectUrl.Query()
	assert.Equal(t, expectedCode, query.Get("code"))
	assert.Equal(t, expectedState, query.Get("state"))
}

func assertStoredAuthCodeMatchesInput(t *testing.T, storedAuthCode AuthCode, in AuthorizeInput, expectedCode string, expectedExpiry time.Time) {
	assert.Equal(t, expectedCode, storedAuthCode.Code)
	assert.Equal(t, in.ClientID, storedAuthCode.ClientID)
	assert.Equal(t, strconv.Itoa(in.UserID), storedAuthCode.UserID)
	normalizedRedirectUri, err := normalizeRedirectUri(in.RedirectURI)
	assert.Nil(t, err)
	assert.Equal(t, normalizedRedirectUri, storedAuthCode.RedirectURI)
	assert.Equal(t, in.Nonce, storedAuthCode.Nonce)
	assert.Equal(t, in.CodeChallenge, storedAuthCode.CodeChallenge)
	assert.Equal(t, in.CodeChallengeMethod, storedAuthCode.CodeChallengeMethod)
	assert.Equal(t, expectedExpiry, storedAuthCode.Expiry)
}

func TestOidcServiceImpl_Authorize_HappyPath(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := authorizeInput()

	testObjects.ClientService.EXPECT().CheckClientIdAndRedirectUri(in.ClientID, in.RedirectURI).Return(nil)
	testObjects.UserRepo.EXPECT().GetUserById(in.UserID).Return(&tools.User{Id: in.UserID}, nil)
	testObjects.AuthHelper.EXPECT().GenerateSecret().Return("code-123", nil)
	testObjects.Clock.EXPECT().Now().Return(sampleTime)

	var storedAuthCode AuthCode
	testObjects.Cache.EXPECT().
		StoreAuthCode(mock.Anything).
		Run(func(authCode AuthCode) {
			storedAuthCode = authCode
		})

	redirectString, err := testObjects.Service.Authorize(in)
	assert.Nil(t, err)

	assertAuthorizeRedirect(t, redirectString, "code-123", "state-1")
	assertStoredAuthCodeMatchesInput(t, storedAuthCode, in, "code-123", sampleTime.Add(authCodeTTL))
}

func TestOidcServiceImpl_Authorize_WhenResponseTypeNotCode_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := authorizeInput()
	in.ResponseType = "token"

	_, err := testObjects.Service.Authorize(in)
	assert.Equal(t, "unsupported response_type, expected 'code'", u.ExtractError(err))
}

func TestOidcServiceImpl_Authorize_WhenCodeChallengeMethodNotS256_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := authorizeInput()
	in.CodeChallengeMethod = "plain"

	_, err := testObjects.Service.Authorize(in)
	assert.Equal(t, "unsupported code_challenge_method, expected 'S256' or empty", u.ExtractError(err))
}

func pkceChallengeFromVerifier(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func authCodeGrantInput() AuthCodeGrantInput {
	return AuthCodeGrantInput{
		Code:         "auth-code-1",
		RedirectURI:  "https://sampleapp.localhost/callback",
		CodeVerifier: "verifier-1",
		ClientID:     "client-1",
		ClientSecret: "secret-1",
	}
}

func storedAuthCodeFromInput(in AuthCodeGrantInput) AuthCode {
	return AuthCode{
		Code:                in.Code,
		ClientID:            in.ClientID,
		UserID:              "7",
		RedirectURI:         in.RedirectURI,
		Nonce:               "nonce-1",
		CodeChallenge:       pkceChallengeFromVerifier(in.CodeVerifier),
		CodeChallengeMethod: "S256",
		Expiry:              sampleTime,
	}
}

func TestTokenWithAuthCode_HappyPath_ReturnsTokens(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := authCodeGrantInput()
	authCode := storedAuthCodeFromInput(in)
	expectedClaims := &IDTokenClaims{
		Sub:               authCode.UserID,
		Iss:               "https://quollix.localhost",
		Aud:               authCode.ClientID,
		Nonce:             authCode.Nonce,
		Role:              "admin",
		Groups:            []string{"admins"},
		Exp:               sampleTime.Add(accessTokenTTL).Unix(),
		Name:              "Test User",
		PreferredUsername: "testuser",
		Email:             "test@example.com",
	}

	testObjects.ClientService.EXPECT().CheckClientIdAndRedirectUri(in.ClientID, in.RedirectURI).Return(nil)
	testObjects.ClientService.EXPECT().AuthenticateClient(in.ClientID, in.ClientSecret).Return(nil)
	testObjects.Cache.EXPECT().ConsumeAuthCode(in.Code).Return(authCode, true)
	testObjects.Clock.EXPECT().Now().Return(sampleTime)

	testObjects.ClientService.EXPECT().
		BuildIdTokenClaims(authCode.UserID, authCode.ClientID, authCode.Nonce).
		Return(expectedClaims, nil)

	testObjects.IdTokenIssuer.EXPECT().
		Sign(mock.MatchedBy(func(claims *IDTokenClaims) bool {
			return reflect.DeepEqual(expectedClaims, claims)
		})).
		Return("id-token-1", nil)
	testObjects.TokenIssuer.EXPECT().IssueForUser(authCode.UserID, authCode.ClientID).Return("access-token-1", "refresh-token-1", nil)

	res, err := testObjects.Service.TokenWithAuthCode(in)
	assert.Nil(t, err)

	assert.Equal(t, "access-token-1", res.AccessToken)
	assert.Equal(t, "refresh-token-1", res.RefreshToken)
	assert.Equal(t, "id-token-1", res.IDToken)
	assert.Equal(t, "Bearer", res.TokenType)
	assert.Equal(t, accessTokenExpiry, res.ExpiresIn)
}

func TestTokenWithAuthCode_WhenMissingCode_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)

	_, err := testObjects.Service.TokenWithAuthCode(AuthCodeGrantInput{
		Code:         "",
		CodeVerifier: "verifier-1",
	})
	assert.Equal(t, "grant missing code", u.ExtractError(err))
}

func TestTokenWithAuthCode_WhenPkceUsedAndMissingCodeVerifier_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := authCodeGrantInput()
	in.CodeVerifier = ""

	authCode := storedAuthCodeFromInput(in)
	authCode.CodeChallengeMethod = "S256"
	authCode.CodeChallenge = pkceChallengeFromVerifier("verifier-expected")

	testObjects.ClientService.EXPECT().CheckClientIdAndRedirectUri(in.ClientID, in.RedirectURI).Return(nil)
	testObjects.ClientService.EXPECT().AuthenticateClient(in.ClientID, in.ClientSecret).Return(nil)
	testObjects.Cache.EXPECT().ConsumeAuthCode(in.Code).Return(authCode, true)
	testObjects.Clock.EXPECT().Now().Return(sampleTime)

	_, err := testObjects.Service.TokenWithAuthCode(in)
	assert.Equal(t, "grant missing code_verifier", u.ExtractError(err))
}

func TestTokenWithAuthCode_WhenGrantNotFound_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := authCodeGrantInput()

	testObjects.ClientService.EXPECT().CheckClientIdAndRedirectUri(in.ClientID, in.RedirectURI).Return(nil)
	testObjects.ClientService.EXPECT().AuthenticateClient(in.ClientID, in.ClientSecret).Return(nil)
	testObjects.Cache.EXPECT().ConsumeAuthCode(in.Code).Return(AuthCode{}, false)

	_, err := testObjects.Service.TokenWithAuthCode(in)
	assert.Equal(t, "grant not found", u.ExtractError(err))
}

func TestTokenWithAuthCode_WhenGrantExpired_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := authCodeGrantInput()

	authCode := storedAuthCodeFromInput(in)
	authCode.Expiry = sampleTime.Add(-1 * time.Minute)

	testObjects.ClientService.EXPECT().CheckClientIdAndRedirectUri(in.ClientID, in.RedirectURI).Return(nil)
	testObjects.ClientService.EXPECT().AuthenticateClient(in.ClientID, in.ClientSecret).Return(nil)
	testObjects.Cache.EXPECT().ConsumeAuthCode(in.Code).Return(authCode, true)
	testObjects.Clock.EXPECT().Now().Return(sampleTime)

	_, err := testObjects.Service.TokenWithAuthCode(in)
	assert.Equal(t, "grant expired", u.ExtractError(err))
}

func TestTokenWithAuthCode_WhenGrantClientIdMismatch_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := authCodeGrantInput()

	authCode := storedAuthCodeFromInput(in)
	authCode.ClientID = "other-client"

	testObjects.ClientService.EXPECT().CheckClientIdAndRedirectUri(in.ClientID, in.RedirectURI).Return(nil)
	testObjects.ClientService.EXPECT().AuthenticateClient(in.ClientID, in.ClientSecret).Return(nil)
	testObjects.Cache.EXPECT().ConsumeAuthCode(in.Code).Return(authCode, true)
	testObjects.Clock.EXPECT().Now().Return(sampleTime)

	_, err := testObjects.Service.TokenWithAuthCode(in)
	assert.Equal(t, "grant client_id mismatch", u.ExtractError(err))
}

func TestTokenWithAuthCode_WhenGrantRedirectUriMismatch_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := authCodeGrantInput()

	authCode := storedAuthCodeFromInput(in)
	authCode.RedirectURI = "https://other.localhost/callback"

	testObjects.ClientService.EXPECT().CheckClientIdAndRedirectUri(in.ClientID, in.RedirectURI).Return(nil)
	testObjects.ClientService.EXPECT().AuthenticateClient(in.ClientID, in.ClientSecret).Return(nil)
	testObjects.Cache.EXPECT().ConsumeAuthCode(in.Code).Return(authCode, true)
	testObjects.Clock.EXPECT().Now().Return(sampleTime)

	_, err := testObjects.Service.TokenWithAuthCode(in)
	assert.Equal(t, "grant redirect_uri mismatch", u.ExtractError(err))
}

func TestTokenWithAuthCode_WhenVerifierMismatch_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := authCodeGrantInput()
	in.CodeVerifier = "verifier-wrong"

	authCode := storedAuthCodeFromInput(in)
	authCode.CodeChallenge = pkceChallengeFromVerifier("verifier-correct")

	testObjects.ClientService.EXPECT().CheckClientIdAndRedirectUri(in.ClientID, in.RedirectURI).Return(nil)
	testObjects.ClientService.EXPECT().AuthenticateClient(in.ClientID, in.ClientSecret).Return(nil)
	testObjects.Cache.EXPECT().ConsumeAuthCode(in.Code).Return(authCode, true)
	testObjects.Clock.EXPECT().Now().Return(sampleTime)

	_, err := testObjects.Service.TokenWithAuthCode(in)
	assert.Equal(t, "invalid code_verifier", u.ExtractError(err))
}

func TestTokenWithAuthCode_WhenStoredCodeChallengeMethodNotS256_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := authCodeGrantInput()

	authCode := storedAuthCodeFromInput(in)
	authCode.CodeChallengeMethod = "plain"

	testObjects.ClientService.EXPECT().CheckClientIdAndRedirectUri(in.ClientID, in.RedirectURI).Return(nil)
	testObjects.ClientService.EXPECT().AuthenticateClient(in.ClientID, in.ClientSecret).Return(nil)
	testObjects.Cache.EXPECT().ConsumeAuthCode(in.Code).Return(authCode, true)
	testObjects.Clock.EXPECT().Now().Return(sampleTime)

	_, err := testObjects.Service.TokenWithAuthCode(in)
	assert.Equal(t, "unsupported code_challenge_method in stored auth code, expected 'S256'", u.ExtractError(err))
}

func TestTokenWithAuthCode_WhenClientSecretMissing_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := authCodeGrantInput()
	in.ClientSecret = ""

	testObjects.ClientService.EXPECT().CheckClientIdAndRedirectUri(in.ClientID, in.RedirectURI).Return(nil)
	testObjects.ClientService.EXPECT().AuthenticateClient(in.ClientID, in.ClientSecret).
		Return(u.Logger.NewError("confidential client missing client secret"))

	_, err := testObjects.Service.TokenWithAuthCode(in)
	assert.Equal(t, "confidential client missing client secret", u.ExtractError(err))
}

func refreshGrantInput() RefreshTokenGrantInput {
	return RefreshTokenGrantInput{
		RefreshToken: "refresh-token-1",
		ClientID:     "client-1",
		ClientSecret: "secret-1",
	}
}

func storedRefreshTokenFromInput(testObjects oidcServiceTestObjects, in RefreshTokenGrantInput) RefreshToken {
	return RefreshToken{
		Token:    in.RefreshToken,
		UserID:   "7",
		ClientID: in.ClientID,
		Expiry:   sampleTime.Add(2 * time.Minute),
		FamilyId: "family-1",
	}
}

func executeRefreshGrantSuccess(t *testing.T, stored RefreshToken) (*TokenResponse, RefreshToken, error) {
	testObjects := newOidcServiceTestObjects(t)
	in := refreshGrantInput()
	expectedClaims := &IDTokenClaims{
		Sub:               stored.UserID,
		Name:              "Test User",
		PreferredUsername: "testuser",
		Email:             "test@example.com",
	}
	testObjects.ClientService.EXPECT().AuthenticateClient(in.ClientID, in.ClientSecret).Return(nil)
	testObjects.Cache.EXPECT().ConsumeRefreshToken(in.RefreshToken).Return(&stored, nil)
	testObjects.Clock.EXPECT().Now().Return(sampleTime)

	testObjects.ClientService.EXPECT().
		BuildIdTokenClaims(stored.UserID, in.ClientID, "").
		Return(expectedClaims, nil)

	testObjects.IdTokenIssuer.EXPECT().
		Sign(mock.MatchedBy(func(claims *IDTokenClaims) bool {
			return reflect.DeepEqual(expectedClaims, claims)
		})).
		Return("id-token-1", nil)

	testObjects.TokenIssuer.EXPECT().
		IssueAccessToken(stored.UserID, in.ClientID).
		Return("access-token-new", nil)

	testObjects.AuthHelper.EXPECT().
		GenerateSecret().
		Return("refresh-token-new", nil)

	var storedNewRefreshToken RefreshToken
	testObjects.Cache.EXPECT().
		StoreRefreshToken(mock.Anything).
		Run(func(refreshToken RefreshToken) {
			storedNewRefreshToken = refreshToken
		})

	res, err := testObjects.Service.TokenWithRefresh(in)
	return res, storedNewRefreshToken, err
}

func TestTokenWithRefresh_HappyPath_ReturnsNewAccessTokenAndRotatedRefreshToken(t *testing.T) {
	in := refreshGrantInput()
	stored := storedRefreshTokenFromInput(oidcServiceTestObjects{}, in)
	res, storedNewRefreshToken, err := executeRefreshGrantSuccess(t, stored)
	assert.Nil(t, err)

	assert.Equal(t, "access-token-new", res.AccessToken)
	assert.Equal(t, "refresh-token-new", res.RefreshToken)
	assert.Equal(t, "id-token-1", res.IDToken)
	assert.Equal(t, "Bearer", res.TokenType)
	assert.Equal(t, accessTokenExpiry, res.ExpiresIn)

	assert.Equal(t, "refresh-token-new", storedNewRefreshToken.Token)
	assert.Equal(t, stored.UserID, storedNewRefreshToken.UserID)
	assert.Equal(t, stored.ClientID, storedNewRefreshToken.ClientID)
	assert.Equal(t, stored.FamilyId, storedNewRefreshToken.FamilyId)

	candidateExpiry := sampleTime.Add(refreshTokenTTL)
	expectedExpiry := stored.Expiry
	if candidateExpiry.Before(expectedExpiry) {
		expectedExpiry = candidateExpiry
	}
	assert.Equal(t, expectedExpiry, storedNewRefreshToken.Expiry)
}

func TestTokenWithRefresh_WhenStoredRefreshTokenExpiresAfterTtl_ClampsNewRefreshTokenExpiry(t *testing.T) {
	in := refreshGrantInput()
	stored := storedRefreshTokenFromInput(oidcServiceTestObjects{}, in)
	stored.Expiry = sampleTime.Add(refreshTokenTTL + time.Hour)

	_, storedNewRefreshToken, err := executeRefreshGrantSuccess(t, stored)
	assert.Nil(t, err)
	assert.Equal(t, sampleTime.Add(refreshTokenTTL), storedNewRefreshToken.Expiry)
}

func TestTokenWithRefresh_WhenMissingRefreshToken_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)

	_, err := testObjects.Service.TokenWithRefresh(RefreshTokenGrantInput{
		RefreshToken: "",
		ClientID:     "client-1",
		ClientSecret: "secret-1",
	})
	assert.Equal(t, "grant missing refresh_token", u.ExtractError(err))
}

func TestTokenWithRefresh_WhenRefreshTokenNotFound_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := refreshGrantInput()

	testObjects.ClientService.EXPECT().AuthenticateClient(in.ClientID, in.ClientSecret).Return(nil)
	testObjects.Cache.EXPECT().ConsumeRefreshToken(in.RefreshToken).Return((*RefreshToken)(nil), u.Logger.NewError("refresh token not found"))

	_, err := testObjects.Service.TokenWithRefresh(in)
	assert.Equal(t, "refresh token not found", u.ExtractError(err))
}

func TestTokenWithRefresh_WhenClientIdMismatch_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := refreshGrantInput()

	stored := storedRefreshTokenFromInput(testObjects, in)
	stored.ClientID = "other-client"

	testObjects.ClientService.EXPECT().AuthenticateClient(in.ClientID, in.ClientSecret).Return(nil)
	testObjects.Cache.EXPECT().ConsumeRefreshToken(in.RefreshToken).Return(&stored, nil)

	_, err := testObjects.Service.TokenWithRefresh(in)
	assert.Equal(t, "refresh token client_id mismatch", u.ExtractError(err))
}

func TestTokenWithRefresh_WhenRefreshTokenExpired_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := refreshGrantInput()

	stored := storedRefreshTokenFromInput(testObjects, in)
	stored.Expiry = sampleTime.Add(-1 * time.Minute)

	testObjects.ClientService.EXPECT().AuthenticateClient(in.ClientID, in.ClientSecret).Return(nil)
	testObjects.Cache.EXPECT().ConsumeRefreshToken(in.RefreshToken).Return(&stored, nil)
	testObjects.Clock.EXPECT().Now().Return(sampleTime)

	_, err := testObjects.Service.TokenWithRefresh(in)
	assert.Equal(t, "refresh token expired", u.ExtractError(err))
}

func TestUserinfo_WhenAccessTokenMissing_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)

	_, err := testObjects.Service.Userinfo("")
	assert.Equal(t, "access token missing", u.ExtractError(err))
}

func TestUserinfo_WhenAccessTokenNotFound_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)

	testObjects.Cache.EXPECT().GetAccessToken("access-token-1").Return(AccessToken{}, false)

	_, err := testObjects.Service.Userinfo("access-token-1")
	assert.Equal(t, "access token not found", u.ExtractError(err))
}

func TestUserinfo_WhenAccessTokenExpired_ReturnsError(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)

	testObjects.Clock.EXPECT().Now().Return(sampleTime)
	testObjects.Cache.EXPECT().
		GetAccessToken("access-token-1").
		Return(AccessToken{
			Token:    "access-token-1",
			UserID:   "7",
			ClientID: "client-1",
			Expiry:   sampleTime.Add(-1 * time.Minute),
		}, true)

	_, err := testObjects.Service.Userinfo("access-token-1")
	assert.Equal(t, "access token expired", u.ExtractError(err))
}

func TestUserinfo_HappyPath_ReturnsUserinfoIncludingProfileFields(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)

	testObjects.Clock.EXPECT().Now().Return(sampleTime)
	testObjects.Cache.EXPECT().
		GetAccessToken("access-token-1").
		Return(AccessToken{
			Token:    "access-token-1",
			UserID:   "7",
			ClientID: "client-1",
			Expiry:   sampleTime.Add(2 * time.Minute),
		}, true)

	testObjects.ClientService.EXPECT().
		GetOidcUserInfo("7").
		Return(&UserinfoResponse{
			Sub:               "7",
			Role:              "admin",
			Groups:            []string{"admins", "devs"},
			Name:              "Test User",
			PreferredUsername: "testuser",
			Email:             "test@example.com",
		}, nil)

	res, err := testObjects.Service.Userinfo("access-token-1")
	assert.Nil(t, err)

	assert.Equal(t, "7", res.Sub)
	assert.Equal(t, "admin", res.Role)
	assert.Equal(t, []string{"admins", "devs"}, res.Groups)
	assert.Equal(t, "Test User", res.Name)
	assert.Equal(t, "testuser", res.PreferredUsername)
	assert.Equal(t, "test@example.com", res.Email)
}

func TestDiscovery_ReturnsCorrectDiscoveryConfig(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	testObjects.ConfigsRepo.EXPECT().GetConfig(configs.ConfigKeys.ServerHost).Return("example.com", nil)

	res, err := testObjects.Service.Discovery()
	assert.Nil(t, err)

	expectedIssuer := "https://quollix.example.com"
	assert.Equal(t, expectedIssuer, res.Issuer)
	assert.Equal(t, expectedIssuer+tools.Paths.BackendOidcAuthorize, res.AuthorizationEndpoint)
	assert.Equal(t, expectedIssuer+tools.Paths.BackendOidcToken, res.TokenEndpoint)
	assert.Equal(t, expectedIssuer+tools.Paths.BackendOidcJwks, res.JwksURI)
	assert.Equal(t, expectedIssuer+tools.Paths.BackendOidcUserinfo, res.UserinfoEndpoint)
	assert.Equal(t, []string{"client_secret_basic", "client_secret_post"}, res.TokenEndpointAuthMethodsSupported)
	assert.Equal(t, []string{"S256"}, res.CodeChallengeMethodsSupported)
}

func TestGetJwks_DelegatesToIdTokenIssuer(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)

	expected := JWKS{Keys: []JWK{{Kty: "RSA"}}}
	testObjects.IdTokenIssuer.EXPECT().GetJwks().Return(expected)

	actual := testObjects.Service.GetJwks()
	assert.Equal(t, expected, actual)
}

func TestOidcServiceImpl_Authorize_WhenCodeChallengeMethodEmpty_AllowsAuthorize(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := authorizeInput()
	in.CodeChallenge = ""
	in.CodeChallengeMethod = ""

	testObjects.ClientService.EXPECT().CheckClientIdAndRedirectUri(in.ClientID, in.RedirectURI).Return(nil)
	testObjects.UserRepo.EXPECT().GetUserById(in.UserID).Return(&tools.User{Id: in.UserID}, nil)
	testObjects.AuthHelper.EXPECT().GenerateSecret().Return("code-123", nil)
	testObjects.Cache.EXPECT().StoreAuthCode(mock.Anything)
	testObjects.Clock.EXPECT().Now().Return(sampleTime)

	redirectString, err := testObjects.Service.Authorize(in)
	assert.Nil(t, err)
	assertAuthorizeRedirect(t, redirectString, "code-123", "state-1")
}

func TestTokenWithAuthCode_WhenNoPkce_DoesNotRequireCodeVerifier(t *testing.T) {
	testObjects := newOidcServiceTestObjects(t)
	in := authCodeGrantInput()
	in.CodeVerifier = ""

	authCode := storedAuthCodeFromInput(in)
	authCode.CodeChallenge = ""
	authCode.CodeChallengeMethod = ""
	expectedClaims := &IDTokenClaims{Sub: authCode.UserID}

	testObjects.ClientService.EXPECT().CheckClientIdAndRedirectUri(in.ClientID, in.RedirectURI).Return(nil)
	testObjects.ClientService.EXPECT().AuthenticateClient(in.ClientID, in.ClientSecret).Return(nil)
	testObjects.Cache.EXPECT().ConsumeAuthCode(in.Code).Return(authCode, true)
	testObjects.Clock.EXPECT().Now().Return(sampleTime)

	testObjects.ClientService.EXPECT().
		BuildIdTokenClaims(authCode.UserID, authCode.ClientID, authCode.Nonce).
		Return(expectedClaims, nil)

	testObjects.IdTokenIssuer.EXPECT().
		Sign(mock.MatchedBy(func(claims *IDTokenClaims) bool {
			return reflect.DeepEqual(expectedClaims, claims)
		})).
		Return("id-token-1", nil)

	testObjects.TokenIssuer.EXPECT().
		IssueForUser(authCode.UserID, authCode.ClientID).
		Return("access-token-1", "refresh-token-1", nil)

	res, err := testObjects.Service.TokenWithAuthCode(in)
	assert.Nil(t, err)
	assert.Equal(t, "access-token-1", res.AccessToken)
}
