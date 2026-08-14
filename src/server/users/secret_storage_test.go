package users

import (
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestSecretAndCookieStorage_LoadCookieViaSecretForApp(t *testing.T) {
	storage := newSecretAndCookieStorageForTest()
	secret, err := storage.GenerateSecretForCookie("cookie-value", "sample-app")
	assert.Nil(t, err)

	cookieValue, err := storage.LoadCookieViaSecret(secret, "sample-app")

	assert.Nil(t, err)
	assert.Equal(t, "cookie-value", cookieValue)
}

func TestSecretAndCookieStorage_LoadCookieViaSecretRejectsUnknownSecret(t *testing.T) {
	storage := newSecretAndCookieStorageForTest()

	_, err := storage.LoadCookieViaSecret("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "sample-app")

	assert.NotNil(t, err)
	assert.Equal(t, SecretDoesNotExistError, u.ExtractError(err))
}

func TestSecretAndCookieStorage_LoadCookieViaSecretRejectsDifferentApp(t *testing.T) {
	storage := newSecretAndCookieStorageForTest()
	secret, err := storage.GenerateSecretForCookie("cookie-value", "sample-app")
	assert.Nil(t, err)

	_, err = storage.LoadCookieViaSecret(secret, "other-app")

	assert.NotNil(t, err)
	assert.Equal(t, SecretDoesNotExistError, u.ExtractError(err))
}

func TestSecretAndCookieStorage_LoadCookieViaSecretConsumesSecretAfterSuccess(t *testing.T) {
	storage := newSecretAndCookieStorageForTest()
	secret, err := storage.GenerateSecretForCookie("cookie-value", "sample-app")
	assert.Nil(t, err)
	_, err = storage.LoadCookieViaSecret(secret, "sample-app")
	assert.Nil(t, err)

	_, err = storage.LoadCookieViaSecret(secret, "sample-app")

	assert.NotNil(t, err)
	assert.Equal(t, SecretDoesNotExistError, u.ExtractError(err))
}

func TestSecretAndCookieStorage_LoadCookieViaSecretConsumesSecretAfterAppMismatch(t *testing.T) {
	storage := newSecretAndCookieStorageForTest()
	secret, err := storage.GenerateSecretForCookie("cookie-value", "sample-app")
	assert.Nil(t, err)
	_, err = storage.LoadCookieViaSecret(secret, "other-app")
	assert.NotNil(t, err)

	_, err = storage.LoadCookieViaSecret(secret, "sample-app")

	assert.NotNil(t, err)
	assert.Equal(t, SecretDoesNotExistError, u.ExtractError(err))
}

func newSecretAndCookieStorageForTest() *SecretAndCookieStorageImpl {
	return &SecretAndCookieStorageImpl{
		AuthHelper: &u.AuthHelperImpl{},
	}
}
