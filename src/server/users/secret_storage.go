package users

import (
	"sync"
	"time"

	u "github.com/quollix/common/utils"
)

var SecretDoesNotExistError = "secret does not exist"

type SecretAndCookieStorage interface {
	LoadCookieViaSecret(secret string) (string, error)
	GenerateSecretForCookie(cookieValue string) (string, error)
}

type SecretAndCookieStorageImpl struct {
	Secrets    sync.Map `wire:"-"`
	AuthHelper u.AuthHelper
}

func (s *SecretAndCookieStorageImpl) LoadCookieViaSecret(secret string) (string, error) {
	cookieValue, ok := s.Secrets.Load(secret)
	if !ok {
		return "", u.Logger.NewError(SecretDoesNotExistError)
	}
	cookieValueString := cookieValue.(string)
	s.Secrets.Delete(secret)
	return cookieValueString, nil
}

func (s *SecretAndCookieStorageImpl) GenerateSecretForCookie(cookieValue string) (string, error) {
	secret, err := s.AuthHelper.GenerateSecret()
	if err != nil {
		return "", err
	}
	s.Secrets.Store(secret, cookieValue)
	time.AfterFunc(3*time.Second, func() {
		// Secrets should be consumed almost immediately by users, so storage is only temporary.
		s.Secrets.Delete(secret)
	})
	return secret, nil
}
