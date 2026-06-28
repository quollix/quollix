package frontend_pages

type AccountPage struct {
	Frame *FrameType
}

func (a *AccountPage) AssertSetPasswordFormPresent() *AccountPage {
	a.Frame.Assert.ElementPresent("#set-password-section")
	return a
}

func (a *AccountPage) AssertSetPasswordFormNotPresent() *AccountPage {
	a.Frame.Assert.ElementNotPresent("#set-password-section")
	return a
}

func (a *AccountPage) AssertChangePasswordFormPresent() *AccountPage {
	a.Frame.Assert.ElementPresent("#change-password-section")
	return a
}

func (a *AccountPage) AssertChangePasswordFormNotPresent() *AccountPage {
	a.Frame.Assert.ElementNotPresent("#change-password-section")
	return a
}

func (a *AccountPage) AssertSetPasswordFormState() *AccountPage {
	return a.AssertSetPasswordFormPresent().AssertChangePasswordFormNotPresent()
}

func (a *AccountPage) AssertChangePasswordFormState() *AccountPage {
	return a.AssertChangePasswordFormPresent().AssertSetPasswordFormNotPresent()
}

func (a *AccountPage) EnterSetPassword(newPassword, confirmPassword string) *AccountPage {
	a.Frame.Controls.SetInputValue("#set-password-new-input", newPassword)
	a.Frame.Controls.SetInputValue("#set-password-confirm-input", confirmPassword)
	return a
}

func (a *AccountPage) SaveSetPassword() *AccountPage {
	a.Frame.Controls.GetRequiredElement("#set-password-button").MustClick()
	return a
}

func (a *AccountPage) SaveSetPasswordAndWaitForReload() *AccountPage {
	a.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		a.SaveSetPassword()
	})
	return a
}

func (a *AccountPage) EnterChangePassword(currentPassword, newPassword, confirmPassword string) *AccountPage {
	a.Frame.Controls.SetInputValue("#change-password-current-input", currentPassword)
	a.Frame.Controls.SetInputValue("#change-password-new-input", newPassword)
	a.Frame.Controls.SetInputValue("#change-password-confirm-input", confirmPassword)
	return a
}

func (a *AccountPage) SaveChangePassword() *AccountPage {
	a.Frame.Controls.GetRequiredElement("#change-password-button").MustClick()
	return a
}

func (a *AccountPage) SaveChangePasswordAndWaitForReload() *AccountPage {
	a.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		a.SaveChangePassword()
	})
	return a
}
