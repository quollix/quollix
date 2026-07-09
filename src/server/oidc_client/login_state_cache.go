package oidc_client

import (
	"sync"
	"time"

	u "github.com/quollix/common/utils"
)

type OidcLoginStateCache interface {
	StoreLoginState(state OidcLoginState)
	ConsumeLoginState(state string, now time.Time) (OidcLoginState, bool)
	CleanupExpiredLoginStates()
}

type OidcLoginState struct {
	State      string
	ProviderId int
	ExpiresAt  time.Time
}

func NewOidcLoginStateCache(osWrapper u.OsWrapper) OidcLoginStateCache {
	return &OidcLoginStateCacheImpl{
		OsWrapper:   osWrapper,
		LoginStates: make(map[string]OidcLoginState),
	}
}

type OidcLoginStateCacheImpl struct {
	mu sync.Mutex

	OsWrapper   u.OsWrapper
	LoginStates map[string]OidcLoginState
}

func (c *OidcLoginStateCacheImpl) StoreLoginState(state OidcLoginState) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.LoginStates[state.State] = state
}

func (c *OidcLoginStateCacheImpl) ConsumeLoginState(state string, now time.Time) (OidcLoginState, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	loginState, found := c.LoginStates[state]
	if !found {
		return OidcLoginState{}, false
	}
	delete(c.LoginStates, state)
	if now.After(loginState.ExpiresAt) {
		return OidcLoginState{}, false
	}
	return loginState, true
}

func (c *OidcLoginStateCacheImpl) CleanupExpiredLoginStates() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.OsWrapper.Now()
	for state, loginState := range c.LoginStates {
		if now.After(loginState.ExpiresAt) {
			delete(c.LoginStates, state)
		}
	}
}
