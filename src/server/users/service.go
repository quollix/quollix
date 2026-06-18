package users

import (
	"net/http"
	"server/tools"
	"time"

	u "github.com/quollix/common/utils"
	"golang.org/x/crypto/bcrypt"
)

const (
	IncorrectCurrentPasswordError = "current password is incorrect"
	TokenExpiredError             = "This password reset link has expired. Please ask your administrator for a new one."
	InvitationValidityDays        = 3

	// This user ID can never exist in the database, which allows for simple handling of anonymous users in the design.
	AnonymousUserId = -1
)

type UserService interface {
	IsPasswordCorrect(creds Credentials) bool
	Logout(userId int) error
	UserResetsOwnPassword(userId int, oldPassword, newPassword string) error
	InviteUser(username string, email string) (*InvitationDetails, error)
	AcceptNewPasswordViaToken(password, setPasswordToken string) error
	ResetPasswordOfUser(userIdToResetPassword int, currentlyLoggedInUserId int) (string, error)
	DeleteUser(idOfUserToBeDeleted, currentlyLoggedInUserId int) error

	ChangeUsername(userId int, newUsername string) error
	ChangeEmail(userId int, newEmail string) error
	GetUserIdAndRoleFromQuollixRequest(r *http.Request) (int, tools.UserAccessLevel, error)
	GetUserIdAndRoleFromRequestForAudience(r *http.Request, audience string) (int, tools.UserAccessLevel, error)
}

type UserServiceImpl struct {
	UserRepo       UserRepository
	SessionService SessionService
	AuthHelper     u.AuthHelper
	OsWrapper      u.OsWrapper
}

type InvitationDetails struct {
	Token          string
	ExpirationDate time.Time
}

func (s *UserServiceImpl) ChangeUsername(userId int, newUsername string) error {
	user, err := s.UserRepo.GetUserById(userId)
	if err != nil {
		return err
	}
	if user.Username != newUsername {
		doesUsernameExist, err := s.UserRepo.DoesUserExist(newUsername)
		if err != nil {
			return err
		}
		if doesUsernameExist {
			return u.Logger.NewError(UserAlreadyExistsError)
		}
	}
	user.Username = newUsername
	return s.UserRepo.UpdateUser(user)
}

func (s *UserServiceImpl) ChangeEmail(userId int, newEmail string) error {
	user, err := s.UserRepo.GetUserById(userId)
	if err != nil {
		return err
	}
	if user.Email != newEmail {
		doesEmailExist, err := s.UserRepo.DoesEmailExist(newEmail)
		if err != nil {
			return err
		}
		if doesEmailExist {
			return u.Logger.NewError(EmailAlreadyExistsError)
		}
	}
	user.Email = newEmail
	return s.UserRepo.UpdateUser(user)
}

func (s *UserServiceImpl) IsPasswordCorrect(creds Credentials) bool {
	user, err := s.UserRepo.GetUserByUsername(creds.Username)
	if err != nil {
		u.Logger.Info(err)
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(creds.Password))
	if err != nil {
		u.Logger.Info(err)
		return false
	}
	return true
}

func (s *UserServiceImpl) Logout(userId int) error {
	return s.SessionService.DeleteSessionsByUserId(userId)
}

func (s *UserServiceImpl) UserResetsOwnPassword(userId int, currentPassword, newPassword string) error {
	user, err := s.UserRepo.GetUserById(userId)
	if err != nil {
		return err
	}
	if !s.AuthHelper.DoesMatchSaltedHash(currentPassword, user.HashedPassword) {
		return u.Logger.NewError(IncorrectCurrentPasswordError)
	}
	user.HashedPassword, err = s.AuthHelper.SaltAndHash(newPassword)
	if err != nil {
		return err
	}
	err = s.UserRepo.UpdateUser(user)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserServiceImpl) InviteUser(username, email string) (*InvitationDetails, error) {
	doesUsernameExist, err := s.UserRepo.DoesUserExist(username)
	if err != nil {
		return nil, err
	}
	if doesUsernameExist {
		return nil, u.Logger.NewError(UserAlreadyExistsError)
	}

	doesEmailExist, err := s.UserRepo.DoesEmailExist(email)
	if err != nil {
		return nil, err
	}
	if doesEmailExist {
		return nil, u.Logger.NewError(EmailAlreadyExistsError)
	}

	token, err := s.AuthHelper.GenerateSecret()
	if err != nil {
		return nil, err
	}
	now := s.OsWrapper.Now()
	expirationDate := now.AddDate(0, 0, InvitationValidityDays)
	user := NewUser(
		username,
		email,
		"",
		"",
		tools.DefaultTime,
		false,
		token,
		expirationDate,
		now,
	)
	_, err = s.UserRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}
	return &InvitationDetails{
		Token:          token,
		ExpirationDate: expirationDate,
	}, nil
}

func (s *UserServiceImpl) AcceptNewPasswordViaToken(password, setPasswordToken string) error {
	user, err := s.UserRepo.GetUserByToken(setPasswordToken)
	if err != nil {
		return err
	}

	if time.Now().After(user.SetPasswordTokenExpirationDate) {
		return u.Logger.NewError(TokenExpiredError)
	}

	user.HashedPassword, err = s.AuthHelper.SaltAndHash(password)
	if err != nil {
		return err
	}
	user.SetPasswordToken = ""
	user.SetPasswordTokenExpirationDate = tools.DefaultTime
	return s.UserRepo.UpdateUser(user)
}

func (s *UserServiceImpl) ResetPasswordOfUser(userIdToResetPassword, currentlyLoggedInUserId int) (string, error) {
	if userIdToResetPassword == currentlyLoggedInUserId {
		return "", u.Logger.NewError(AdminCanNotResetOwnPasswordError)
	}
	user, err := s.UserRepo.GetUserById(userIdToResetPassword)
	if err != nil {
		return "", err
	}

	user.SetPasswordToken, err = s.AuthHelper.GenerateSecret()
	if err != nil {
		return "", err
	}
	user.SetPasswordTokenExpirationDate = time.Now().AddDate(0, 0, 1).UTC()
	user.HashedPassword = ""
	err = s.UserRepo.UpdateUser(user)
	if err != nil {
		return "", err
	}
	if err = s.SessionService.DeleteSessionsByUserId(userIdToResetPassword); err != nil {
		return "", err
	}

	return user.SetPasswordToken, nil
}

func (s *UserServiceImpl) DeleteUser(idOfUserToBeDeleted, currentlyLoggedInUserId int) error {
	if idOfUserToBeDeleted == currentlyLoggedInUserId {
		return u.Logger.NewError(AdminCanNotDeleteOwnAccountError)
	}
	return s.UserRepo.DeleteUser(idOfUserToBeDeleted)
}

func (s *UserServiceImpl) GetUserIdAndRoleFromQuollixRequest(r *http.Request) (int, tools.UserAccessLevel, error) {
	return s.GetUserIdAndRoleFromRequestForAudience(r, QuollixSessionAudience())
}

func (s *UserServiceImpl) GetUserIdAndRoleFromRequestForAudience(r *http.Request, audience string) (int, tools.UserAccessLevel, error) {
	cookie, err := r.Cookie(tools.BrandAppAuthCookieName)
	if err != nil {
		return AnonymousUserId, tools.AnonymousLevel, nil
	}
	if cookie == nil {
		return AnonymousUserId, tools.AnonymousLevel, nil
	}
	authenticatedSession, err := s.SessionService.GetAuthenticatedSession(cookie.Value, audience)
	if err != nil {
		return AnonymousUserId, tools.AnonymousLevel, nil
	}
	if authenticatedSession.Session.CookieExpirationDate.Before(time.Now().UTC()) {
		return AnonymousUserId, tools.AnonymousLevel, nil
	}
	if authenticatedSession.User.IsAdmin {
		return authenticatedSession.User.Id, tools.AdminLevel, nil
	} else {
		return authenticatedSession.User.Id, tools.UserLevel, nil
	}
}

func NewUser(
	username string,
	email string,
	hashedPassword string,
	hashedCookieValue string,
	cookieExpirationDate time.Time,
	isAdmin bool,
	setPasswordToken string,
	setPasswordTokenExpirationDate time.Time,
	creationDate time.Time,
) *tools.User {
	return &tools.User{
		Username:                       username,
		Email:                          email,
		HashedPassword:                 hashedPassword,
		HashedCookieValue:              hashedCookieValue,
		CookieExpirationDate:           cookieExpirationDate,
		IsAdmin:                        isAdmin,
		SetPasswordToken:               setPasswordToken,
		SetPasswordTokenExpirationDate: setPasswordTokenExpirationDate,
		CreationDate:                   creationDate,
	}
}
