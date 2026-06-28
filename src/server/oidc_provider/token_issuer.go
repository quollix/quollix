package oidc_provider

import (
	u "github.com/quollix/common/utils"
)

type TokenIssuer interface {
	IssueForUser(userIdString string, clientId string) (accessToken string, refreshToken string, err error)
	IssueAccessToken(userIdString string, clientId string) (accessToken string, err error)
}

type TokenIssuerImpl struct {
	Cache      OidcCache
	AuthHelper u.AuthHelper
	Clock      Clock
}

func (i *TokenIssuerImpl) IssueAccessToken(userIdString string, clientId string) (string, error) {
	token, err := i.AuthHelper.GenerateSecret()
	if err != nil {
		return "", err
	}

	i.Cache.StoreAccessToken(AccessToken{
		Token:    token,
		UserID:   userIdString,
		ClientID: clientId,
		Expiry:   i.Clock.Now().Add(accessTokenTTL),
	})

	return token, nil
}

func (i *TokenIssuerImpl) IssueForUser(userIdString string, clientId string) (string, string, error) {
	accessToken, err := i.IssueAccessToken(userIdString, clientId)
	if err != nil {
		return "", "", err
	}

	familyId, err := i.AuthHelper.GenerateSecret()
	if err != nil {
		return "", "", err
	}

	refreshToken, err := i.AuthHelper.GenerateSecret()
	if err != nil {
		return "", "", err
	}

	i.Cache.StoreRefreshToken(RefreshToken{
		Token:    refreshToken,
		UserID:   userIdString,
		ClientID: clientId,
		Expiry:   i.Clock.Now().Add(refreshTokenTTL),
		FamilyId: familyId,
	})

	return accessToken, refreshToken, nil
}
