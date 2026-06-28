package users

import (
	"strings"

	u "github.com/quollix/common/utils"
)

const (
	ReservedEmailDomain              = "example.invalid"
	ReservedEmailMustMatchUserError  = "example.invalid email addresses must match the username"
	ReservedEmailRenameConflictError = "renaming this user would also change their example.invalid email, but the target email already exists"
	reservedEmailDomainWithSeparator = "@" + ReservedEmailDomain
)

func ValidateReservedEmail(username, email string) error {
	if !IsReservedEmail(email) {
		return nil
	}
	if email == ReservedEmailForUsername(username) {
		return nil
	}
	return u.Logger.NewError(ReservedEmailMustMatchUserError)
}

func IsReservedEmail(email string) bool {
	return strings.HasSuffix(email, reservedEmailDomainWithSeparator)
}

func ReservedEmailForUsername(username string) string {
	return username + reservedEmailDomainWithSeparator
}
