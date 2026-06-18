//go:build acceptance

package pages

import "github.com/quollix/common/assert"

type UserEditPage struct {
	Frame *FrameType
}

func (u *UserEditPage) AssertCurrentValues(expectedUsername, expectedEmail string) *UserEditPage {
	assert.Equal(u.Frame.t, expectedUsername, u.GetUsername())
	assert.Equal(u.Frame.t, expectedEmail, u.GetEmail())
	return u
}

func (u *UserEditPage) GetUsername() string {
	return GetInputValue(u.Frame.t, u.Frame.page, "#change-username-input")
}

func (u *UserEditPage) GetEmail() string {
	return GetInputValue(u.Frame.t, u.Frame.page, "#change-email-input")
}

func (u *UserEditPage) ChangeUsername(newUsername string) *UserEditPage {
	SetInputValue(u.Frame.t, u.Frame.page, "#change-username-input", newUsername)

	saveButton := GetRequiredElement(u.Frame.t, u.Frame.page, "#change-username-save-button")
	saveButton.MustClick()
	u.Frame.ConfirmDialog()
	u.Frame.AssertSnackbarVisibleWithTextEventually("Username updated.")
	return u
}

func (u *UserEditPage) ChangeEmail(newEmail string) *UserEditPage {
	SetInputValue(u.Frame.t, u.Frame.page, "#change-email-input", newEmail)

	saveButton := GetRequiredElement(u.Frame.t, u.Frame.page, "#change-email-save-button")
	saveButton.MustClick()
	u.Frame.ConfirmDialog()
	u.Frame.AssertSnackbarVisibleWithTextEventually("Email updated.")
	return u
}
