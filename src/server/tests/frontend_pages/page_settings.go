package frontend_pages

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

func (s *SettingsPage) AssertBaseDomainValue(expected string) *SettingsPage {
	err := tools.Eventually(func() error {
		if actual := s.Frame.Controls.GetInputValue("#base-domain"); actual != expected {
			return fmt.Errorf("unexpected base domain value: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.T, err)
	return s
}

func (s *SettingsPage) ChangeBaseDomain(value string) *SettingsPage {
	s.Frame.Controls.SetInputValue("#base-domain", value)
	return s
}

func (s *SettingsPage) SaveBaseDomainAndAssertSuccessSnackbar() *SettingsPage {
	s.Frame.Controls.GetRequiredElement("#settings-base-domain-save-button").MustClick()
	s.Frame.Browser.ConfirmDialog()
	s.Frame.Assert.SnackbarVisibleWithTextEventually("Base domain saved successfully.")
	s.Frame.Assert.PagePath(tools.Paths.FrontendSettings)
	return s
}

func (s *SettingsPage) StartDnsChallenge() *SettingsPage {
	s.Frame.Controls.GetRequiredElement("#settings-dns01-start-button").MustClick()
	s.Frame.Browser.ConfirmDialog()
	return s
}

func (s *SettingsPage) ResetCertificateAndAssertSuccessSnackbar() *SettingsPage {
	s.Frame.Controls.GetRequiredElement("#settings-certificate-reset-button").MustClick()
	s.Frame.Browser.ConfirmDialog()
	s.Frame.Assert.SnackbarVisibleWithTextEventually("Certificate has been reset to a self-signed certificate.")
	s.Frame.Assert.PagePath(tools.Paths.FrontendSettings)
	return s
}

func (s *SettingsPage) AssertCertificateOperationText(expected string) *SettingsPage {
	err := tools.Eventually(func() error {
		indicator := s.Frame.Controls.GetRequiredElement("#certificate-operation-indicator")
		if !isElementVisible(s.Frame.T, indicator) {
			return fmt.Errorf("certificate operation indicator is not visible yet")
		}

		textElement := s.Frame.Controls.GetRequiredElement("#certificate-operation-text")
		text, err := textElement.Text()
		assert.Nil(s.Frame.T, err)
		if actual := strings.TrimSpace(text); actual != expected {
			return fmt.Errorf("unexpected certificate operation text: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.T, err)
	return s
}

func (s *SettingsPage) AssertDnsChallengeResult(recordName, wildcardKeyAuth string) *SettingsPage {
	err := tools.Eventually(func() error {
		result := s.Frame.Controls.GetRequiredElement("#dns01ChallengeResult")
		if !isElementVisible(s.Frame.T, result) {
			return fmt.Errorf("dns challenge result is not visible yet")
		}
		if actual := s.Frame.Controls.GetInputValue("#dns01ChallengeRecordName"); actual != recordName {
			return fmt.Errorf("unexpected dns challenge record name: %q", actual)
		}
		if actual := s.Frame.Controls.GetInputValue("#dns01ChallengeRecordValue"); actual != wildcardKeyAuth {
			return fmt.Errorf("unexpected dns challenge record value: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.T, err)
	return s
}

func (s *SettingsPage) EnterBackupServerHostAndPort(host, port string) *SettingsPage {
	s.Frame.Controls.SetInputValue(`input[name="backupServerHost"]`, host)
	s.Frame.Controls.SetInputValue(`input[name="backupServerSshPort"]`, port)
	return s
}

func (s *SettingsPage) AssertBackupServerKnownHostsValue(expected string) *SettingsPage {
	err := tools.Eventually(func() error {
		if actual := s.Frame.Controls.GetInputValue("#backup-server-known-hosts"); actual != expected {
			return fmt.Errorf("unexpected backup server known hosts value: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.T, err)
	return s
}

func (s *SettingsPage) GetBackupServerKnownHosts() string {
	return s.Frame.Controls.GetInputValue("#backup-server-known-hosts")
}

func (s *SettingsPage) GetKnownHostsAndAssertValid() *SettingsPage {
	s.Frame.Controls.GetRequiredElement("#backup-server-get-known-hosts-button").MustClick()

	err := tools.EventuallyWithTimeout(backupOperationTimeout, 50*time.Millisecond, func() error {
		actual := s.Frame.Controls.GetInputValue("#backup-server-known-hosts")
		if actual == "" {
			return fmt.Errorf("backup server known hosts value is still empty")
		}
		return validation.Validate("SshKnownHosts", validation.FieldKnownHosts, actual)
	})
	assert.Nil(s.Frame.T, err)
	return s
}

func (s *SettingsPage) EnterBackupServerSshUserAndPassword(user, password string) *SettingsPage {
	s.Frame.Controls.SetInputValue("#backup-server-ssh-user", user)
	s.Frame.Controls.SetInputValue("#backupServerSshPassword", password)
	return s
}

func (s *SettingsPage) EnterBackupServerSshPassword(password string) *SettingsPage {
	s.Frame.Controls.SetInputValue("#backupServerSshPassword", password)
	return s
}

func (s *SettingsPage) TestBackupServerConnectionAndAssertSuccess() *SettingsPage {
	s.Frame.Controls.GetRequiredElement("#backup-server-test-connection-button").MustClick()
	s.Frame.Assert.SnackbarVisibleWithTextEventuallyWithin("Connection test successful.", backupOperationTimeout)
	return s
}

func (s *SettingsPage) EnterBackupServerEncryptionPassword(password string) *SettingsPage {
	s.Frame.Controls.SetInputValue("#backupServerEncryptionPassword", password)
	return s
}

func (s *SettingsPage) ToggleBackupServerSshPasswordVisibility() *SettingsPage {
	s.Frame.Controls.GetRequiredElement("#backupServerSshPasswordToggle").MustClick()
	return s
}

func (s *SettingsPage) ToggleBackupServerEncryptionPasswordVisibility() *SettingsPage {
	s.Frame.Controls.GetRequiredElement("#backupServerEncryptionPasswordToggle").MustClick()
	return s
}

func (s *SettingsPage) AssertBackupServerSshPasswordVisibility(visible bool) *SettingsPage {
	expectedType := "password"
	if visible {
		expectedType = "text"
	}

	err := tools.Eventually(func() error {
		if actual := s.Frame.Controls.GetInputType("#backupServerSshPassword"); actual != expectedType {
			return fmt.Errorf("unexpected backup server ssh password input type: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.T, err)
	return s
}

func (s *SettingsPage) AssertBackupServerEncryptionPasswordVisibility(visible bool) *SettingsPage {
	expectedType := "password"
	if visible {
		expectedType = "text"
	}

	err := tools.Eventually(func() error {
		if actual := s.Frame.Controls.GetInputType("#backupServerEncryptionPassword"); actual != expectedType {
			return fmt.Errorf("unexpected backup server encryption password input type: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.T, err)
	return s
}

func (s *SettingsPage) AssertBackupServerSshPasswordValue(expected string) *SettingsPage {
	err := tools.Eventually(func() error {
		if actual := s.Frame.Controls.GetInputValue("#backupServerSshPassword"); actual != expected {
			return fmt.Errorf("unexpected backup server ssh password value: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.T, err)
	return s
}

func (s *SettingsPage) AssertBackupServerEncryptionPasswordValue(expected string) *SettingsPage {
	err := tools.Eventually(func() error {
		if actual := s.Frame.Controls.GetInputValue("#backupServerEncryptionPassword"); actual != expected {
			return fmt.Errorf("unexpected backup server encryption password value: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.T, err)
	return s
}

func (s *SettingsPage) SetBackupServerEnabled(enabled bool) *SettingsPage {
	s.Frame.Controls.SetCheckboxValue("#backup-server-enabled-checkbox", enabled)
	return s
}

func (s *SettingsPage) SaveBackupServerAndAssertSuccessSnackbar() *SettingsPage {
	s.Frame.Controls.GetRequiredElement("#backup-server-save-button").MustClick()
	s.Frame.Assert.SnackbarVisibleWithTextEventuallyWithin("Backup server settings saved.", backupOperationTimeout)
	s.Frame.Assert.PagePath(tools.Paths.FrontendSettings)
	return s
}

func (s *SettingsPage) SaveBackupServerAndAssertErrorSnackbar(expected string) *SettingsPage {
	s.Frame.Controls.GetRequiredElement("#backup-server-save-button").MustClick()
	s.Frame.Assert.SnackbarVisibleWithTextEventuallyWithin(expected, backupOperationTimeout)
	s.Frame.Assert.PagePath(tools.Paths.FrontendSettings)
	return s
}

func (s *SettingsPage) ResetBackupServerAndAssertSuccessSnackbar() *SettingsPage {
	s.Frame.Controls.GetRequiredElement("#backup-server-reset-button").MustClick()
	s.Frame.Browser.ConfirmDialog()
	s.Frame.Assert.SnackbarVisibleWithTextEventually("Backup server settings have been reset.")
	s.Frame.Assert.PagePath(tools.Paths.FrontendSettings)
	return s
}

func (s *SettingsPage) WaitUntilBackupServerConfigMatches(expected *tools.BackupServerConfigs) *SettingsPage {
	err := tools.Eventually(func() error {
		configs, readErr := s.Frame.Client.Settings.ReadSshConfigs()
		assert.Nil(s.Frame.T, readErr)
		if *configs != *expected {
			return fmt.Errorf("backup server config does not match expected yet")
		}
		return nil
	})
	assert.Nil(s.Frame.T, err)
	return s
}

func (s *SettingsPage) PurgeBackupServerAndAssertSuccessSnackbar() *SettingsPage {
	s.Frame.Controls.GetRequiredElement("#backup-server-purge-button").MustClick()
	s.Frame.Browser.ConfirmDialog()
	s.Frame.Assert.SnackbarVisibleWithTextEventuallyWithin("Backup server has been purged.", backupOperationTimeout)
	s.Frame.Assert.PagePath(tools.Paths.FrontendSettings)
	return s
}

func (s *SettingsPage) ReadBackupServerFormValues() *tools.BackupServerConfigs {
	return &tools.BackupServerConfigs{
		IsEnabled:          s.Frame.Controls.GetCheckboxValue("#backup-server-enabled-checkbox"),
		Host:               s.Frame.Controls.GetInputValue(`input[name="backupServerHost"]`),
		SshPort:            s.Frame.Controls.GetInputValue(`input[name="backupServerSshPort"]`),
		SshUser:            s.Frame.Controls.GetInputValue("#backup-server-ssh-user"),
		SshPassword:        s.Frame.Controls.GetInputValue("#backupServerSshPassword"),
		SshKnownHosts:      s.Frame.Controls.GetInputValue("#backup-server-known-hosts"),
		EncryptionPassword: s.Frame.Controls.GetInputValue("#backupServerEncryptionPassword"),
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
	assert.Nil(s.Frame.T, err)
	return s
}

func (s *SettingsPage) EnterMaintenanceTimezone(value string) *SettingsPage {
	s.Frame.Controls.SetInputValue("#iana-timezone", value)
	return s
}

func (s *SettingsPage) AssertMaintenanceTimezoneValue(expected string) *SettingsPage {
	err := tools.Eventually(func() error {
		if actual := s.Frame.Controls.GetInputValue("#iana-timezone"); actual != expected {
			return fmt.Errorf("unexpected maintenance timezone: %q", actual)
		}
		return nil
	})
	assert.Nil(s.Frame.T, err)
	return s
}

func (s *SettingsPage) AssertMaintenanceTimezoneOptionPresent(expected string) *SettingsPage {
	options, err := s.Frame.Controls.GetRequiredElement("#ianaTimezoneOptions").Elements("option")
	assert.Nil(s.Frame.T, err)
	for _, option := range options {
		value, valueErr := option.Attribute("value")
		assert.Nil(s.Frame.T, valueErr)
		if value != nil && *value == expected {
			return s
		}
	}
	assert.True(s.Frame.T, false)
	return s
}

func (s *SettingsPage) AssertMaintenanceWindowOptionCount(expected int) *SettingsPage {
	options, err := s.Frame.Controls.GetRequiredElement("#maintenance-window-start-hour").Elements("option")
	assert.Nil(s.Frame.T, err)
	assert.Equal(s.Frame.T, expected, len(options))
	return s
}

func (s *SettingsPage) SelectMaintenanceWindow(label string) *SettingsPage {
	s.Frame.Controls.GetRequiredElement("#maintenance-window-start-hour").MustSelect(label)
	return s
}

func (s *SettingsPage) AssertSelectedMaintenanceWindow(label string) *SettingsPage {
	err := tools.Eventually(func() error {
		selectedOption, selectErr := s.Frame.Controls.GetRequiredElement("#maintenance-window-start-hour").Element("option:checked")
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
	assert.Nil(s.Frame.T, err)
	return s
}

func (s *SettingsPage) SaveMaintenanceConfigAndAssertSuccessSnackbar() *SettingsPage {
	s.Frame.Controls.GetRequiredElement("#maintenance-save-button").MustClick()
	s.Frame.Assert.SnackbarVisibleWithTextEventually("Maintenance settings saved.")
	s.Frame.Assert.PagePath(tools.Paths.FrontendSettings)
	return s
}

func (s *SettingsPage) AssertNextMaintenanceExecutionHour(expectedHour int) *SettingsPage {
	err := tools.Eventually(func() error {
		text, textErr := s.Frame.Controls.GetRequiredElement("#next-maintenance-execution-value").Text()
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
	assert.Nil(s.Frame.T, err)
	return s
}

func (s *SettingsPage) EnterRetentionPolicyValues(policy *retention.RetentionPolicy) *SettingsPage {
	s.Frame.Controls.SetInputValue("#retention-keep-pre-update", fmt.Sprintf("%d", policy.KeepPreUpdate))
	s.Frame.Controls.SetInputValue("#retention-keep-daily", fmt.Sprintf("%d", policy.KeepDaily))
	s.Frame.Controls.SetInputValue("#retention-keep-weekly", fmt.Sprintf("%d", policy.KeepWeekly))
	s.Frame.Controls.SetInputValue("#retention-keep-monthly", fmt.Sprintf("%d", policy.KeepMonthly))
	s.Frame.Controls.SetInputValue("#retention-keep-yearly", fmt.Sprintf("%d", policy.KeepYearly))
	return s
}

func (s *SettingsPage) ReadRetentionPolicyValues() *retention.RetentionPolicy {
	return &retention.RetentionPolicy{
		KeepPreUpdate: parseInt(s.Frame.T, s.Frame.Controls.GetInputValue("#retention-keep-pre-update")),
		KeepDaily:     parseInt(s.Frame.T, s.Frame.Controls.GetInputValue("#retention-keep-daily")),
		KeepWeekly:    parseInt(s.Frame.T, s.Frame.Controls.GetInputValue("#retention-keep-weekly")),
		KeepMonthly:   parseInt(s.Frame.T, s.Frame.Controls.GetInputValue("#retention-keep-monthly")),
		KeepYearly:    parseInt(s.Frame.T, s.Frame.Controls.GetInputValue("#retention-keep-yearly")),
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
	assert.Nil(s.Frame.T, err)
	return s
}

func (s *SettingsPage) SaveRetentionPolicyAndAssertSuccessSnackbar() *SettingsPage {
	s.Frame.Controls.GetRequiredElement("#retention-policy-save-button").MustClick()
	s.Frame.Browser.ConfirmDialog()
	s.Frame.Assert.SnackbarVisibleWithTextEventually("Retention policy saved.")
	s.Frame.Assert.PagePath(tools.Paths.FrontendSettings)
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
