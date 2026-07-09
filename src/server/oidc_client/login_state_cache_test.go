package oidc_client

import (
	"server/tools"
	"testing"
	"time"

	"github.com/quollix/common/assert"
)

var oidcLoginStateCacheNow = time.Date(2026, time.June, 21, 10, 0, 0, 0, time.UTC)

func TestOidcLoginStateCacheImpl_ConsumeLoginStateReturnsStoredStateOnlyOnce(t *testing.T) {
	cache := NewOidcLoginStateCache(tools.NewCommonOsWrapperMock(t))
	cache.StoreLoginState(OidcLoginState{
		State:      "state",
		ProviderId: 12,
		ExpiresAt:  oidcLoginStateCacheNow.Add(time.Minute),
	})

	loginState, found := cache.ConsumeLoginState("state", oidcLoginStateCacheNow)
	_, foundAgain := cache.ConsumeLoginState("state", oidcLoginStateCacheNow)

	assert.True(t, found)
	assert.Equal(t, 12, loginState.ProviderId)
	assert.False(t, foundAgain)
}

func TestOidcLoginStateCacheImpl_ConsumeLoginStateReturnsFalseForExpiredState(t *testing.T) {
	cache := NewOidcLoginStateCache(tools.NewCommonOsWrapperMock(t))
	cache.StoreLoginState(OidcLoginState{
		State:      "state",
		ProviderId: 12,
		ExpiresAt:  oidcLoginStateCacheNow.Add(-time.Minute),
	})

	_, found := cache.ConsumeLoginState("state", oidcLoginStateCacheNow)

	assert.False(t, found)
}

func TestOidcLoginStateCacheImpl_CleanupExpiredLoginStatesRemovesOnlyExpiredStates(t *testing.T) {
	osWrapper := tools.NewCommonOsWrapperMock(t)
	cache := NewOidcLoginStateCache(osWrapper)
	cache.StoreLoginState(OidcLoginState{
		State:      "expired-state",
		ProviderId: 12,
		ExpiresAt:  oidcLoginStateCacheNow.Add(-time.Minute),
	})
	cache.StoreLoginState(OidcLoginState{
		State:      "valid-state",
		ProviderId: 13,
		ExpiresAt:  oidcLoginStateCacheNow.Add(time.Minute),
	})

	osWrapper.EXPECT().Now().Return(oidcLoginStateCacheNow)
	cache.CleanupExpiredLoginStates()

	_, expiredFound := cache.ConsumeLoginState("expired-state", oidcLoginStateCacheNow)
	validLoginState, validFound := cache.ConsumeLoginState("valid-state", oidcLoginStateCacheNow)
	assert.False(t, expiredFound)
	assert.True(t, validFound)
	assert.Equal(t, 13, validLoginState.ProviderId)
}
