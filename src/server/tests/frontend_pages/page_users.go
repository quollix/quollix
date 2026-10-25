package frontend_pages

import (
	"strings"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/browsertest"
	"github.com/quollix/common/quollix/api"
	utils "github.com/quollix/common/utils"
)

type UsersPage struct {
	Frame *FrameType
}

type UserListEntry struct {
	Name                                string
	Email                               string
	Role                                string
	CanChangeRole                       bool
	CanPromoteToAdmin                   bool
	IsEnabled                           bool
	CanEnableOrDisableUser              bool
	CanDeleteUser                       bool
	Created                             string
	InvitationExpiration                string
	PasswordLinkCellText                string
	PasswordLinkPresent                 bool
	EditButtonPresent                   bool
	ResetButtonPresent                  bool
	SendPasswordResetEmailButtonPresent bool
}

const (
	createUsernameInputSelector      = "#create-user-name-input"
	createUserEmailInputSelector     = "#create-user-email-input"
	createUserButtonSelector         = "#create-user-button"
	createUserViaEmailButtonSelector = "#create-user-via-email-button"
)

func (u *UsersPage) CreateUser(username, email string) *UsersPage {
	u.Frame.Assert.PagePath(api.Paths.FrontendUsers)
	u.Frame.Controls.GetRequiredElement(createUsernameInputSelector).MustInput(username)
	u.Frame.Controls.GetRequiredElement(createUserEmailInputSelector).MustInput(email)
	u.Frame.Controls.GetRequiredElement(createUserButtonSelector).MustClick()
	u.Frame.Assert.SnackbarVisibleWithTextEventually("User invited successfully.")
	u.Frame.Assert.PagePath(api.Paths.FrontendUsers)
	return u
}

func (u *UsersPage) CreateUserViaEmail(username, email string) *UsersPage {
	u.Frame.Assert.PagePath(api.Paths.FrontendUsers)
	u.Frame.Controls.GetRequiredElement(createUsernameInputSelector).MustInput(username)
	u.Frame.Controls.GetRequiredElement(createUserEmailInputSelector).MustInput(email)
	u.Frame.Controls.GetRequiredElement(createUserViaEmailButtonSelector).MustClick()
	u.Frame.Assert.SnackbarVisibleWithTextEventually("Invitation email sent successfully.")
	u.Frame.Assert.PagePath(api.Paths.FrontendUsers)
	return u
}

func (u *UsersPage) SetInviteEmail(email string) *UsersPage {
	u.Frame.Assert.PagePath(api.Paths.FrontendUsers)
	u.Frame.Page.MustElement(createUserEmailInputSelector).MustSelectAllText().MustInput(email)
	return u
}

func (u *UsersPage) AssertInviteEmailRequired(expected bool) *UsersPage {
	required, err := u.Frame.Page.MustElement(createUserEmailInputSelector).Property("required")
	assert.Nil(u.Frame.T, err)
	assert.Equal(u.Frame.T, expected, required.Bool())
	return u
}

func (u *UsersPage) AssertInviteViaEmailButtonDisabled(expected bool) *UsersPage {
	disabled, err := u.Frame.Page.MustElement(createUserViaEmailButtonSelector).Property("disabled")
	assert.Nil(u.Frame.T, err)
	assert.Equal(u.Frame.T, expected, disabled.Bool())
	return u
}

func (u *UsersPage) ListUsers() []UserListEntry {
	users, err := tryListUsers(u.Frame.Page)
	assert.Nil(u.Frame.T, err)
	return users
}

func tryListUsers(page *browsertest.Page) ([]UserListEntry, error) {
	rows, err := page.Elements(`tr.user-row`)
	if err != nil {
		return nil, err
	}

	out := make([]UserListEntry, 0, len(rows))
	for _, row := range rows {
		nameCell, err := row.Element(".user-name-cell")
		if err != nil {
			return nil, err
		}
		name, err := nameCell.Text()
		if err != nil {
			return nil, err
		}

		emailCell, err := row.Element(".user-email-cell")
		if err != nil {
			return nil, err
		}
		email, err := emailCell.Text()
		if err != nil {
			return nil, err
		}

		roleCell, err := row.Element(".user-role-cell")
		if err != nil {
			return nil, err
		}
		roleSelectPresent, _, err := roleCell.Has("select.user-role-select")
		if err != nil {
			return nil, err
		}
		roleDisabled := true
		adminOptionDisabled := true
		var role string
		if roleSelectPresent {
			roleSelect, err := roleCell.Element("select.user-role-select")
			if err != nil {
				return nil, err
			}
			roleValue, err := roleSelect.Property("value")
			if err != nil {
				return nil, err
			}
			role = roleValue.String()
			roleDisabledValue, err := roleSelect.Property("disabled")
			if err != nil {
				return nil, err
			}
			roleDisabled = roleDisabledValue.Bool()
			adminOption, err := roleSelect.Element(`option[value="admin"]`)
			if err != nil {
				return nil, err
			}
			adminOptionDisabledValue, err := adminOption.Property("disabled")
			if err != nil {
				return nil, err
			}
			adminOptionDisabled = adminOptionDisabledValue.Bool()
		} else {
			roleText, err := roleCell.Text()
			if err != nil {
				return nil, err
			}
			role = strings.ToLower(strings.TrimSpace(roleText))
		}

		enabledCell, err := row.Element(".user-enabled-cell")
		if err != nil {
			return nil, err
		}
		enabledCheckbox, err := enabledCell.Element(".user-enabled-checkbox")
		if err != nil {
			return nil, err
		}
		isEnabled, err := enabledCheckbox.Property("checked")
		if err != nil {
			return nil, err
		}
		enabledControlDisabled, err := enabledCheckbox.Property("disabled")
		if err != nil {
			return nil, err
		}

		createdCell, err := row.Element(".user-created-cell")
		if err != nil {
			return nil, err
		}
		created, err := createdCell.Text()
		if err != nil {
			return nil, err
		}

		invitationCell, err := row.Element(".user-invitation-expiration-cell")
		if err != nil {
			return nil, err
		}
		invitationExpiration, err := invitationCell.Text()
		if err != nil {
			return nil, err
		}

		passwordLinkCell, err := row.Element(".user-password-link-cell")
		if err != nil {
			return nil, err
		}
		passwordLinkCellText, err := passwordLinkCell.Text()
		if err != nil {
			return nil, err
		}
		passwordLinkPresent, _, err := passwordLinkCell.Has(".copy-to-clipboard-button")
		if err != nil {
			return nil, err
		}

		actionsCell, err := row.Element(".user-actions-cell")
		if err != nil {
			return nil, err
		}
		editButtonPresent, _, err := actionsCell.Has(`button.user-edit-button`)
		if err != nil {
			return nil, err
		}
		resetButtonPresent, _, err := actionsCell.Has(`button.user-reset-password-button`)
		if err != nil {
			return nil, err
		}
		sendPasswordResetEmailButtonPresent, _, err := actionsCell.Has(`button.user-send-password-reset-email-button`)
		if err != nil {
			return nil, err
		}
		deleteButton, err := actionsCell.Element(`button.user-delete-button`)
		if err != nil {
			return nil, err
		}
		deleteButtonDisabled, err := deleteButton.Property("disabled")
		if err != nil {
			return nil, err
		}

		out = append(out, UserListEntry{
			Name:                                strings.TrimSpace(name),
			Email:                               strings.TrimSpace(email),
			Role:                                strings.TrimSpace(role),
			CanChangeRole:                       roleSelectPresent && !roleDisabled,
			CanPromoteToAdmin:                   roleSelectPresent && !adminOptionDisabled,
			IsEnabled:                           isEnabled.Bool(),
			CanEnableOrDisableUser:              !enabledControlDisabled.Bool(),
			CanDeleteUser:                       !deleteButtonDisabled.Bool(),
			Created:                             strings.TrimSpace(created),
			InvitationExpiration:                strings.TrimSpace(invitationExpiration),
			PasswordLinkCellText:                strings.TrimSpace(passwordLinkCellText),
			PasswordLinkPresent:                 passwordLinkPresent,
			EditButtonPresent:                   editButtonPresent,
			ResetButtonPresent:                  resetButtonPresent,
			SendPasswordResetEmailButtonPresent: sendPasswordResetEmailButtonPresent,
		})
	}
	return out, nil
}

func (u *UsersPage) SetUserRole(username, role string) *UsersPage {
	selectElement := u.Frame.Controls.GetRequiredElement(`tr.user-row[data-username="` + username + `"] select.user-role-select`)
	selectElement.MustSelect(role)
	u.Frame.Browser.ConfirmDialog()
	expectedRole := strings.ToLower(role)
	expectedCanEnableOrDisableUser := expectedRole == "user"
	err := utils.EventuallyWithTimeout(browserTimeout, 50*time.Millisecond, func() error {
		users, err := tryListUsers(u.Frame.Page)
		if err != nil {
			return err
		}
		for _, user := range users {
			if user.Name != username {
				continue
			}
			if user.Role == expectedRole && user.CanEnableOrDisableUser == expectedCanEnableOrDisableUser {
				return nil
			}
			return utils.Logger.NewError("user role and controls have not updated yet")
		}
		return utils.Logger.NewError("user role and controls have not updated yet")
	})
	assert.Nil(u.Frame.T, err)
	u.Frame.Assert.SnackbarVisibleWithTextEventually("'" + username + "' role updated.")
	return u
}

func (u *UsersPage) AssertUserInList(username, email string) *UsersPage {
	err := utils.EventuallyWithTimeout(defaultTimeout, 50*time.Millisecond, func() error {
		users, err := tryListUsers(u.Frame.Page)
		if err != nil {
			return err
		}
		for _, entry := range users {
			if entry.Name == username && entry.Email == email {
				return nil
			}
		}
		return utils.Logger.NewError("user not found in frontend list")
	})
	assert.Nil(u.Frame.T, err)
	return u
}

func (u *UsersPage) GetRequiredUser(username string) *UserListEntry {
	users := u.ListUsers()
	for _, user := range users {
		if user.Name == username {
			userCopy := user
			return &userCopy
		}
	}
	assert.True(u.Frame.T, false)
	return nil
}

func (u *UsersPage) SetUserEnabled(username string, isEnabled bool) *UsersPage {
	rows, err := u.Frame.Page.Elements(`tr.user-row`)
	assert.Nil(u.Frame.T, err)

	for _, row := range rows {
		nameCell, cellErr := row.Element(".user-name-cell")
		assert.Nil(u.Frame.T, cellErr)
		name, textErr := nameCell.Text()
		assert.Nil(u.Frame.T, textErr)
		if strings.TrimSpace(name) != username {
			continue
		}

		checkbox, checkboxErr := row.Element(".user-enabled-checkbox")
		assert.Nil(u.Frame.T, checkboxErr)
		checkbox.MustClick()
		if !isEnabled {
			u.Frame.Browser.ConfirmDialog()
		}
		status := "disabled"
		if isEnabled {
			status = "enabled"
		}
		u.Frame.Assert.SnackbarVisibleWithTextEventually("User " + status + " successfully.")
		assert.Equal(u.Frame.T, isEnabled, u.Frame.Controls.GetCheckboxValue(rowCheckboxSelector(username)))
		assert.Equal(u.Frame.T, isEnabled, u.GetRequiredUser(username).CanPromoteToAdmin)
		return u
	}

	u.Frame.T.Fatalf("user not found in user table: %s", username)
	return nil
}

func rowCheckboxSelector(username string) string {
	return `tr.user-row[data-username="` + username + `"] .user-enabled-checkbox`
}

func (u *UsersPage) OpenEditPageForUser(username string) *UserEditPage {
	rows, err := u.Frame.Page.Elements(`tr.user-row`)
	assert.Nil(u.Frame.T, err)

	for _, row := range rows {
		nameCell, err := row.Element(".user-name-cell")
		assert.Nil(u.Frame.T, err)
		name, err := nameCell.Text()
		assert.Nil(u.Frame.T, err)
		if strings.TrimSpace(name) != username {
			continue
		}

		editButton, err := row.Element("button.user-edit-button")
		assert.Nil(u.Frame.T, err)
		u.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
			editButton.MustClick()
		})
		u.Frame.Assert.PagePath(api.Paths.FrontendUserEdit)
		return u.Frame.Pages.UserEditPage
	}

	u.Frame.T.Fatalf("user not found in user table: %s", username)
	return nil
}

func (u *UsersPage) SendPasswordResetEmail(username string) *UsersPage {
	rows, err := u.Frame.Page.Elements(`tr.user-row`)
	assert.Nil(u.Frame.T, err)

	for _, row := range rows {
		nameCell, cellErr := row.Element(".user-name-cell")
		assert.Nil(u.Frame.T, cellErr)
		name, textErr := nameCell.Text()
		assert.Nil(u.Frame.T, textErr)
		if strings.TrimSpace(name) != username {
			continue
		}

		button, buttonErr := row.Element("button.user-send-password-reset-email-button")
		assert.Nil(u.Frame.T, buttonErr)
		button.MustClick()
		u.Frame.Browser.ConfirmDialog()
		u.Frame.Assert.SnackbarVisibleWithTextEventually("Password reset email sent successfully.")
		return u
	}

	u.Frame.T.Fatalf("user not found in user table: %s", username)
	return nil
}
