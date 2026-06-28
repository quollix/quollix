package api_client

import (
	"encoding/json"
	"strconv"

	"server/tools"
	"server/users"
)

type UsersClient struct {
	quollix *QuollixClient
}

func (c *UsersClient) List() ([]tools.User, error) {
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersList, nil)
	if err != nil {
		return nil, err
	}
	var users []tools.User
	err = json.Unmarshal(responseBody, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (c *UsersClient) GetByUsername(username string) (tools.User, bool, error) {
	users, err := c.List()
	if err != nil {
		return tools.User{}, false, err
	}
	for _, user := range users {
		if user.Username == username {
			return user, true, nil
		}
	}
	return tools.User{}, false, nil
}

func (c *UsersClient) Delete(userId int) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersDelete, tools.NumberString{Value: strconv.Itoa(userId)})
	return err
}

func (c *UsersClient) Invite(username, email string) error {
	request := users.InviteUserRequest{Username: username, Email: email}
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersInviteUser, request)
	return err
}

func (c *UsersClient) SetPasswordViaToken(password, token string) error {
	request := users.AcceptNewPasswordViaTokenRequest{Password: password, Token: token}
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersSetPassword, request)
	return err
}

func (c *UsersClient) SetOwnPassword(password string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersSetOwnPassword, tools.SetOwnPasswordRequest{
		NewPassword: password,
	})
	return err
}

func (c *UsersClient) ResetPassword(userId int) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersResetPassword, tools.NumberString{Value: strconv.Itoa(userId)})
	return err
}

func (c *UsersClient) ChangeUsername(userId int, newUsername string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersChangeUsername, users.ChangeUsernameRequest{
		UserId:   strconv.Itoa(userId),
		Username: newUsername,
	})
	return err
}

func (c *UsersClient) ChangeEmail(userId int, newEmail string) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersChangeEmail, users.ChangeEmailRequest{
		UserId:   strconv.Itoa(userId),
		NewEmail: newEmail,
	})
	return err
}

func (c *UsersClient) SetEnabled(userId int, isEnabled bool) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersSetEnabled, users.SetUserEnabledRequest{
		UserId:    strconv.Itoa(userId),
		IsEnabled: isEnabled,
	})
	return err
}
