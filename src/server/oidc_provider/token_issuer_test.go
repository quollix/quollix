package oidc_provider

import (
	"testing"
	"time"

	"server/tools"

	"github.com/quollix/common/assert"
	"github.com/stretchr/testify/mock"
)

type tokenIssuerFakeClock struct {
	NowValue time.Time
}

func (c tokenIssuerFakeClock) Now() time.Time {
	return c.NowValue
}

func TestTokenIssuerImpl_IssueAccessToken_HappyPath_StoresTokenWithExpiryAndReturnsToken(t *testing.T) {
	cache := NewOidcCacheMock(t)
	authHelper := tools.NewAuthHelperMock(t)
	clock := tokenIssuerFakeClock{NowValue: time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)}

	issuer := &TokenIssuerImpl{
		Cache:      cache,
		AuthHelper: authHelper,
		Clock:      clock,
	}

	authHelper.EXPECT().GenerateSecret().Return("access-token-1", nil)

	var stored AccessToken
	cache.EXPECT().
		StoreAccessToken(mock.Anything).
		Run(func(accessToken AccessToken) {
			stored = accessToken
		})

	token, err := issuer.IssueAccessToken("7", "client-1")
	assert.Nil(t, err)
	assert.Equal(t, "access-token-1", token)

	assert.Equal(t, "access-token-1", stored.Token)
	assert.Equal(t, "7", stored.UserID)
	assert.Equal(t, "client-1", stored.ClientID)
	assert.Equal(t, clock.Now().Add(accessTokenTTL), stored.Expiry)
}

func TestTokenIssuerImpl_IssueForUser_HappyPath_StoresAccessAndRefreshAndReturnsBoth(t *testing.T) {
	cache := NewOidcCacheMock(t)
	authHelper := tools.NewAuthHelperMock(t)
	clock := tokenIssuerFakeClock{NowValue: time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)}

	issuer := &TokenIssuerImpl{
		Cache:      cache,
		AuthHelper: authHelper,
		Clock:      clock,
	}

	authHelper.EXPECT().GenerateSecret().Return("access-token-1", nil).Once()
	authHelper.EXPECT().GenerateSecret().Return("family-1", nil).Once()
	authHelper.EXPECT().GenerateSecret().Return("refresh-token-1", nil).Once()

	var storedAccessToken AccessToken
	cache.EXPECT().
		StoreAccessToken(mock.Anything).
		Run(func(accessToken AccessToken) {
			storedAccessToken = accessToken
		}).
		Once()

	var storedRefreshToken RefreshToken
	cache.EXPECT().
		StoreRefreshToken(mock.Anything).
		Run(func(refreshToken RefreshToken) {
			storedRefreshToken = refreshToken
		}).
		Once()

	accessToken, refreshToken, err := issuer.IssueForUser("7", "client-1")
	assert.Nil(t, err)
	assert.Equal(t, "access-token-1", accessToken)
	assert.Equal(t, "refresh-token-1", refreshToken)

	assert.Equal(t, "access-token-1", storedAccessToken.Token)
	assert.Equal(t, "7", storedAccessToken.UserID)
	assert.Equal(t, "client-1", storedAccessToken.ClientID)
	assert.Equal(t, clock.Now().Add(accessTokenTTL), storedAccessToken.Expiry)

	assert.Equal(t, "refresh-token-1", storedRefreshToken.Token)
	assert.Equal(t, "7", storedRefreshToken.UserID)
	assert.Equal(t, "client-1", storedRefreshToken.ClientID)
	assert.Equal(t, "family-1", storedRefreshToken.FamilyId)
	assert.Equal(t, clock.Now().Add(refreshTokenTTL), storedRefreshToken.Expiry)
}
