package component

import (
	"encoding/json"
	"server/tools"
	"server/users"
)

type QuollixAuthClient struct {
	quollix *QuollixClient
}

func (c *QuollixAuthClient) Login(username, password string) error {
	loginCredentials := users.Credentials{Username: username, Password: password}
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendLogin, loginCredentials)
	return err
}

func (c *QuollixAuthClient) Logout() error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersLogout, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *QuollixAuthClient) GetCurrentUser() (*tools.User, error) {
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendCheckAuth, nil)
	if err != nil {
		return nil, err
	}
	var authResponse tools.User
	err = json.Unmarshal(responseBody, &authResponse)
	if err != nil {
		return nil, err
	}
	return &authResponse, nil
}

func (c *QuollixAuthClient) ChangePassword(oldPassword, newPassword string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersChangeOwnPassword, tools.ChangeOwnPasswordRequest{
		CurrentPassword: oldPassword,
		NewPassword:     newPassword,
	})
	if err != nil {
		return err
	}
	return nil
}
