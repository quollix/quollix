package src

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"strings"
	"time"

	"server/system_config_migrations"
	"server/tests/api_client"

	u "github.com/quollix/common/utils"
)

const initialAdminPasswordLogSearchText = "initial admin password"

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
	generatedPassword, err := waitForGeneratedInitialAdminPassword()
	if err != nil {
		return u.Logger.NewError("could not read generated initial admin password from container logs", "error", err.Error())
	}
	if generatedPassword == "password" {
		return u.Logger.NewError("generated password should not be 'password', but random generated", "actual", generatedPassword)
	}

	client := api_client.NewQuollixClient()
	if err := client.Auth.SignIn(system_config_migrations.DefaultInitialAdminName, generatedPassword); err != nil {
		return u.Logger.NewError("could not sign in with generated initial admin password", "error", err.Error())
	}

	currentUser, err := client.Auth.GetCurrentUser()
	if err != nil {
		return u.Logger.NewError("could not verify generated-password sign-in session", "error", err.Error())
	}
	if currentUser.Username != system_config_migrations.DefaultInitialAdminName || !currentUser.IsAdmin {
		return u.Logger.NewError("generated-password sign-in returned unexpected user", "username", currentUser.Username, "is_admin", currentUser.IsAdmin)
	}

	return nil
}

func waitForGeneratedInitialAdminPassword() (string, error) {
	var lastErr error
	for range 30 {
		logs, err := readQuollixContainerLogs()
		if err != nil {
			lastErr = err
		} else if password, err := extractGeneratedInitialAdminPassword(logs); err == nil {
			return password, nil
		} else {
			lastErr = err
		}
		time.Sleep(500 * time.Millisecond)
	}

	if lastErr == nil {
		lastErr = u.Logger.NewError("password log line was not found")
	}
	return "", lastErr
}

func readQuollixContainerLogs() (string, error) {
	output, err := exec.Command("docker", "logs", BrandAppContainerName).CombinedOutput() // #nosec G204: ci-runner reads logs from its fixed local test container
	if err != nil {
		return "", u.Logger.NewError("docker logs failed", "error", err.Error())
	}
	return string(output), nil
}

func extractGeneratedInitialAdminPassword(logs string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(logs))
	for scanner.Scan() {
		if password, found := extractGeneratedInitialAdminPasswordFromLine(scanner.Text()); found {
			return password, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", u.Logger.NewError(err.Error())
	}
	return "", u.Logger.NewError("password log line was not found")
}

func extractGeneratedInitialAdminPasswordFromLine(line string) (string, bool) {
	var fields map[string]any
	if err := json.Unmarshal([]byte(line), &fields); err != nil {
		return "", false
	}

	message := jsonStringField(fields, "message")
	if message == "" {
		message = jsonStringField(fields, "msg")
	}
	if !strings.Contains(message, initialAdminPasswordLogSearchText) {
		return "", false
	}

	password := jsonStringField(fields, "password")
	if password == "" {
		return "", false
	}
	return password, true
}

func jsonStringField(fields map[string]any, key string) string {
	value, ok := fields[key].(string)
	if !ok {
		return ""
	}
	return value
}
