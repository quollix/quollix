package oidc_client

import (
	"net/http"
	"server/tools"
	"server/users"
	"strings"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

const MissingSubjectClaimError = "OIDC provider did not return a subject claim"

type LoginService interface {
	LoginWithClaims(providerId int, claims OidcLoginClaims) (*http.Cookie, error)
}

type LoginServiceImpl struct {
	ProviderRepo       OidcAuthProviderRepository
	UserAuthMethodRepo UserAuthMethodRepository
	UserRepo           users.UserRepository
	UserResolver       OidcUserResolver
	SessionService     users.SessionService
	OsWrapper          u.OsWrapper
}

type OidcLoginClaims struct {
	Subject           string `json:"sub"`
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
	Nickname          string `json:"nickname"`
}

func (s *LoginServiceImpl) LoginWithClaims(providerId int, claims OidcLoginClaims) (*http.Cookie, error) {
	claims.Subject = strings.TrimSpace(claims.Subject)
	if err := validateLoginClaims(claims); err != nil {
		return nil, err
	}
	if _, err := s.ProviderRepo.GetProviderById(providerId); err != nil {
		return nil, err
	}

	authenticatedAt := s.OsWrapper.Now()
	method, found, err := s.UserAuthMethodRepo.GetUserAuthMethodByProviderAndSubject(providerId, claims.Subject)
	if err != nil {
		return nil, err
	}
	if found {
		user, err := s.UserRepo.GetUserById(method.UserId)
		if err != nil {
			return nil, err
		}
		if !user.IsEnabled {
			return nil, u.Logger.NewError(users.UserDisabledError)
		}
		if err = s.UserAuthMethodRepo.UpdateLastOidcAuthenticatedAt(method.Id, authenticatedAt); err != nil {
			return nil, err
		}
		return s.SessionService.GenerateAndSaveCookie(method.UserId, users.QuollixSessionAudience())
	}

	username, email, err := s.UserResolver.ResolveUser(claims)
	if err != nil {
		return nil, err
	}

	user := users.NewUser(username, email, "", "", tools.DefaultTime, false, "", tools.DefaultTime, authenticatedAt)
	user.Id, err = s.UserRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}
	_, err = s.UserAuthMethodRepo.CreateUserAuthMethod(&UserAuthMethod{
		UserId:                  user.Id,
		OidcAuthProviderId:      providerId,
		ExternalSubject:         claims.Subject,
		LastOidcAuthenticatedAt: authenticatedAt,
	})
	if err != nil {
		return nil, err
	}
	return s.SessionService.GenerateAndSaveCookie(user.Id, users.QuollixSessionAudience())
}

func validateLoginClaims(claims OidcLoginClaims) error {
	if strings.TrimSpace(claims.Subject) == "" {
		return u.Logger.NewError(MissingSubjectClaimError)
	}
	claimValidations := []struct {
		fieldName     string
		validationTag string
		value         string
	}{
		{"sub", validation.FieldOidcSubject, claims.Subject},
		{"email", validation.FieldOidcClaim, claims.Email},
		{"preferred_username", validation.FieldOidcClaim, claims.PreferredUsername},
		{"name", validation.FieldOidcClaim, claims.Name},
		{"nickname", validation.FieldOidcClaim, claims.Nickname},
	}
	for _, claim := range claimValidations {
		if err := validation.Validate(claim.fieldName, claim.validationTag, claim.value); err != nil {
			return err
		}
	}
	return nil
}
