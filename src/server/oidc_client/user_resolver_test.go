package oidc_client

import (
	"strings"
	"testing"

	"server/users"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestOidcUserResolverImpl_ResolveUsername_UsesPreferredUsernameFirst(t *testing.T) {
	username, err := resolveUsernameForTest(t, OidcLoginClaims{
		PreferredUsername: "preferred",
		Name:              "name",
		Nickname:          "nickname",
	}, "preferred")

	assert.Nil(t, err)
	assert.Equal(t, "preferred", username)
}

func TestOidcUserResolverImpl_ResolveUsername_UsesNameFallback(t *testing.T) {
	username, err := resolveUsernameForTest(t, OidcLoginClaims{
		Name:     "name",
		Nickname: "nickname",
	}, "name")

	assert.Nil(t, err)
	assert.Equal(t, "name", username)
}

func TestOidcUserResolverImpl_ResolveUsername_UsesNicknameFallback(t *testing.T) {
	username, err := resolveUsernameForTest(t, OidcLoginClaims{
		Nickname: "nickname",
	}, "nickname")

	assert.Nil(t, err)
	assert.Equal(t, "nickname", username)
}

func TestOidcUserResolverImpl_ResolveUsername_UsesNextFallbackWhenClaimIsInvalid(t *testing.T) {
	username, err := resolveUsernameForTest(t, OidcLoginClaims{
		PreferredUsername: "xy",
		Name:              "valid-name",
	}, "valid-name")

	assert.Nil(t, err)
	assert.Equal(t, "valid-name", username)
}

func TestOidcUserResolverImpl_ResolveUsername_TrimsClaimValue(t *testing.T) {
	username, err := resolveUsernameForTest(t, OidcLoginClaims{
		PreferredUsername: "  preferred  ",
	}, "preferred")

	assert.Nil(t, err)
	assert.Equal(t, "preferred", username)
}

func TestOidcUserResolverImpl_ResolveUsername_NormalizesCommonClaimCharacters(t *testing.T) {
	username, err := resolveUsernameForTest(t, OidcLoginClaims{
		PreferredUsername: "  Tom Example  ",
	}, "tom-example")

	assert.Nil(t, err)
	assert.Equal(t, "tom-example", username)
}

func TestOidcUserResolverImpl_ResolveUsername_NormalizesPunctuationRuns(t *testing.T) {
	username, err := resolveUsernameForTest(t, OidcLoginClaims{
		Name: "Dr. Tom_O'Connor",
	}, "dr-tom_o-connor")

	assert.Nil(t, err)
	assert.Equal(t, "dr-tom_o-connor", username)
}

func TestOidcUserResolverImpl_ResolveUsername_ReturnsErrorWhenNoValidClaimExists(t *testing.T) {
	testObjects := newOidcUserResolverTestObjects(t)

	_, _, err := testObjects.Resolver.ResolveUser(OidcLoginClaims{
		PreferredUsername: "xy",
		Name:              "!!",
		Nickname:          "",
	})

	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), NoValidUsernameClaimError))
}

func TestOidcUserResolverImpl_ResolveUser_WhenEmailIsMissingUsesSyntheticEmailForResolvedUsername(t *testing.T) {
	testObjects := newOidcUserResolverTestObjects(t)
	testObjects.UserRepo.EXPECT().GetHighestGeneratedUsernameSuffix("external", users.UsernameMaxLength).Return(0, true, nil)

	username, email, err := testObjects.Resolver.ResolveUser(OidcLoginClaims{
		PreferredUsername: "external",
	})

	assert.Nil(t, err)
	assert.Equal(t, "external1", username)
	assert.Equal(t, "external1@example.invalid", email)
}

func TestOidcUserResolverImpl_ResolveUser_WhenEmailAlreadyExistsReturnsOidcLoginEmailAlreadyExistsError(t *testing.T) {
	testObjects := newOidcUserResolverTestObjects(t)
	testObjects.UserRepo.EXPECT().GetHighestGeneratedUsernameSuffix("external", users.UsernameMaxLength).Return(0, false, nil)
	testObjects.UserRepo.EXPECT().DoesEmailExist("external@example.com").Return(true, nil)

	_, _, err := testObjects.Resolver.ResolveUser(OidcLoginClaims{
		Email:             "external@example.com",
		PreferredUsername: "external",
	})

	assert.Equal(t, OidcLoginEmailAlreadyExistsError, u.ExtractError(err))
}

func TestOidcUserResolverImpl_ResolveUser_WhenUsernameCollidesKeepsSuffixWithinUsernameLengthLimit(t *testing.T) {
	testObjects := newOidcUserResolverTestObjects(t)
	testObjects.UserRepo.EXPECT().GetHighestGeneratedUsernameSuffix("abcdefghijklmnopqrst", users.UsernameMaxLength).Return(0, true, nil)

	username, email, err := testObjects.Resolver.ResolveUser(OidcLoginClaims{
		PreferredUsername: "abcdefghijklmnopqrst",
	})

	assert.Nil(t, err)
	assert.Equal(t, "abcdefghijklmnopqrs1", username)
	assert.Equal(t, "abcdefghijklmnopqrs1@example.invalid", email)
	assert.Equal(t, users.UsernameMaxLength, len(username))
}

func TestUsernameCandidate_WhenSuffixHasTwoDigitsTruncatesBase(t *testing.T) {
	username := usernameCandidate("abcdefghijklmnopqrst", 12)

	assert.Equal(t, "abcdefghijklmnopqr12", username)
	assert.Equal(t, users.UsernameMaxLength, len(username))
}

type oidcUserResolverTestObjects struct {
	Resolver *OidcUserResolverImpl
	UserRepo *users.UserRepositoryMock
}

func newOidcUserResolverTestObjects(t *testing.T) oidcUserResolverTestObjects {
	userRepo := users.NewUserRepositoryMock(t)
	return oidcUserResolverTestObjects{
		Resolver: &OidcUserResolverImpl{UserRepo: userRepo},
		UserRepo: userRepo,
	}
}

func resolveUsernameForTest(t *testing.T, claims OidcLoginClaims, expectedUsername string) (string, error) {
	testObjects := newOidcUserResolverTestObjects(t)
	testObjects.UserRepo.EXPECT().GetHighestGeneratedUsernameSuffix(expectedUsername, users.UsernameMaxLength).Return(0, false, nil)
	username, _, err := testObjects.Resolver.ResolveUser(claims)
	return username, err
}
