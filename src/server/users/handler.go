package users

import (
	"net/http"
	"server/tools"
	"strconv"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

const (
	IncorrectLoginCredentialsError   = "Incorrect username or password"
	UserAlreadyExistsError           = "user already exists"
	EmailAlreadyExistsError          = "email already exists"
	AdminCanNotDeleteOwnAccountError = "admin cannot delete own account"
	AdminCanNotResetOwnPasswordError = "admin cannot reset own password"
)

var (
	ExpectedUserDeletionErrors  = u.MapOf(AdminCanNotDeleteOwnAccountError)
	ExpectedPasswordResetErrors = u.MapOf(AdminCanNotResetOwnPasswordError)
	expectedCookieNotFoundError = u.MapOf("cookie not found")
	expectedTokenExpiredErrors  = u.MapOf(TokenExpiredError, UserNotFoundError)
)

type UserHandler struct {
	UserRepo       UserRepository
	AuthService    AuthenticationService
	UserService    UserService
	SessionService SessionService
	SecretStorage  SecretAndCookieStorage
}

func (s *UserHandler) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := s.UserRepo.ListUsers()
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, users)
}

func (s *UserHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	userIdString, ok := validation.ReadBody[tools.NumberString](w, r)
	if !ok {
		return
	}

	idOfUserToBeDeleted, err := strconv.Atoi(userIdString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	currentlyLoggedInUser, err := GetAuthFromContext(r)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	if err := s.UserService.DeleteUser(idOfUserToBeDeleted, currentlyLoggedInUser.Id); err != nil {
		u.WriteResponseError(w, ExpectedUserDeletionErrors, err)
		return
	}
}

func (s *UserHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthFromContext(r)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	err = s.UserService.Logout(user.Id)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

var changeOwnPasswordExpectedErrors = u.MapOf(IncorrectCurrentPasswordError)

func (s *UserHandler) UserChangesOwnPasswordHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[tools.ChangeOwnPasswordRequest](w, r)
	if !ok {
		return
	}

	user, err := GetAuthFromContext(r)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	err = s.UserService.UserResetsOwnPassword(user.Id, request.CurrentPassword, request.NewPassword)
	if err != nil {
		u.WriteResponseError(w, changeOwnPasswordExpectedErrors, err)
		return
	}
}

func (s *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	creds, ok := validation.ReadBody[Credentials](w, r)
	if !ok {
		return
	}

	if !s.UserService.IsPasswordCorrect(*creds) {
		u.WriteResponseErrorAlways(w, u.Logger.NewError(IncorrectLoginCredentialsError))
		return
	}

	user, err := s.UserRepo.GetUserByUsername(creds.Username)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	cookie, err := s.SessionService.GenerateAndSaveCookie(user.Id, QuollixSessionAudience())
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	http.SetCookie(w, cookie)
}

func (s *UserHandler) SecretHandler(w http.ResponseWriter, r *http.Request) {
	u.Logger.Debug("SecretHandler called")
	cookie, err := r.Cookie(tools.BrandAppAuthCookieName)
	if err != nil {
		u.WriteResponseError(w, expectedCookieNotFoundError, err)
		return
	}

	secret, err := s.SecretStorage.GenerateSecretForCookie(cookie.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	u.SendJsonResponse(w, secret)
}

func (s *UserHandler) CheckAuthHandler(w http.ResponseWriter, r *http.Request) {
	r, err := s.AuthService.GetRequestWithAuthContext(w, r)
	if err != nil {
		u.WriteResponseError(w, expectedCookieNotFoundError, err)
		return
	}

	user, ok := r.Context().Value(tools.AuthKey).(tools.User)
	if !ok {
		u.WriteResponseErrorAlways(w, u.Logger.NewError(AuthNotFoundInContextError))
		return
	}
	u.SendJsonResponse(w, user)
}

var expectedUserInvitationError = u.MapOf(UserAlreadyExistsError, EmailAlreadyExistsError)
var expectedChangeUsernameError = u.MapOf(UserAlreadyExistsError)
var expectedChangeEmailError = u.MapOf(EmailAlreadyExistsError)

type InviteUserRequest struct {
	Username string `json:"username" validate:"username"`
	Email    string `json:"email" validate:"email"`
}

func (s *UserHandler) InviteUserHandler(w http.ResponseWriter, r *http.Request) {
	userInvitationRequest, ok := validation.ReadBody[InviteUserRequest](w, r)
	if !ok {
		return
	}
	_, err := s.UserService.InviteUser(userInvitationRequest.Username, userInvitationRequest.Email)
	if err != nil {
		u.WriteResponseError(w, expectedUserInvitationError, err)
		return
	}
}

type AcceptNewPasswordViaTokenRequest struct {
	Password string `validate:"password"`
	Token    string `validate:"secret"`
}

func (s *UserHandler) AcceptNewPasswordViaTokenHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[AcceptNewPasswordViaTokenRequest](w, r)
	if !ok {
		return
	}
	err := s.UserService.AcceptNewPasswordViaToken(request.Password, request.Token)
	if err != nil {
		u.WriteResponseError(w, expectedTokenExpiredErrors, err)
		return
	}
}

func (s *UserHandler) ResetPasswordAndCreateTokenHandler(w http.ResponseWriter, r *http.Request) {
	userIdString, ok := validation.ReadBody[tools.NumberString](w, r)
	if !ok {
		return
	}
	userIdToResetPassword, err := strconv.Atoi(userIdString.Value)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	currentlyLoggedInUser, err := GetAuthFromContext(r)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	_, err = s.UserService.ResetPasswordOfUser(userIdToResetPassword, currentlyLoggedInUser.Id)
	if err != nil {
		u.WriteResponseError(w, ExpectedPasswordResetErrors, err)
		return
	}
}

type ChangeUsernameRequest struct {
	UserId   string `json:"user_id" validate:"number"`
	Username string `json:"username" validate:"username"`
}

func (s *UserHandler) ChangeUsernameHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[ChangeUsernameRequest](w, r)
	if !ok {
		return
	}

	userIdInt, err := strconv.Atoi(request.UserId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	err = s.UserService.ChangeUsername(userIdInt, request.Username)
	if err != nil {
		u.WriteResponseError(w, expectedChangeUsernameError, err)
		return
	}
}

type ChangeEmailRequest struct {
	UserId   string `json:"user_id" validate:"number"`
	NewEmail string `json:"new_email" validate:"email"`
}

func (s *UserHandler) ChangeEmailHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[ChangeEmailRequest](w, r)
	if !ok {
		return
	}

	userIdInt, err := strconv.Atoi(request.UserId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	err = s.UserService.ChangeEmail(userIdInt, request.NewEmail)
	if err != nil {
		u.WriteResponseError(w, expectedChangeEmailError, err)
		return
	}
}
