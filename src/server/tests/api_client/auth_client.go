package api_client

import (
	"encoding/json"

	"server/tools"
	"server/users"
)

type AuthClient struct {
	quollix *QuollixClient
}

func (c *AuthClient) SignIn(username, password string) error {
	signInCredentials := users.Credentials{Username: username, Password: password}
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSignIn, signInCredentials)
	return err
}

func (c *AuthClient) SignOut() error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersSignOut, nil)
	return err
}

func (c *AuthClient) GetCurrentUser() (*tools.User, error) {
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

func (c *AuthClient) ChangePassword(oldPassword, newPassword string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersChangeOwnPassword, tools.ChangeOwnPasswordRequest{
		CurrentPassword: oldPassword,
		NewPassword:     newPassword,
	})
	return err
}
