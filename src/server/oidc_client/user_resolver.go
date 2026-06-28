package oidc_client

import (
	"strconv"
	"strings"

	"server/users"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

const (
	NoValidUsernameClaimError        = "OIDC provider did not return a valid username claim"
	OidcLoginEmailAlreadyExistsError = "OIDC sign-in failed because the provider email is already used by another user"
)

type OidcUserResolver interface {
	ResolveUser(claims OidcLoginClaims) (string, string, error)
}

type OidcUserResolverImpl struct {
	UserRepo users.UserRepository
}

func (r *OidcUserResolverImpl) ResolveUser(claims OidcLoginClaims) (string, string, error) {
	username, err := r.resolveUsername(claims)
	if err != nil {
		return "", "", err
	}
	username, err = r.findAvailableUsername(username)
	if err != nil {
		return "", "", err
	}
	email, err := r.resolveEmail(username, claims.Email)
	if err != nil {
		return "", "", err
	}
	return username, email, nil
}

func (r *OidcUserResolverImpl) resolveUsername(claims OidcLoginClaims) (string, error) {
	usernameClaims := []string{
		claims.PreferredUsername,
		claims.Name,
		claims.Nickname,
	}
	for _, username := range usernameClaims {
		username = normalizeUsernameClaim(username)
		if validation.Validate("username", validation.FieldUsername, username) == nil {
			return username, nil
		}
	}
	return "", u.Logger.NewError(NoValidUsernameClaimError)
}

func (r *OidcUserResolverImpl) findAvailableUsername(username string) (string, error) {
	highestSuffix, exists, err := r.UserRepo.GetHighestGeneratedUsernameSuffix(username, users.UsernameMaxLength)
	if err != nil {
		return "", err
	}
	if !exists {
		return username, nil
	}
	return usernameCandidate(username, highestSuffix+1), nil
}

func usernameCandidate(username string, suffix int) string {
	suffixText := strconv.Itoa(suffix)
	baseLength := min(len(username), users.UsernameMaxLength-len(suffixText))
	return username[:baseLength] + suffixText
}

func (r *OidcUserResolverImpl) resolveEmail(username, claimEmail string) (string, error) {
	email := strings.TrimSpace(claimEmail)
	if email == "" {
		return users.ReservedEmailForUsername(username), nil
	}
	if err := validation.Validate("email", validation.FieldEmail, email); err != nil {
		return "", err
	}
	exists, err := r.UserRepo.DoesEmailExist(email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", u.Logger.NewError(OidcLoginEmailAlreadyExistsError)
	}
	return email, nil
}

func normalizeUsernameClaim(username string) string {
	var normalized strings.Builder
	previousSeparator := false
	for _, char := range strings.ToLower(strings.TrimSpace(username)) {
		if isUsernameChar(char) {
			normalized.WriteRune(char)
			previousSeparator = false
			continue
		}
		if !previousSeparator {
			normalized.WriteRune('-')
			previousSeparator = true
		}
	}
	return strings.Trim(normalized.String(), "-_")
}

func isUsernameChar(char rune) bool {
	return char >= 'a' && char <= 'z' ||
		char >= '0' && char <= '9' ||
		char == '_' ||
		char == '-'
}
