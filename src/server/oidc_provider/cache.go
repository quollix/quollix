package oidc_provider

import (
	"sync"
	"time"

	u "github.com/quollix/common/utils"
)

type OidcCache interface {
	StoreAuthCode(ac AuthCode)
	ConsumeAuthCode(code string) (AuthCode, bool)

	StoreAccessToken(at AccessToken)
	GetAccessToken(token string) (AccessToken, bool)

	StoreRefreshToken(rt RefreshToken)
	ConsumeRefreshToken(token string) (*RefreshToken, error)

	Cleanup()
}

func NewOidcCache(osWrapper u.OsWrapper) OidcCache {
	return &OidcCacheImpl{
		OsWrapper:              osWrapper,
		AuthCodes:              make(map[string]AuthCode),
		AccessTokens:           make(map[string]AccessToken),
		RefreshTokens:          make(map[string]RefreshToken),
		UsedRefreshTokens:      make(map[string]RefreshToken),
		RevokedRefreshFamilies: make(map[string]time.Time),
	}
}

type OidcCacheImpl struct {
	mu sync.Mutex

	OsWrapper u.OsWrapper

	AuthCodes    map[string]AuthCode
	AccessTokens map[string]AccessToken

	RefreshTokens          map[string]RefreshToken
	UsedRefreshTokens      map[string]RefreshToken
	RevokedRefreshFamilies map[string]time.Time
}

func (c *OidcCacheImpl) StoreAuthCode(ac AuthCode) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AuthCodes[ac.Code] = ac
}

func (c *OidcCacheImpl) ConsumeAuthCode(code string) (AuthCode, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ac, ok := c.AuthCodes[code]
	if ok {
		delete(c.AuthCodes, code)
	}
	return ac, ok
}

func (c *OidcCacheImpl) StoreAccessToken(at AccessToken) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.AccessTokens[at.Token] = at
}

func (c *OidcCacheImpl) GetAccessToken(token string) (AccessToken, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	at, ok := c.AccessTokens[token]
	return at, ok
}

func (c *OidcCacheImpl) StoreRefreshToken(rt RefreshToken) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, isRevoked := c.RevokedRefreshFamilies[rt.FamilyId]; isRevoked {
		return
	}
	c.RefreshTokens[rt.Token] = rt
}

func (c *OidcCacheImpl) revokeRefreshTokenFamily(familyId string) {
	c.RevokedRefreshFamilies[familyId] = c.OsWrapper.Now()

	for tokenValue, refreshToken := range c.RefreshTokens {
		if refreshToken.FamilyId == familyId {
			delete(c.RefreshTokens, tokenValue)
		}
	}
}

func (c *OidcCacheImpl) ConsumeRefreshToken(token string) (*RefreshToken, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if activeRefreshToken, ok := c.RefreshTokens[token]; ok {
		if _, isRevoked := c.RevokedRefreshFamilies[activeRefreshToken.FamilyId]; isRevoked {
			delete(c.RefreshTokens, token)
			return nil, u.Logger.NewError("refresh token family revoked")
		}

		delete(c.RefreshTokens, token)
		c.UsedRefreshTokens[token] = activeRefreshToken
		return &activeRefreshToken, nil
	}

	if usedRefreshToken, ok := c.UsedRefreshTokens[token]; ok {
		now := c.OsWrapper.Now()
		if now.After(usedRefreshToken.Expiry) {
			delete(c.UsedRefreshTokens, token)
			return nil, u.Logger.NewError("refresh token not found")
		}
		c.revokeRefreshTokenFamily(usedRefreshToken.FamilyId)
		return nil, u.Logger.NewError("refresh token reuse detected")
	}

	return nil, u.Logger.NewError("refresh token not found")
}

func (c *OidcCacheImpl) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.OsWrapper.Now()

	for codeValue, authCode := range c.AuthCodes {
		if now.After(authCode.Expiry) {
			delete(c.AuthCodes, codeValue)
		}
	}

	for tokenValue, accessToken := range c.AccessTokens {
		if now.After(accessToken.Expiry) {
			delete(c.AccessTokens, tokenValue)
		}
	}

	for tokenValue, refreshToken := range c.RefreshTokens {
		if now.After(refreshToken.Expiry) {
			delete(c.RefreshTokens, tokenValue)
			continue
		}
		if _, isRevoked := c.RevokedRefreshFamilies[refreshToken.FamilyId]; isRevoked {
			delete(c.RefreshTokens, tokenValue)
		}
	}

	for tokenValue, refreshToken := range c.UsedRefreshTokens {
		if now.After(refreshToken.Expiry) {
			delete(c.UsedRefreshTokens, tokenValue)
		}
	}

	for familyId, revokedAt := range c.RevokedRefreshFamilies {
		if now.After(revokedAt.Add(refreshTokenTTL)) {
			delete(c.RevokedRefreshFamilies, familyId)
		}
	}
}
