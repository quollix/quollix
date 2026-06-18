package component

import (
	"encoding/json"
	"server/tools"
	"server/users"
	"strconv"

	"github.com/quollix/common/assert"
)

type QuollixUsersClient struct {
	quollix *QuollixClient
}

func (c *QuollixUsersClient) List() []tools.User {
	responseBody, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersList, nil)
	assert.Nil(c.quollix.T, err)
	var users []tools.User
	err = json.Unmarshal(responseBody, &users)
	assert.Nil(c.quollix.T, err)
	return users
}

func (c *QuollixUsersClient) Delete(userId int) error {
	userIdString := strconv.Itoa(userId)
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersDelete, tools.NumberString{Value: userIdString})
	if err != nil {
		return err
	}
	return nil
}

func (c *QuollixUsersClient) GetByUsername(username string) tools.User {
	users := c.List()
	for _, user := range users {
		if user.Username == username {
			return user
		}
	}
	c.quollix.T.Fatal("No user found")
	return tools.User{}
}

func (c *QuollixUsersClient) Invite(username, email string) error {
	request := users.InviteUserRequest{Username: username, Email: email}
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersInviteUser, request)
	if err != nil {
		return err
	}
	return nil
}

func (c *QuollixUsersClient) SetPasswordViaToken(password, token string) error {
	request := users.AcceptNewPasswordViaTokenRequest{Password: password, Token: token}
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersSetPassword, request)
	if err != nil {
		return err
	}
	return nil
}

func (c *QuollixUsersClient) ResetPassword(userId int) error {
	userIdString := strconv.Itoa(userId)
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersResetPassword, tools.NumberString{Value: userIdString})
	if err != nil {
		return err
	}
	return nil
}

func (c *QuollixUsersClient) ChangeUsername(userId int, newUsername string) error {
	userIdString := strconv.Itoa(userId)
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersChangeUsername, users.ChangeUsernameRequest{
		UserId:   userIdString,
		Username: newUsername,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *QuollixUsersClient) ChangeEmail(userId int, newEmail string) error {
	userIdString := strconv.Itoa(userId)
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendUsersChangeEmail, users.ChangeEmailRequest{
		UserId:   userIdString,
		NewEmail: newEmail,
	})
	if err != nil {
		return err
	}
	return nil
}
