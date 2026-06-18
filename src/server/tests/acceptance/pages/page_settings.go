//go:build acceptance

package pages

import (
	"fmt"
	"server/maintenance/retention"
	"server/tools"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
	"github.com/quollix/common/validation"
)

type SettingsPage struct {
	Frame *FrameType
}

func (s *SettingsPage) AssertHostValue(expected string) *SettingsPage {
	err := tools.Eventually(func() error {
		if actual := GetInputValue(s.Frame.t, s.Frame.page, "#host"); actual != expected {
			return fmt.Errorf("unexpected host value: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.t, err)
	return s
}

func (s *SettingsPage) ChangeHost(value string) *SettingsPage {
	SetInputValue(s.Frame.t, s.Frame.page, "#host", value)
	return s
}

func (s *SettingsPage) SaveHostAndAssertSuccessSnackbar() *SettingsPage {
	GetRequiredElement(s.Frame.t, s.Frame.page, "#settings-host-save-button").MustClick()
	s.Frame.ConfirmDialog()
	s.Frame.AssertSnackbarVisibleWithTextEventually("Host saved successfully.")
	s.Frame.AssertPagePath(tools.Paths.FrontendSettings)
	return s
}

func (s *SettingsPage) StartDnsChallenge() *SettingsPage {
	GetRequiredElement(s.Frame.t, s.Frame.page, "#settings-dns01-start-button").MustClick()
	s.Frame.ConfirmDialog()
	return s
}

func (s *SettingsPage) ResetCertificateAndAssertSuccessSnackbar() *SettingsPage {
	GetRequiredElement(s.Frame.t, s.Frame.page, "#settings-certificate-reset-button").MustClick()
	s.Frame.ConfirmDialog()
	s.Frame.AssertSnackbarVisibleWithTextEventually("Certificate has been reset to a self-signed certificate.")
	s.Frame.AssertPagePath(tools.Paths.FrontendSettings)
	return s
}

func (s *SettingsPage) AssertCertificateOperationText(expected string) *SettingsPage {
	err := tools.Eventually(func() error {
		indicator := GetRequiredElement(s.Frame.t, s.Frame.page, "#certificate-operation-indicator")
		if !isElementVisible(s.Frame.t, indicator) {
			return fmt.Errorf("certificate operation indicator is not visible yet")
		}

		textElement := GetRequiredElement(s.Frame.t, s.Frame.page, "#certificate-operation-text")
		text, err := textElement.Text()
		assert.Nil(s.Frame.t, err)
		if actual := strings.TrimSpace(text); actual != expected {
			return fmt.Errorf("unexpected certificate operation text: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.t, err)
	return s
}

func (s *SettingsPage) AssertDnsChallengeResult(recordName, wildcardKeyAuth string) *SettingsPage {
	err := tools.Eventually(func() error {
		result := GetRequiredElement(s.Frame.t, s.Frame.page, "#dns01ChallengeResult")
		if !isElementVisible(s.Frame.t, result) {
			return fmt.Errorf("dns challenge result is not visible yet")
		}
		if actual := GetInputValue(s.Frame.t, s.Frame.page, "#dns01ChallengeRecordName"); actual != recordName {
			return fmt.Errorf("unexpected dns challenge record name: %q", actual)
		}
		if actual := GetInputValue(s.Frame.t, s.Frame.page, "#dns01ChallengeRecordValue"); actual != wildcardKeyAuth {
			return fmt.Errorf("unexpected dns challenge record value: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.t, err)
	return s
}

func (s *SettingsPage) EnterBackupServerHostAndPort(host, port string) *SettingsPage {
	SetInputValue(s.Frame.t, s.Frame.page, `input[name="backupServerHost"]`, host)
	SetInputValue(s.Frame.t, s.Frame.page, `input[name="backupServerSshPort"]`, port)
	return s
}

func (s *SettingsPage) AssertBackupServerKnownHostsValue(expected string) *SettingsPage {
	err := tools.Eventually(func() error {
		if actual := GetInputValue(s.Frame.t, s.Frame.page, "#backup-server-known-hosts"); actual != expected {
			return fmt.Errorf("unexpected backup server known hosts value: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.t, err)
	return s
}

func (s *SettingsPage) GetBackupServerKnownHosts() string {
	return GetInputValue(s.Frame.t, s.Frame.page, "#backup-server-known-hosts")
}

func (s *SettingsPage) GetKnownHostsAndAssertValid() *SettingsPage {
	GetRequiredElement(s.Frame.t, s.Frame.page, "#backup-server-get-known-hosts-button").MustClick()

	err := tools.EventuallyWithTimeout(backupOperationTimeout, 50*time.Millisecond, func() error {
		actual := GetInputValue(s.Frame.t, s.Frame.page, "#backup-server-known-hosts")
		if actual == "" {
			return fmt.Errorf("backup server known hosts value is still empty")
		}
		return validation.Validate("SshKnownHosts", validation.FieldKnownHosts, actual)
	})
	assert.Nil(s.Frame.t, err)
	return s
}

func (s *SettingsPage) EnterBackupServerSshUserAndPassword(user, password string) *SettingsPage {
	SetInputValue(s.Frame.t, s.Frame.page, "#backup-server-ssh-user", user)
	SetInputValue(s.Frame.t, s.Frame.page, "#backupServerSshPassword", password)
	return s
}

func (s *SettingsPage) EnterBackupServerSshPassword(password string) *SettingsPage {
	SetInputValue(s.Frame.t, s.Frame.page, "#backupServerSshPassword", password)
	return s
}

func (s *SettingsPage) TestBackupServerConnectionAndAssertSuccess() *SettingsPage {
	GetRequiredElement(s.Frame.t, s.Frame.page, "#backup-server-test-connection-button").MustClick()
	s.Frame.AssertSnackbarVisibleWithTextEventuallyWithin("Connection test successful.", backupOperationTimeout)
	return s
}

func (s *SettingsPage) EnterBackupServerEncryptionPassword(password string) *SettingsPage {
	SetInputValue(s.Frame.t, s.Frame.page, "#backupServerEncryptionPassword", password)
	return s
}

func (s *SettingsPage) ToggleBackupServerSshPasswordVisibility() *SettingsPage {
	GetRequiredElement(s.Frame.t, s.Frame.page, "#backupServerSshPasswordToggle").MustClick()
	return s
}

func (s *SettingsPage) ToggleBackupServerEncryptionPasswordVisibility() *SettingsPage {
	GetRequiredElement(s.Frame.t, s.Frame.page, "#backupServerEncryptionPasswordToggle").MustClick()
	return s
}

func (s *SettingsPage) AssertBackupServerSshPasswordVisibility(visible bool) *SettingsPage {
	expectedType := "password"
	if visible {
		expectedType = "text"
	}

	err := tools.Eventually(func() error {
		if actual := GetInputType(s.Frame.t, s.Frame.page, "#backupServerSshPassword"); actual != expectedType {
			return fmt.Errorf("unexpected backup server ssh password input type: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.t, err)
	return s
}

func (s *SettingsPage) AssertBackupServerEncryptionPasswordVisibility(visible bool) *SettingsPage {
	expectedType := "password"
	if visible {
		expectedType = "text"
	}

	err := tools.Eventually(func() error {
		if actual := GetInputType(s.Frame.t, s.Frame.page, "#backupServerEncryptionPassword"); actual != expectedType {
			return fmt.Errorf("unexpected backup server encryption password input type: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.t, err)
	return s
}

func (s *SettingsPage) AssertBackupServerSshPasswordValue(expected string) *SettingsPage {
	err := tools.Eventually(func() error {
		if actual := GetInputValue(s.Frame.t, s.Frame.page, "#backupServerSshPassword"); actual != expected {
			return fmt.Errorf("unexpected backup server ssh password value: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.t, err)
	return s
}

func (s *SettingsPage) AssertBackupServerEncryptionPasswordValue(expected string) *SettingsPage {
	err := tools.Eventually(func() error {
		if actual := GetInputValue(s.Frame.t, s.Frame.page, "#backupServerEncryptionPassword"); actual != expected {
			return fmt.Errorf("unexpected backup server encryption password value: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.t, err)
	return s
}

func (s *SettingsPage) SetBackupServerEnabled(enabled bool) *SettingsPage {
	SetCheckboxValue(s.Frame.t, s.Frame.page, "#backup-server-enabled-checkbox", enabled)
	return s
}

func (s *SettingsPage) SaveBackupServerAndAssertSuccessSnackbar() *SettingsPage {
	GetRequiredElement(s.Frame.t, s.Frame.page, "#backup-server-save-button").MustClick()
	s.Frame.AssertSnackbarVisibleWithTextEventuallyWithin("Backup server settings saved.", backupOperationTimeout)
	s.Frame.AssertPagePath(tools.Paths.FrontendSettings)
	return s
}

func (s *SettingsPage) SaveBackupServerAndAssertErrorSnackbar(expected string) *SettingsPage {
	GetRequiredElement(s.Frame.t, s.Frame.page, "#backup-server-save-button").MustClick()
	s.Frame.AssertSnackbarVisibleWithTextEventuallyWithin(expected, backupOperationTimeout)
	s.Frame.AssertPagePath(tools.Paths.FrontendSettings)
	return s
}

func (s *SettingsPage) ResetBackupServerAndAssertSuccessSnackbar() *SettingsPage {
	GetRequiredElement(s.Frame.t, s.Frame.page, "#backup-server-reset-button").MustClick()
	s.Frame.ConfirmDialog()
	s.Frame.AssertSnackbarVisibleWithTextEventually("Backup server settings have been reset.")
	s.Frame.AssertPagePath(tools.Paths.FrontendSettings)
	return s
}

func (s *SettingsPage) WaitUntilBackupServerConfigMatches(expected *tools.BackupServerConfigs) *SettingsPage {
	err := tools.Eventually(func() error {
		configs, readErr := s.Frame.Client.Settings.ReadSshConfigs()
		if readErr != nil {
			return readErr
		}
		if *configs != *expected {
			return fmt.Errorf("backup server config does not match expected yet")
		}
		return nil
	})
	assert.Nil(s.Frame.t, err)
	return s
}

func (s *SettingsPage) PurgeBackupServerAndAssertSuccessSnackbar() *SettingsPage {
	GetRequiredElement(s.Frame.t, s.Frame.page, "#backup-server-purge-button").MustClick()
	s.Frame.ConfirmDialog()
	s.Frame.AssertSnackbarVisibleWithTextEventuallyWithin("Backup server has been purged.", backupOperationTimeout)
	s.Frame.AssertPagePath(tools.Paths.FrontendSettings)
	return s
}

func (s *SettingsPage) ReadBackupServerFormValues() *tools.BackupServerConfigs {
	return &tools.BackupServerConfigs{
		IsEnabled:          GetCheckboxValue(s.Frame.t, s.Frame.page, "#backup-server-enabled-checkbox"),
		Host:               GetInputValue(s.Frame.t, s.Frame.page, `input[name="backupServerHost"]`),
		SshPort:            GetInputValue(s.Frame.t, s.Frame.page, `input[name="backupServerSshPort"]`),
		SshUser:            GetInputValue(s.Frame.t, s.Frame.page, "#backup-server-ssh-user"),
		SshPassword:        GetInputValue(s.Frame.t, s.Frame.page, "#backupServerSshPassword"),
		SshKnownHosts:      GetInputValue(s.Frame.t, s.Frame.page, "#backup-server-known-hosts"),
		EncryptionPassword: GetInputValue(s.Frame.t, s.Frame.page, "#backupServerEncryptionPassword"),
	}
}

func (s *SettingsPage) AssertBackupServerFormValues(expected *tools.BackupServerConfigs) *SettingsPage {
	err := tools.Eventually(func() error {
		actual := s.ReadBackupServerFormValues()
		if *actual != *expected {
			return fmt.Errorf("unexpected backup server form values")
		}
		return nil
	})
	assert.Nil(s.Frame.t, err)
	return s
}

func (s *SettingsPage) EnterMaintenanceTimezone(value string) *SettingsPage {
	SetInputValue(s.Frame.t, s.Frame.page, "#iana-timezone", value)
	return s
}

func (s *SettingsPage) AssertMaintenanceTimezoneValue(expected string) *SettingsPage {
	err := tools.Eventually(func() error {
		if actual := GetInputValue(s.Frame.t, s.Frame.page, "#iana-timezone"); actual != expected {
			return fmt.Errorf("unexpected maintenance timezone: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.t, err)
	return s
}

func (s *SettingsPage) AssertMaintenanceTimezoneOptionPresent(expected string) *SettingsPage {
	options, err := GetRequiredElement(s.Frame.t, s.Frame.page, "#ianaTimezoneOptions").Elements("option")
	assert.Nil(s.Frame.t, err)
	for _, option := range options {
		value, valueErr := option.Attribute("value")
		assert.Nil(s.Frame.t, valueErr)
		if value != nil && *value == expected {
			return s
		}
	}
	assert.True(s.Frame.t, false)
	return s
}

func (s *SettingsPage) AssertMaintenanceWindowOptionCount(expected int) *SettingsPage {
	options, err := GetRequiredElement(s.Frame.t, s.Frame.page, "#maintenance-window-start-hour").Elements("option")
	assert.Nil(s.Frame.t, err)
	assert.Equal(s.Frame.t, expected, len(options))
	return s
}

func (s *SettingsPage) SelectMaintenanceWindow(label string) *SettingsPage {
	GetRequiredElement(s.Frame.t, s.Frame.page, "#maintenance-window-start-hour").MustSelect(label)
	return s
}

func (s *SettingsPage) AssertSelectedMaintenanceWindow(label string) *SettingsPage {
	err := tools.Eventually(func() error {
		selectedOption, selectErr := GetRequiredElement(s.Frame.t, s.Frame.page, "#maintenance-window-start-hour").Element("option:checked")
		if selectErr != nil {
			return selectErr
		}
		text, textErr := selectedOption.Text()
		if textErr != nil {
			return textErr
		}
		text = strings.TrimSpace(text)
		if text != label {
			return fmt.Errorf("unexpected selected maintenance window: %q", text)
		}
		return nil
	})
	assert.Nil(s.Frame.t, err)
	return s
}

func (s *SettingsPage) SaveMaintenanceConfigAndAssertSuccessSnackbar() *SettingsPage {
	GetRequiredElement(s.Frame.t, s.Frame.page, "#maintenance-save-button").MustClick()
	s.Frame.AssertSnackbarVisibleWithTextEventually("Maintenance settings saved.")
	s.Frame.AssertPagePath(tools.Paths.FrontendSettings)
	return s
}

func (s *SettingsPage) AssertNextMaintenanceExecutionHour(expectedHour int) *SettingsPage {
	err := tools.Eventually(func() error {
		text, textErr := GetRequiredElement(s.Frame.t, s.Frame.page, "#next-maintenance-execution-value").Text()
		if textErr != nil {
			return textErr
		}
		nextMaintenanceAt, parseErr := time.Parse("Mon, 02 Jan 2006 15:04", text)
		if parseErr != nil {
			return parseErr
		}
		if nextMaintenanceAt.Hour() != expectedHour {
			return fmt.Errorf("unexpected next maintenance execution hour: %d", nextMaintenanceAt.Hour())
		}
		return nil
	})
	assert.Nil(s.Frame.t, err)
	return s
}

func (s *SettingsPage) EnterRetentionPolicyValues(policy *retention.RetentionPolicy) *SettingsPage {
	SetInputValue(s.Frame.t, s.Frame.page, "#retention-keep-pre-update", fmt.Sprintf("%d", policy.KeepPreUpdate))
	SetInputValue(s.Frame.t, s.Frame.page, "#retention-keep-daily", fmt.Sprintf("%d", policy.KeepDaily))
	SetInputValue(s.Frame.t, s.Frame.page, "#retention-keep-weekly", fmt.Sprintf("%d", policy.KeepWeekly))
	SetInputValue(s.Frame.t, s.Frame.page, "#retention-keep-monthly", fmt.Sprintf("%d", policy.KeepMonthly))
	SetInputValue(s.Frame.t, s.Frame.page, "#retention-keep-yearly", fmt.Sprintf("%d", policy.KeepYearly))
	return s
}

func (s *SettingsPage) ReadRetentionPolicyValues() *retention.RetentionPolicy {
	return &retention.RetentionPolicy{
		KeepPreUpdate: parseInt(s.Frame.t, GetInputValue(s.Frame.t, s.Frame.page, "#retention-keep-pre-update")),
		KeepDaily:     parseInt(s.Frame.t, GetInputValue(s.Frame.t, s.Frame.page, "#retention-keep-daily")),
		KeepWeekly:    parseInt(s.Frame.t, GetInputValue(s.Frame.t, s.Frame.page, "#retention-keep-weekly")),
		KeepMonthly:   parseInt(s.Frame.t, GetInputValue(s.Frame.t, s.Frame.page, "#retention-keep-monthly")),
		KeepYearly:    parseInt(s.Frame.t, GetInputValue(s.Frame.t, s.Frame.page, "#retention-keep-yearly")),
	}
}

func (s *SettingsPage) AssertRetentionPolicyValues(expected *retention.RetentionPolicy) *SettingsPage {
	err := tools.Eventually(func() error {
		actual := s.ReadRetentionPolicyValues()
		if *actual != *expected {
			return fmt.Errorf("unexpected retention policy values: %+v", *actual)
		}
		return nil
	})
	assert.Nil(s.Frame.t, err)
	return s
}

func (s *SettingsPage) SaveRetentionPolicyAndAssertSuccessSnackbar() *SettingsPage {
	GetRequiredElement(s.Frame.t, s.Frame.page, "#retention-policy-save-button").MustClick()
	s.Frame.ConfirmDialog()
	s.Frame.AssertSnackbarVisibleWithTextEventually("Retention policy saved.")
	s.Frame.AssertPagePath(tools.Paths.FrontendSettings)
	return s
}

func parseInt(t *testing.T, value string) int {
	intValue, err := strconv.Atoi(value)
	assert.Nil(t, err)
	return intValue
}

func isElementVisible(t *testing.T, element *rod.Element) bool {
	style, err := element.Attribute("style")
	assert.Nil(t, err)
	if style == nil {
		return true
	}
	return !strings.Contains(strings.ToLower(strings.TrimSpace(*style)), "display:none") &&
		!strings.Contains(strings.ToLower(strings.TrimSpace(*style)), "display: none")
}
