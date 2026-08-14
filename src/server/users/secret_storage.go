package users

import (
	"sync"
	"time"

	u "github.com/quollix/common/utils"
)

var SecretDoesNotExistError = "secret does not exist"

type SecretAndCookieStorage interface {
	LoadCookieViaSecret(secret string, appName string) (string, error)
	GenerateSecretForCookie(cookieValue string, appName string) (string, error)
}

type SecretAndCookieStorageImpl struct {
	Secrets    sync.Map `wire:"-"`
	AuthHelper u.AuthHelper
}

type secretEntry struct {
	CookieValue string
	AppName     string
}

func (s *SecretAndCookieStorageImpl) LoadCookieViaSecret(secret string, appName string) (string, error) {
	rawEntry, ok := s.Secrets.LoadAndDelete(secret)
	if !ok {
		return "", u.Logger.NewError(SecretDoesNotExistError)
	}
	entry := rawEntry.(secretEntry)
	if entry.AppName != appName {
		return "", u.Logger.NewError(SecretDoesNotExistError)
	}
	return entry.CookieValue, nil
}

func (s *SecretAndCookieStorageImpl) GenerateSecretForCookie(cookieValue string, appName string) (string, error) {
	secret, err := s.AuthHelper.GenerateSecret()
	if err != nil {
		return "", err
	}
	s.Secrets.Store(secret, secretEntry{
		CookieValue: cookieValue,
		AppName:     appName,
	})
	time.AfterFunc(3*time.Second, func() {
		// Secrets should be consumed almost immediately by users, so storage is only temporary.
		s.Secrets.Delete(secret)
	})
	return secret, nil
}
