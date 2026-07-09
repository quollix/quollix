//go:build integration

package repository

import (
	"server/tools"
	"server/users"
	"testing"
	"time"

	"github.com/quollix/common/assert"
)

func TestSessionReadings(t *testing.T) {
	user := initSessionRepoTest(t)
	defer cleanupSessionRepoTest()

	expectedSession := getSampleSession(user.Id)
	var err error
	expectedSession.SessionId, err = SessionRepo.CreateSession(expectedSession)
	assert.Nil(t, err)

	actualSession, err := SessionRepo.GetAuthenticatedSession(
		expectedSession.HashedCookieValue,
		expectedSession.Audience,
	)
	assert.Nil(t, err)
	assertSessionEquality(t, expectedSession, &actualSession.Session)
	assertUserEquality(t, user, &actualSession.User)
}

func TestSessionAudienceIsolation(t *testing.T) {
	user := initSessionRepoTest(t)
	defer cleanupSessionRepoTest()

	session := getSampleSession(user.Id)
	session.Audience = users.SessionAudience("samplemaintainer", "sampleapp")
	session.HashedCookieValue = "audience-isolated-cookie"
	_, err := SessionRepo.CreateSession(session)
	assert.Nil(t, err)

	actualSession, err := SessionRepo.GetAuthenticatedSession(
		session.HashedCookieValue,
		users.SessionAudience("othermaintainer", "sampleapp"),
	)
	assert.NotNil(t, err)
	assert.Nil(t, actualSession)

	actualSession, err = SessionRepo.GetAuthenticatedSession(
		"wrong-cookie",
		session.Audience,
	)
	assert.NotNil(t, err)
	assert.Nil(t, actualSession)
}

func TestSessionExpirationDateUpdate(t *testing.T) {
	user := initSessionRepoTest(t)
	defer cleanupSessionRepoTest()

	session := getSampleSession(user.Id)
	session.HashedCookieValue = "cookie-to-update"
	_, err := SessionRepo.CreateSession(session)
	assert.Nil(t, err)

	updatedExpirationDate := time.Date(2026, time.June, 11, 0, 0, 0, 0, time.UTC)
	err = SessionRepo.UpdateCookieExpirationDate(
		session.HashedCookieValue,
		session.Audience,
		updatedExpirationDate,
	)
	assert.Nil(t, err)

	actualSession, err := SessionRepo.GetAuthenticatedSession(
		session.HashedCookieValue,
		session.Audience,
	)
	assert.Nil(t, err)
	assert.Equal(t, updatedExpirationDate, actualSession.Session.CookieExpirationDate)
}

func TestSessionDeletion(t *testing.T) {
	user := initSessionRepoTest(t)
	defer cleanupSessionRepoTest()

	quollixSession := getSampleSession(user.Id)
	quollixSession.HashedCookieValue = "quollix-cookie"
	_, err := SessionRepo.CreateSession(quollixSession)
	assert.Nil(t, err)

	appSession := getSampleSession(user.Id)
	appSession.Audience = users.SessionAudience("samplemaintainer", "sampleapp")
	appSession.HashedCookieValue = "app-cookie"
	_, err = SessionRepo.CreateSession(appSession)
	assert.Nil(t, err)

	assert.Nil(t, SessionRepo.DeleteSessionsByUserId(user.Id))

	actualSession, err := SessionRepo.GetAuthenticatedSession(quollixSession.HashedCookieValue, quollixSession.Audience)
	assert.NotNil(t, err)
	assert.Nil(t, actualSession)

	actualSession, err = SessionRepo.GetAuthenticatedSession(appSession.HashedCookieValue, appSession.Audience)
	assert.NotNil(t, err)
	assert.Nil(t, actualSession)
}

func TestSessionUserDeletionCascade(t *testing.T) {
	user := initSessionRepoTest(t)
	defer cleanupSessionRepoTest()

	session := getSampleSession(user.Id)
	session.HashedCookieValue = "cascade-cookie"
	_, err := SessionRepo.CreateSession(session)
	assert.Nil(t, err)

	assert.Nil(t, UserRepo.DeleteUser(user.Id))

	actualSession, err := SessionRepo.GetAuthenticatedSession(session.HashedCookieValue, session.Audience)
	assert.NotNil(t, err)
	assert.Nil(t, actualSession)
}

func TestExpiredSessionDeletion(t *testing.T) {
	user := initSessionRepoTest(t)
	defer cleanupSessionRepoTest()

	expiredSession := getSampleSession(user.Id)
	expiredSession.HashedCookieValue = "expired-cookie"
	expiredSession.CookieExpirationDate = time.Now().UTC().Add(-time.Hour).Round(time.Microsecond)
	_, err := SessionRepo.CreateSession(expiredSession)
	assert.Nil(t, err)

	futureSession := getSampleSession(user.Id)
	futureSession.Audience = users.SessionAudience("samplemaintainer", "sampleapp")
	futureSession.HashedCookieValue = "future-cookie"
	futureSession.CookieExpirationDate = time.Now().UTC().Add(time.Hour).Round(time.Microsecond)
	futureSession.SessionId, err = SessionRepo.CreateSession(futureSession)
	assert.Nil(t, err)

	assert.Nil(t, SessionRepo.DeleteExpiredSessions())

	actualSession, err := SessionRepo.GetAuthenticatedSession(expiredSession.HashedCookieValue, expiredSession.Audience)
	assert.NotNil(t, err)
	assert.Nil(t, actualSession)

	actualSession, err = SessionRepo.GetAuthenticatedSession(futureSession.HashedCookieValue, futureSession.Audience)
	assert.Nil(t, err)
	assertSessionEquality(t, futureSession, &actualSession.Session)
}

func initSessionRepoTest(t *testing.T) *tools.User {
	InitDeps()

	user := GetSampleAdminUser()
	var err error
	user.Id, err = UserRepo.CreateUser(user)
	assert.Nil(t, err)

	return user
}

func cleanupSessionRepoTest() {
	SessionRepo.Wipe()
	UserRepo.Wipe()
}

func getSampleSession(userId int) *users.UserSession {
	return &users.UserSession{
		UserId:               userId,
		Audience:             users.QuollixSessionAudience(),
		HashedCookieValue:    "hashed-cookie-value",
		CookieExpirationDate: time.Date(2026, time.June, 10, 0, 0, 0, 0, time.UTC),
	}
}

func assertSessionEquality(t *testing.T, expected, actual *users.UserSession) {
	assert.Equal(t, expected.SessionId, actual.SessionId)
	assert.Equal(t, expected.UserId, actual.UserId)
	assert.Equal(t, expected.Audience, actual.Audience)
	assert.Equal(t, expected.HashedCookieValue, actual.HashedCookieValue)
	assert.Equal(t, expected.CookieExpirationDate.Round(time.Microsecond), actual.CookieExpirationDate)
}
