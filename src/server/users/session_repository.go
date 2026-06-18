package users

import (
	"server/tools"
	"time"

	u "github.com/quollix/common/utils"
)

const (
	sessionInsert = `
		INSERT INTO user_sessions (
			user_id,
			audience,
			hashed_cookie_value,
			cookie_expiration_date
		)
		VALUES ($1, $2, $3, $4)
		RETURNING session_id
	`

	authenticatedSessionSelect = `
		SELECT
			users.user_id,
			users.username,
			users.email,
			users.hashed_password,
			users.is_admin,
			users.set_password_token,
			users.set_password_token_expiration_date,
			users.creation_date,
			user_sessions.session_id,
			user_sessions.user_id,
			user_sessions.audience,
			user_sessions.hashed_cookie_value,
			user_sessions.cookie_expiration_date
		FROM users
		JOIN user_sessions ON users.user_id = user_sessions.user_id
		WHERE user_sessions.hashed_cookie_value = $1 AND user_sessions.audience = $2
	`
)

type SessionRepository interface {
	CreateSession(session *UserSession) (int, error)
	GetAuthenticatedSession(hashedCookie, audience string) (*AuthenticatedSession, error)
	UpdateCookieExpirationDate(hashedCookie, audience string, cookieExpirationDate time.Time) error
	DeleteSessionsByUserId(userId int) error
	DeleteExpiredSessions() error
}

type SessionRepositoryImpl struct {
	DbProvider tools.DatabaseConnector
}

func (r *SessionRepositoryImpl) CreateSession(session *UserSession) (int, error) {
	var id int
	err := r.DbProvider.GetDB().QueryRow(
		sessionInsert,
		session.UserId,
		session.Audience,
		session.HashedCookieValue,
		session.CookieExpirationDate,
	).Scan(&id)
	if err != nil {
		return 0, u.Logger.NewError(err.Error())
	}
	return id, nil
}

func (r *SessionRepositoryImpl) GetAuthenticatedSession(hashedCookie, audience string) (*AuthenticatedSession, error) {
	row := r.DbProvider.GetDB().QueryRow(authenticatedSessionSelect, hashedCookie, audience)
	authenticatedSession, err := scanAuthenticatedSession(row)
	if err != nil {
		return nil, u.Logger.NewError(CookieNotFoundError)
	}
	return authenticatedSession, nil
}

func (r *SessionRepositoryImpl) UpdateCookieExpirationDate(hashedCookie, audience string, cookieExpirationDate time.Time) error {
	_, err := r.DbProvider.GetDB().Exec(
		`UPDATE user_sessions
		 SET cookie_expiration_date = $3
		 WHERE hashed_cookie_value = $1 AND audience = $2`,
		hashedCookie,
		audience,
		cookieExpirationDate,
	)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *SessionRepositoryImpl) DeleteSessionsByUserId(userId int) error {
	_, err := r.DbProvider.GetDB().Exec("DELETE FROM user_sessions WHERE user_id = $1", userId)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *SessionRepositoryImpl) DeleteExpiredSessions() error {
	_, err := r.DbProvider.GetDB().Exec("DELETE FROM user_sessions WHERE cookie_expiration_date < NOW()")
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

// only used during testing
func (r *SessionRepositoryImpl) Wipe() {
	_, err := r.DbProvider.GetDB().Exec("DELETE FROM user_sessions")
	if err != nil {
		u.Logger.Error(err)
	}
}

type sessionRowScanner interface {
	Scan(dest ...any) error
}

func scanAuthenticatedSession(row sessionRowScanner) (*AuthenticatedSession, error) {
	var authenticatedSession AuthenticatedSession
	if err := row.Scan(
		&authenticatedSession.User.Id,
		&authenticatedSession.User.Username,
		&authenticatedSession.User.Email,
		&authenticatedSession.User.HashedPassword,
		&authenticatedSession.User.IsAdmin,
		&authenticatedSession.User.SetPasswordToken,
		&authenticatedSession.User.SetPasswordTokenExpirationDate,
		&authenticatedSession.User.CreationDate,
		&authenticatedSession.Session.SessionId,
		&authenticatedSession.Session.UserId,
		&authenticatedSession.Session.Audience,
		&authenticatedSession.Session.HashedCookieValue,
		&authenticatedSession.Session.CookieExpirationDate,
	); err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	authenticatedSession.User.SetPasswordTokenExpirationDate = authenticatedSession.User.SetPasswordTokenExpirationDate.UTC()
	authenticatedSession.User.CreationDate = authenticatedSession.User.CreationDate.UTC()
	authenticatedSession.Session.CookieExpirationDate = authenticatedSession.Session.CookieExpirationDate.UTC()
	return &authenticatedSession, nil
}
