package oidc_provider

import (
	"strings"
	"unicode"
	"unicode/utf8"

	u "github.com/quollix/common/utils"
)

const (
	oidcProtocolOpaqueMaxLength  = 1024
	oidcProtocolRedirectMaxBytes = 2048
)

type oidcInputField struct {
	name          string
	value         string
	required      bool
	maxLength     int
	missingErrMsg string
}

func validateAuthorizeInput(in AuthorizeInput) error {
	return validateOidcInput(
		requiredOidcField("response_type", in.ResponseType),
		requiredOidcField("client_id", in.ClientID),
		requiredOidcRedirectField("redirect_uri", in.RedirectURI),
		optionalOidcField("state", in.State),
		optionalOidcField("nonce", in.Nonce),
		optionalOidcField("code_challenge", in.CodeChallenge),
		optionalOidcField("code_challenge_method", in.CodeChallengeMethod),
	)
}

func validateAuthCodeGrantInput(in AuthCodeGrantInput) error {
	return validateOidcInput(
		requiredOidcFieldWithMissingError("code", in.Code, "grant missing code"),
		requiredOidcRedirectField("redirect_uri", in.RedirectURI),
		requiredOidcField("client_id", in.ClientID),
		optionalOidcField("client_secret", in.ClientSecret),
		optionalOidcField("code_verifier", in.CodeVerifier),
	)
}

func validateRefreshTokenGrantInput(in RefreshTokenGrantInput) error {
	return validateOidcInput(
		requiredOidcFieldWithMissingError("refresh_token", in.RefreshToken, "grant missing refresh_token"),
		requiredOidcField("client_id", in.ClientID),
		optionalOidcField("client_secret", in.ClientSecret),
	)
}

func requiredOidcField(name string, value string) oidcInputField {
	return requiredOidcFieldWithMissingError(name, value, "")
}

func requiredOidcFieldWithMissingError(name string, value string, missingErrMsg string) oidcInputField {
	return oidcInputField{
		name:          name,
		value:         value,
		required:      true,
		maxLength:     oidcProtocolOpaqueMaxLength,
		missingErrMsg: missingErrMsg,
	}
}

func requiredOidcRedirectField(name string, value string) oidcInputField {
	return oidcInputField{
		name:      name,
		value:     value,
		required:  true,
		maxLength: oidcProtocolRedirectMaxBytes,
	}
}

func optionalOidcField(name string, value string) oidcInputField {
	return oidcInputField{
		name:      name,
		value:     value,
		maxLength: oidcProtocolOpaqueMaxLength,
	}
}

func validateOidcInput(fields ...oidcInputField) error {
	for _, field := range fields {
		if err := validateOidcField(field); err != nil {
			return err
		}
	}
	return nil
}

func validateOidcField(field oidcInputField) error {
	if field.required && strings.TrimSpace(field.value) == "" {
		if field.missingErrMsg != "" {
			return u.Logger.NewError(field.missingErrMsg)
		}
		return u.Logger.NewError("missing OIDC protocol input", "field", field.name)
	}
	if len(field.value) > field.maxLength {
		return u.Logger.NewError("OIDC protocol input is too long", "field", field.name, "max_length", field.maxLength)
	}
	if !utf8.ValidString(field.value) {
		return u.Logger.NewError("OIDC protocol input must be valid UTF-8", "field", field.name)
	}
	for _, r := range field.value {
		if unicode.IsControl(r) {
			return u.Logger.NewError("OIDC protocol input contains control characters", "field", field.name)
		}
	}
	return nil
}
