package oidc_provider

import (
	"strings"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestValidateAuthorizeInput_WhenStateContainsControlCharacterReturnsError(t *testing.T) {
	in := authorizeInputForValidationTest()
	in.State = "state\n"

	err := validateAuthorizeInput(in)

	assert.Equal(t, "OIDC protocol input contains control characters", u.ExtractError(err))
}

func TestValidateAuthorizeInput_WhenRedirectUriIsTooLongReturnsError(t *testing.T) {
	in := authorizeInputForValidationTest()
	in.RedirectURI = "https://client.example.com/" + strings.Repeat("a", oidcProtocolRedirectMaxBytes)

	err := validateAuthorizeInput(in)

	assert.Equal(t, "OIDC protocol input is too long", u.ExtractError(err))
}

func TestValidateAuthorizeInput_WhenOptionalFieldsAreEmptyReturnsNil(t *testing.T) {
	in := authorizeInputForValidationTest()
	in.State = ""
	in.Nonce = ""
	in.CodeChallenge = ""
	in.CodeChallengeMethod = ""

	err := validateAuthorizeInput(in)

	assert.Nil(t, err)
}

func TestValidateAuthCodeGrantInput_WhenCodeContainsControlCharacterReturnsError(t *testing.T) {
	in := authCodeGrantInputForValidationTest()
	in.Code = "code\t"

	err := validateAuthCodeGrantInput(in)

	assert.Equal(t, "OIDC protocol input contains control characters", u.ExtractError(err))
}

func TestValidateAuthCodeGrantInput_WhenCodeIsMissingKeepsExistingErrorMessage(t *testing.T) {
	in := authCodeGrantInputForValidationTest()
	in.Code = ""

	err := validateAuthCodeGrantInput(in)

	assert.Equal(t, "grant missing code", u.ExtractError(err))
}

func TestValidateRefreshTokenGrantInput_WhenRefreshTokenIsMissingKeepsExistingErrorMessage(t *testing.T) {
	in := refreshTokenGrantInputForValidationTest()
	in.RefreshToken = ""

	err := validateRefreshTokenGrantInput(in)

	assert.Equal(t, "grant missing refresh_token", u.ExtractError(err))
}

func authorizeInputForValidationTest() AuthorizeInput {
	return AuthorizeInput{
		ResponseType:        "code",
		ClientID:            "client-id",
		RedirectURI:         "https://client.example.com/callback",
		State:               "state",
		Nonce:               "nonce",
		CodeChallenge:       "code-challenge",
		CodeChallengeMethod: "S256",
		UserID:              1,
	}
}

func authCodeGrantInputForValidationTest() AuthCodeGrantInput {
	return AuthCodeGrantInput{
		Code:         "code",
		RedirectURI:  "https://client.example.com/callback",
		CodeVerifier: "code-verifier",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}
}

func refreshTokenGrantInputForValidationTest() RefreshTokenGrantInput {
	return RefreshTokenGrantInput{
		RefreshToken: "refresh-token",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}
}
