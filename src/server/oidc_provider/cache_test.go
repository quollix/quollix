package oidc_provider

import (
	"server/tools"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

var (
	sampleNow    = time.Date(2026, 2, 27, 10, 0, 0, 0, time.UTC)
	samplePast   = sampleNow.Add(-1 * time.Second)
	sampleFuture = sampleNow.Add(1 * time.Hour)

	sampleAuthCode = AuthCode{
		Code:   "code-123",
		Expiry: sampleFuture,
	}

	sampleAccessToken = AccessToken{
		Token:  "access-123",
		Expiry: sampleFuture,
	}

	sampleRefreshToken = RefreshToken{
		Token:    "refresh-123",
		FamilyId: "family-1",
		Expiry:   sampleFuture,
	}
)

type oidcCacheTestObjects struct {
	Cache         *OidcCacheImpl
	OsWrapperMock *tools.CommonOsWrapperMock
}

func setupOidcCacheTest(t *testing.T) oidcCacheTestObjects {
	osWrapperMock := tools.NewCommonOsWrapperMock(t)
	cache := NewOidcCache(osWrapperMock).(*OidcCacheImpl)
	return oidcCacheTestObjects{
		Cache:         cache,
		OsWrapperMock: osWrapperMock,
	}
}

func TestStoreAuthCode_ThenConsume_ReturnsAndDeletes(t *testing.T) {
	testObjects := setupOidcCacheTest(t)

	testObjects.Cache.StoreAuthCode(sampleAuthCode)

	consumedAuthCode, ok := testObjects.Cache.ConsumeAuthCode(sampleAuthCode.Code)
	assert.True(t, ok)
	assert.Equal(t, sampleAuthCode.Code, consumedAuthCode.Code)

	_, ok = testObjects.Cache.ConsumeAuthCode(sampleAuthCode.Code)
	assert.False(t, ok)
}

func TestConsumeAuthCode_WhenMissing_ReturnsFalse(t *testing.T) {
	testObjects := setupOidcCacheTest(t)

	_, ok := testObjects.Cache.ConsumeAuthCode("missing-code")
	assert.False(t, ok)
}

func TestStoreAccessToken_ThenGet_ReturnsToken(t *testing.T) {
	testObjects := setupOidcCacheTest(t)

	testObjects.Cache.StoreAccessToken(sampleAccessToken)

	foundAccessToken, ok := testObjects.Cache.GetAccessToken(sampleAccessToken.Token)
	assert.True(t, ok)
	assert.Equal(t, sampleAccessToken.Token, foundAccessToken.Token)
}

func TestGetAccessToken_WhenMissing_ReturnsFalse(t *testing.T) {
	testObjects := setupOidcCacheTest(t)

	_, ok := testObjects.Cache.GetAccessToken("missing-access")
	assert.False(t, ok)
}

func TestStoreRefreshToken_ThenConsume_ReturnsAndMovesToUsed_ThenReuseDetected(t *testing.T) {
	testObjects := setupOidcCacheTest(t)

	testObjects.OsWrapperMock.EXPECT().Now().Return(sampleNow)

	testObjects.Cache.StoreRefreshToken(sampleRefreshToken)

	foundRefreshToken, err := testObjects.Cache.ConsumeRefreshToken(sampleRefreshToken.Token)
	assert.Nil(t, err)
	assert.Equal(t, sampleRefreshToken.Token, foundRefreshToken.Token)

	_, err = testObjects.Cache.ConsumeRefreshToken(sampleRefreshToken.Token)
	assert.Equal(t, "refresh token reuse detected", u.ExtractError(err))
}

func TestConsumeRefreshToken_WhenUsedButExpired_ReturnsNotFound(t *testing.T) {
	testObjects := setupOidcCacheTest(t)

	testObjects.OsWrapperMock.EXPECT().Now().Return(sampleNow)

	testObjects.Cache.UsedRefreshTokens["refresh-123"] = RefreshToken{
		Token:    "refresh-123",
		FamilyId: "family-1",
		Expiry:   samplePast,
	}

	_, err := testObjects.Cache.ConsumeRefreshToken("refresh-123")
	assert.Equal(t, "refresh token not found", u.ExtractError(err))
}

func TestConsumeRefreshToken_WhenActive_MovesToUsedAndDeletesActive(t *testing.T) {
	testObjects := setupOidcCacheTest(t)

	rt := RefreshToken{
		Token:    "refresh-abc",
		FamilyId: "family-xyz",
		Expiry:   sampleFuture,
	}
	testObjects.Cache.StoreRefreshToken(rt)

	found, err := testObjects.Cache.ConsumeRefreshToken(rt.Token)
	assert.Nil(t, err)
	assert.Equal(t, rt.Token, found.Token)

	_, stillActive := testObjects.Cache.RefreshTokens[rt.Token]
	assert.False(t, stillActive)

	_, isUsed := testObjects.Cache.UsedRefreshTokens[rt.Token]
	assert.True(t, isUsed)

	_, isRevoked := testObjects.Cache.RevokedRefreshFamilies[rt.FamilyId]
	assert.False(t, isRevoked)
}

func TestConsumeRefreshToken_WhenReusedAndNotExpired_RevokesFamilyAndDeletesOtherActiveTokens(t *testing.T) {
	testObjects := setupOidcCacheTest(t)

	testObjects.OsWrapperMock.EXPECT().Now().Return(sampleNow).Twice()

	rt1 := RefreshToken{
		Token:    "refresh-1",
		FamilyId: "family-1",
		Expiry:   sampleFuture,
	}
	rt2 := RefreshToken{
		Token:    "refresh-2",
		FamilyId: "family-1",
		Expiry:   sampleFuture,
	}

	testObjects.Cache.StoreRefreshToken(rt1)
	testObjects.Cache.StoreRefreshToken(rt2)

	_, err := testObjects.Cache.ConsumeRefreshToken(rt1.Token)
	assert.Nil(t, err)

	_, err = testObjects.Cache.ConsumeRefreshToken(rt1.Token)
	assert.Equal(t, "refresh token reuse detected", u.ExtractError(err))

	_, isRevoked := testObjects.Cache.RevokedRefreshFamilies["family-1"]
	assert.True(t, isRevoked)

	_, stillActive := testObjects.Cache.RefreshTokens[rt2.Token]
	assert.False(t, stillActive)
}

func TestConsumeRefreshToken_WhenReusedButExpired_RemovesUsedAndReturnsNotFound(t *testing.T) {
	testObjects := setupOidcCacheTest(t)

	testObjects.OsWrapperMock.EXPECT().Now().Return(sampleNow)

	testObjects.Cache.UsedRefreshTokens["refresh-1"] = RefreshToken{
		Token:    "refresh-1",
		FamilyId: "family-1",
		Expiry:   samplePast,
	}

	_, err := testObjects.Cache.ConsumeRefreshToken("refresh-1")
	assert.Equal(t, "refresh token not found", u.ExtractError(err))

	_, stillUsed := testObjects.Cache.UsedRefreshTokens["refresh-1"]
	assert.False(t, stillUsed)

	_, isRevoked := testObjects.Cache.RevokedRefreshFamilies["family-1"]
	assert.False(t, isRevoked)
}

func TestStoreRefreshToken_WhenFamilyRevoked_DoesNotStore(t *testing.T) {
	testObjects := setupOidcCacheTest(t)

	testObjects.Cache.RevokedRefreshFamilies["family-1"] = sampleNow

	rt := RefreshToken{
		Token:    "refresh-1",
		FamilyId: "family-1",
		Expiry:   sampleFuture,
	}
	testObjects.Cache.StoreRefreshToken(rt)

	_, ok := testObjects.Cache.RefreshTokens[rt.Token]
	assert.False(t, ok)
}

func TestConsumeRefreshToken_WhenActiveButFamilyRevoked_ReturnsFamilyRevoked(t *testing.T) {
	testObjects := setupOidcCacheTest(t)

	rt := RefreshToken{
		Token:    "refresh-1",
		FamilyId: "family-1",
		Expiry:   sampleFuture,
	}
	testObjects.Cache.StoreRefreshToken(rt)
	testObjects.Cache.RevokedRefreshFamilies["family-1"] = sampleNow

	_, err := testObjects.Cache.ConsumeRefreshToken(rt.Token)
	assert.Equal(t, "refresh token family revoked", u.ExtractError(err))

	_, stillActive := testObjects.Cache.RefreshTokens[rt.Token]
	assert.False(t, stillActive)
}

func TestCleanup_RemovesExpiredAndOldRevokedFamilies(t *testing.T) {
	testObjects := setupOidcCacheTest(t)

	testObjects.OsWrapperMock.EXPECT().Now().Return(sampleNow)

	testObjects.Cache.AuthCodes["expired-code"] = AuthCode{Code: "expired-code", Expiry: samplePast}
	testObjects.Cache.AuthCodes["valid-code"] = AuthCode{Code: "valid-code", Expiry: sampleFuture}

	testObjects.Cache.AccessTokens["expired-at"] = AccessToken{Token: "expired-at", Expiry: samplePast}
	testObjects.Cache.AccessTokens["valid-at"] = AccessToken{Token: "valid-at", Expiry: sampleFuture}

	testObjects.Cache.RefreshTokens["expired-rt"] = RefreshToken{Token: "expired-rt", FamilyId: "family-exp", Expiry: samplePast}
	testObjects.Cache.RefreshTokens["revoked-rt"] = RefreshToken{Token: "revoked-rt", FamilyId: "family-rev", Expiry: sampleFuture}
	testObjects.Cache.RefreshTokens["valid-rt"] = RefreshToken{Token: "valid-rt", FamilyId: "family-ok", Expiry: sampleFuture}

	testObjects.Cache.UsedRefreshTokens["expired-used"] = RefreshToken{Token: "expired-used", FamilyId: "family-used", Expiry: samplePast}
	testObjects.Cache.UsedRefreshTokens["valid-used"] = RefreshToken{Token: "valid-used", FamilyId: "family-used", Expiry: sampleFuture}

	testObjects.Cache.RevokedRefreshFamilies["family-rev"] = sampleNow
	testObjects.Cache.RevokedRefreshFamilies["family-old"] = sampleNow.Add(-refreshTokenTTL).Add(-1 * time.Second)

	testObjects.Cache.Cleanup()

	_, ok := testObjects.Cache.AuthCodes["expired-code"]
	assert.False(t, ok)
	_, ok = testObjects.Cache.AuthCodes["valid-code"]
	assert.True(t, ok)

	_, ok = testObjects.Cache.AccessTokens["expired-at"]
	assert.False(t, ok)
	_, ok = testObjects.Cache.AccessTokens["valid-at"]
	assert.True(t, ok)

	_, ok = testObjects.Cache.RefreshTokens["expired-rt"]
	assert.False(t, ok)
	_, ok = testObjects.Cache.RefreshTokens["revoked-rt"]
	assert.False(t, ok)
	_, ok = testObjects.Cache.RefreshTokens["valid-rt"]
	assert.True(t, ok)

	_, ok = testObjects.Cache.UsedRefreshTokens["expired-used"]
	assert.False(t, ok)
	_, ok = testObjects.Cache.UsedRefreshTokens["valid-used"]
	assert.True(t, ok)

	_, ok = testObjects.Cache.RevokedRefreshFamilies["family-old"]
	assert.False(t, ok)
	_, ok = testObjects.Cache.RevokedRefreshFamilies["family-rev"]
	assert.True(t, ok)
}
