package oidc_provider

import (
	"net/url"
	"server/apps_basic"
	"server/configs"
	"server/groups"
	"server/users"
	"strconv"

	u "github.com/quollix/common/utils"
)

type OidcClientService interface {
	CheckClientIdAndRedirectUri(clientId, redirectUri string) error
	AuthenticateClient(clientId, clientSecret string) error

	GetOidcUserInfo(userIdString string) (*UserinfoResponse, error)
	BuildIdTokenClaims(userIdString, audience, nonce string) (*IDTokenClaims, error)
}

type OidcClientServiceImpl struct {
	AppRepository   apps_basic.AppRepository
	ConfigsRepo     configs.ConfigsRepository
	UserRepository  users.UserRepository
	GroupRepository groups.GroupRepository
	Clock           Clock
}

func (s *OidcClientServiceImpl) CheckClientIdAndRedirectUri(clientId string, redirectUri string) error {
	app, err := s.AppRepository.GetAppByClientId(clientId)
	if err != nil {
		return err
	}
	host, err := s.ConfigsRepo.GetConfig(configs.ConfigKeys.ServerHost)
	if err != nil {
		return err
	}
	redirectUrl, err := url.Parse(redirectUri)
	if err != nil {
		return u.Logger.NewError("invalid redirect_uri format")
	}

	expectedHostname := app.AppName + "." + host
	if redirectUrl.Scheme != "https" {
		return u.Logger.NewError("wrong redirect_uri scheme")
	}
	if redirectUrl.Hostname() != expectedHostname {
		return u.Logger.NewError("wrong redirect_uri host")
	}
	redirectPort := redirectUrl.Port()
	if redirectPort != "" && redirectPort != "443" {
		return u.Logger.NewError("wrong redirect_uri port")
	}

	return nil
}

func (s *OidcClientServiceImpl) AuthenticateClient(clientId string, clientSecret string) error {
	app, err := s.AppRepository.GetAppByClientId(clientId)
	if err != nil {
		return err
	}

	if app.ClientSecret == "" {
		return u.Logger.NewError("confidential client missing stored client secret")
	}
	if clientSecret == "" {
		return u.Logger.NewError("confidential client missing client secret")
	}
	if app.ClientSecret != clientSecret {
		return u.Logger.NewError("wrong client secret")
	}
	return nil
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

	return &UserinfoResponse{
		Sub:               userIdString,
		Role:              role,
		Groups:            groupNames,
		Name:              user.Username,
		PreferredUsername: user.Username,
		Email:             user.Email,
	}, nil
}

func (s *OidcClientServiceImpl) BuildIdTokenClaims(userIdString string, audience string, nonce string) (*IDTokenClaims, error) {
	host, err := s.ConfigsRepo.GetConfig(configs.ConfigKeys.ServerHost)
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
