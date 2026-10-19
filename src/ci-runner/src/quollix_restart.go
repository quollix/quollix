package src

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/quollix/common/quollix/api_client"
	u "github.com/quollix/common/utils"
)

const restartTestAppURL = "https://sampleapp.localhost/"

func TestQuollixRestart() {
	Tr.Log.TaskDescription("Testing sampleapp access after recreating the Quollix container")
	defer Tr.Cleanup()

	if err := testQuollixRestart(); err != nil {
		Tr.Log.Error("Quollix restart test failed: %v", err)
		Tr.ExitWithError()
	}
}

func testQuollixRestart() error {
	envVars := containerEnv(true, false)
	DeployLocalContainer(false, envVars)

	client := api_client.NewQuollixClient()
	if err := client.Auth.SignIn(defaultAdminName, defaultAdminPassword); err != nil {
		return err
	}
	if err := installRestartTestApp(client); err != nil {
		return err
	}
	secret, err := client.AppAccess.GetSecret("sampleapp")
	if err != nil {
		return err
	}
	cookie, err := client.AppAccess.ExchangeSecretForAppAccessCookie(secret, restartTestAppURL)
	if err != nil {
		return err
	}
	if err := checkRestartTestAppAccess(client, cookie); err != nil {
		return err
	}

	Tr.Cmd().Run("docker rm -f %s", BrandAppContainerName)
	Tr.Cmd().Dir(ServerDir).Run("%s", getDockerCommand(envVars))
	Tr.WaitForWebPageToBeReady(localBrandAppUrl)
	return checkRestartTestAppAccess(client, cookie)
}

func installRestartTestApp(client *api_client.QuollixClient) error {
	versions, err := client.Apps.ListVersions("samplemaintainer", "sampleapp")
	if err != nil {
		return err
	}
	for _, version := range versions {
		if version.Name == "2.0" {
			return client.Apps.InstallFromStoreVersion(version.VersionId)
		}
	}
	return u.Logger.NewError("sampleapp version 2.0 was not found")
}

func checkRestartTestAppAccess(client *api_client.QuollixClient, cookie *http.Cookie) error {
	return u.EventuallyWithTimeout(10*time.Second, 250*time.Millisecond, func() error {
		response, err := client.AppAccess.DoAppRequest(cookie, http.MethodGet, restartTestAppURL, nil)
		if err != nil {
			return err
		}
		defer u.Close(response.Body)
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return u.Logger.NewError(err.Error())
		}
		if response.StatusCode != http.StatusOK || strings.TrimSpace(string(body)) != "this is version 2.0" {
			return u.Logger.NewError("unexpected sampleapp response", "status_code", response.StatusCode, "body", string(body))
		}
		return nil
	})
}
