package frontend_pages

import "github.com/quollix/common/assert"

type UserEditPage struct {
	Frame *FrameType
}

func (u *UserEditPage) AssertCurrentValues(expectedUsername, expectedEmail string) *UserEditPage {
	assert.Equal(u.Frame.T, expectedUsername, u.GetUsername())
	assert.Equal(u.Frame.T, expectedEmail, u.GetEmail())
	return u
}

func (u *UserEditPage) GetUsername() string {
	return u.Frame.Controls.GetInputValue("#change-username-input")
}

func (u *UserEditPage) GetEmail() string {
	return u.Frame.Controls.GetInputValue("#change-email-input")
}

func (u *UserEditPage) ChangeUsername(newUsername string) *UserEditPage {
	u.Frame.Controls.SetInputValue("#change-username-input", newUsername)

	saveButton := u.Frame.Controls.GetRequiredElement("#change-username-save-button")
	saveButton.MustClick()
	u.Frame.Browser.ConfirmDialog()
	u.Frame.Assert.SnackbarVisibleWithTextEventually("Username updated.")
	return u
}

func (u *UserEditPage) ChangeEmail(newEmail string) *UserEditPage {
	u.Frame.Controls.SetInputValue("#change-email-input", newEmail)

	saveButton := u.Frame.Controls.GetRequiredElement("#change-email-save-button")
	saveButton.MustClick()
	u.Frame.Browser.ConfirmDialog()
	u.Frame.Assert.SnackbarVisibleWithTextEventually("Email updated.")
	return u
}
