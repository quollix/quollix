package src

import (
	"github.com/quollix/common/bootstrap"
	"github.com/quollix/common/quollix/api_client"
	u "github.com/quollix/common/utils"
)

func TestInitialAdminPassword() {
	Tr.Log.TaskDescription("Testing generated initial admin password")
	defer Tr.Cleanup()

	if err := testInitialAdminPassword(); err != nil {
		Tr.Log.Error("Generated initial admin password test failed: %v", err)
		Tr.ExitWithError()
		return
	}

	Tr.Log.Info("Generated initial admin password login succeeded")
}

func testInitialAdminPassword() error {
	DeployLocalContainer(false, containerEnv(false, true))
	username, password, err := bootstrap.WaitForGeneratedInitialAdminCredentials(BrandAppContainerName)
	if err != nil {
		return u.Logger.NewError("could not read generated initial admin password from container logs", "error", err.Error())
	}
	if password == "password" {
		return u.Logger.NewError("generated password should not be 'password', but random generated", "actual", password)
	}

	client := api_client.NewQuollixClient()
	if err := client.Auth.SignIn(username, password); err != nil {
		return u.Logger.NewError("could not sign in with generated initial admin password", "error", err.Error())
	}

	currentUser, err := client.Auth.GetCurrentUser()
	if err != nil {
		return u.Logger.NewError("could not verify generated-password sign-in session", "error", err.Error())
	}
	if currentUser.Username != username || !currentUser.IsAdmin {
		return u.Logger.NewError("generated-password sign-in returned unexpected user", "username", currentUser.Username, "is_admin", currentUser.IsAdmin)
	}

	return nil
}
