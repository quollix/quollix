package users

import (
	"net/http"
	"server/tools"
	"strconv"

	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

const (
	IncorrectLoginCredentialsError    = "Incorrect username or password"
	UserAlreadyExistsError            = "user already exists"
	EmailAlreadyExistsError           = "email already exists"
	AdminCanNotDeleteOwnAccountError  = "admin cannot delete own account"
	AdminCanNotResetOwnPasswordError  = "admin cannot reset own password"
	AdminCanNotDisableOwnAccountError = "admin cannot disable own account"
)

var (
	ExpectedUserDeletionErrors   = u.MapOf(AdminCanNotDeleteOwnAccountError)
	ExpectedPasswordResetErrors  = u.MapOf(AdminCanNotResetOwnPasswordError)
	ExpectedSetUserEnabledErrors = u.MapOf(AdminCanNotDisableOwnAccountError)
	expectedCookieNotFoundError  = u.MapOf("cookie not found")
	expectedTokenExpiredErrors   = u.MapOf(TokenExpiredError, UserNotFoundError)
	expectedLoginErrors          = u.MapOf(IncorrectLoginCredentialsError, UserDisabledError)
	expectedSetOwnPasswordErrors = u.MapOf(LocalPasswordAlreadySetError)
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
	userIdString, ok := validation.ReadBody[api.NumberString](w, r)
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

func (s *UserHandler) SignOutHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetAuthFromContext(r)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	err = s.UserService.SignOut(user.Id)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
}

var changeOwnPasswordExpectedErrors = u.MapOf(IncorrectCurrentPasswordError)

func (s *UserHandler) UserSetsOwnPasswordHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[api.SetOwnPasswordRequest](w, r)
	if !ok {
		return
	}

	user, err := GetAuthFromContext(r)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	err = s.UserService.UserSetsOwnPassword(user.Id, request.NewPassword)
	if err != nil {
		u.WriteResponseError(w, expectedSetOwnPasswordErrors, err)
		return
	}
}

func (s *UserHandler) UserChangesOwnPasswordHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[api.ChangeOwnPasswordRequest](w, r)
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

func (s *UserHandler) SignInHandler(w http.ResponseWriter, r *http.Request) {
	creds, ok := validation.ReadBody[api.Credentials](w, r)
	if !ok {
		return
	}

	cookie, err := s.UserService.SignIn(*creds)
	if err != nil {
		u.WriteResponseError(w, expectedLoginErrors, err)
		return
	}
	http.SetCookie(w, cookie)
}

func (s *UserHandler) SecretHandler(w http.ResponseWriter, r *http.Request) {
	u.Logger.Debug("SecretHandler called")
	cookie, err := r.Cookie(api.BrandAppAuthCookieName)
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

	user, ok := r.Context().Value(tools.AuthKey).(api.User)
	if !ok {
		u.WriteResponseErrorAlways(w, u.Logger.NewError(AuthNotFoundInContextError))
		return
	}
	u.SendJsonResponse(w, user)
}

var expectedUserInvitationError = u.MapOf(UserAlreadyExistsError, EmailAlreadyExistsError, ReservedEmailMustMatchUserError)
var expectedChangeUsernameError = u.MapOf(UserAlreadyExistsError, ReservedEmailRenameConflictError)
var expectedChangeEmailError = u.MapOf(EmailAlreadyExistsError, ReservedEmailMustMatchUserError)

func (s *UserHandler) InviteUserHandler(w http.ResponseWriter, r *http.Request) {
	userInvitationRequest, ok := validation.ReadBody[api.InviteUserRequest](w, r)
	if !ok {
		return
	}
	_, err := s.UserService.InviteUser(userInvitationRequest.Username, userInvitationRequest.Email)
	if err != nil {
		u.WriteResponseError(w, expectedUserInvitationError, err)
		return
	}
}

func (s *UserHandler) AcceptNewPasswordViaTokenHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[api.AcceptNewPasswordViaTokenRequest](w, r)
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
	userIdString, ok := validation.ReadBody[api.NumberString](w, r)
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

func (s *UserHandler) ChangeUsernameHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[api.ChangeUsernameRequest](w, r)
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

func (s *UserHandler) ChangeEmailHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[api.ChangeEmailRequest](w, r)
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

func (s *UserHandler) SetUserEnabledHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := validation.ReadBody[api.SetUserEnabledRequest](w, r)
	if !ok {
		return
	}

	userId, err := strconv.Atoi(request.UserId)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	currentlyLoggedInUser, err := GetAuthFromContext(r)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	err = s.UserService.SetUserEnabled(userId, request.IsEnabled, currentlyLoggedInUser.Id)
	if err != nil {
		u.WriteResponseError(w, ExpectedSetUserEnabledErrors, err)
		return
	}
}
