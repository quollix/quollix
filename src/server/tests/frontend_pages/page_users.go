package frontend_pages

import (
	"server/tools"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
)

type UsersPage struct {
	Frame *FrameType
}

type UserListEntry struct {
	Name                                string
	Email                               string
	Role                                string
	IsEnabled                           bool
	Created                             string
	InvitationExpiration                string
	PasswordLinkCellText                string
	PasswordLinkPresent                 bool
	EditButtonPresent                   bool
	ResetButtonPresent                  bool
	SendPasswordResetEmailButtonPresent bool
	DeleteButtonPresent                 bool
}

const (
	createUsernameInputSelector      = "#create-user-name-input"
	createUserEmailInputSelector     = "#create-user-email-input"
	createUserButtonSelector         = "#create-user-button"
	createUserViaEmailButtonSelector = "#create-user-via-email-button"
)

func (u *UsersPage) CreateUser(username, email string) *UsersPage {
	u.Frame.Assert.PagePath(tools.Paths.FrontendUsers)
	u.Frame.Page.MustElement(createUsernameInputSelector).MustInput(username)
	u.Frame.Page.MustElement(createUserEmailInputSelector).MustInput(email)
	u.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		u.Frame.Page.MustElement(createUserButtonSelector).MustClick()
	})
	u.Frame.Assert.PagePath(tools.Paths.FrontendUsers)
	return u
}

func (u *UsersPage) CreateUserViaEmail(username, email string) *UsersPage {
	u.Frame.Assert.PagePath(tools.Paths.FrontendUsers)
	u.Frame.Page.MustElement(createUsernameInputSelector).MustInput(username)
	u.Frame.Page.MustElement(createUserEmailInputSelector).MustInput(email)
	u.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		u.Frame.Page.MustElement(createUserViaEmailButtonSelector).MustClick()
	})
	u.Frame.Assert.PagePath(tools.Paths.FrontendUsers)
	return u
}

func (u *UsersPage) ListUsers() []UserListEntry {
	return listUsers(u.Frame.Page, u.Frame.T)
}

func listUsers(page *rod.Page, t *testing.T) []UserListEntry {
	rows, err := page.Elements(`tr.user-row`)
	assert.Nil(t, err)

	out := make([]UserListEntry, 0, len(rows))
	for _, row := range rows {
		nameCell, err := row.Element(".user-name-cell")
		assert.Nil(t, err)
		name, err := nameCell.Text()
		assert.Nil(t, err)

		emailCell, err := row.Element(".user-email-cell")
		assert.Nil(t, err)
		email, err := emailCell.Text()
		assert.Nil(t, err)

		roleCell, err := row.Element(".user-role-cell")
		assert.Nil(t, err)
		role, err := roleCell.Text()
		assert.Nil(t, err)

		enabledCell, err := row.Element(".user-enabled-cell")
		assert.Nil(t, err)
		enabledCheckbox, err := enabledCell.Element(".user-enabled-checkbox")
		assert.Nil(t, err)
		isEnabled, err := enabledCheckbox.Property("checked")
		assert.Nil(t, err)

		createdCell, err := row.Element(".user-created-cell")
		assert.Nil(t, err)
		created, err := createdCell.Text()
		assert.Nil(t, err)

		invitationCell, err := row.Element(".user-invitation-expiration-cell")
		assert.Nil(t, err)
		invitationExpiration, err := invitationCell.Text()
		assert.Nil(t, err)

		passwordLinkCell, err := row.Element(".user-password-link-cell")
		assert.Nil(t, err)
		passwordLinkCellText, err := passwordLinkCell.Text()
		assert.Nil(t, err)
		passwordLinkPresent, _, err := passwordLinkCell.Has(".copy-to-clipboard-button")
		assert.Nil(t, err)

		actionsCell, err := row.Element(".user-actions-cell")
		assert.Nil(t, err)
		editButtonPresent, _, err := actionsCell.Has(`button.user-edit-button`)
		assert.Nil(t, err)
		resetButtonPresent, _, err := actionsCell.Has(`button.user-reset-password-button`)
		assert.Nil(t, err)
		sendPasswordResetEmailButtonPresent, _, err := actionsCell.Has(`button.user-send-password-reset-email-button`)
		assert.Nil(t, err)
		deleteButtonPresent, _, err := actionsCell.Has(`button.user-delete-button`)
		assert.Nil(t, err)

		out = append(out, UserListEntry{
			Name:                                strings.TrimSpace(name),
			Email:                               strings.TrimSpace(email),
			Role:                                strings.TrimSpace(role),
			IsEnabled:                           isEnabled.Bool(),
			Created:                             strings.TrimSpace(created),
			InvitationExpiration:                strings.TrimSpace(invitationExpiration),
			PasswordLinkCellText:                strings.TrimSpace(passwordLinkCellText),
			PasswordLinkPresent:                 passwordLinkPresent,
			EditButtonPresent:                   editButtonPresent,
			ResetButtonPresent:                  resetButtonPresent,
			SendPasswordResetEmailButtonPresent: sendPasswordResetEmailButtonPresent,
			DeleteButtonPresent:                 deleteButtonPresent,
		})
	}
	return out
}

func (u *UsersPage) AssertUserInList(username, email string) *UsersPage {
	p := u.Frame.Page.Timeout(defaultTimeout)
	defer p.CancelTimeout()
	for {
		users := listUsers(p, u.Frame.T)
		for _, entry := range users {
			if entry.Name == username && entry.Email == email {
				return u
			}
		}
		if p.GetContext().Err() != nil {
			u.Frame.T.Fatalf("user not found in frontend list (username=%s email=%s)", username, email)
		}
		time.Sleep(50 * time.Millisecond)
	}
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
		assert.Equal(u.Frame.T, isEnabled, u.Frame.Controls.GetCheckboxValue(rowCheckboxSelector(username)))
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
		u.Frame.Assert.PagePath(tools.Paths.FrontendUserEdit)
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
