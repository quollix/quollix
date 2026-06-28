package oidc_provider

import (
	"strconv"

	"server/configs"
	"server/groups"
	"server/users"
)

type OidcClientService interface {
	CheckClientIdAndRedirectUri(clientId, redirectUri string) error
	AuthenticateClient(clientId, clientSecret string) error

	GetOidcUserInfo(userIdString string) (*UserinfoResponse, error)
	BuildIdTokenClaims(userIdString, audience, nonce string) (*IDTokenClaims, error)
}

type OidcClientServiceImpl struct {
	ClientResolver   OidcRelyingPartyResolver
	ConfigsService   configs.ConfigsService
	OidcEmailService configs.OidcEmailExposureService
	UserRepository   users.UserRepository
	GroupRepository  groups.GroupRepository
	Clock            Clock
}

func (s *OidcClientServiceImpl) CheckClientIdAndRedirectUri(clientId string, redirectUri string) error {
	return s.ClientResolver.CheckClientIdAndRedirectUri(clientId, redirectUri)
}

func (s *OidcClientServiceImpl) AuthenticateClient(clientId string, clientSecret string) error {
	return s.ClientResolver.AuthenticateClient(clientId, clientSecret)
}

// NOTE: Quollix follows an opinionated OIDC server model optimized for app integration simplicity. Quollix returns a fixed claim set and does not dynamically filter claims based on requested scopes or the optional OIDC `claims` parameter.

func (s *OidcClientServiceImpl) GetOidcUserInfo(userIdString string) (*UserinfoResponse, error) {
	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		return nil, err
	}

	user, err := s.UserRepository.GetUserById(userId)
	if err != nil {
		return nil, err
	}

	role := "user"
	if user.IsAdmin {
		role = "admin"
	}

	userGroups, err := s.GroupRepository.ListGroupsForUser(userId)
	if err != nil {
		return nil, err
	}

	groupNames := make([]string, 0, len(userGroups))
	for _, group := range userGroups {
		groupNames = append(groupNames, group.Name)
	}

	exposeRealEmail, err := s.OidcEmailService.ReadExposeRealEmailInOidcToken()
	if err != nil {
		return nil, err
	}
	email := users.ReservedEmailForUsername(user.Username)
	if exposeRealEmail {
		email = user.Email
	}

	return &UserinfoResponse{
		Sub:               userIdString,
		Role:              role,
		Groups:            groupNames,
		Name:              user.Username,
		PreferredUsername: user.Username,
		Email:             email,
	}, nil
}

func (s *OidcClientServiceImpl) BuildIdTokenClaims(userIdString string, audience string, nonce string) (*IDTokenClaims, error) {
	host, err := s.ConfigsService.GetBaseDomain()
	if err != nil {
		return nil, err
	}
	issuer := "https://quollix." + host

	issuedAt := s.Clock.Now().Unix()
	expiresAt := s.Clock.Now().Add(accessTokenTTL).Unix()

	userInfo, err := s.GetOidcUserInfo(userIdString)
	if err != nil {
		return nil, err
	}

	return &IDTokenClaims{
		Sub:    userIdString,
		Iss:    issuer,
		Aud:    audience,
		Nonce:  nonce,
		Role:   userInfo.Role,
		Groups: userInfo.Groups,
		Exp:    expiresAt,
		Iat:    issuedAt,

		Name:              userInfo.Name,
		PreferredUsername: userInfo.PreferredUsername,
		Email:             userInfo.Email,
	}, nil
}
