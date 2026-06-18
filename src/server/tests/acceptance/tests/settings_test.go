//go:build acceptance

package acceptance

import (
	"server/backup_server"
	"server/certificates"
	"server/maintenance/retention"
	"server/tests/acceptance/pages"
	"server/tests/component"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/validation"
)

func TestSettingsHost(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.GoToSettingsPage().
		AssertHostValue("localhost").
		ChangeHost("localhost2").
		SaveHostAndAssertSuccessSnackbar()

	host, err := frame.Client.Settings.GetHostValue()
	assert.Nil(t, err)
	assert.Equal(t, "localhost2", host)
}

func TestSettingsCertificateDnsChallenge(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.GoToSettingsPage().
		StartDnsChallenge().
		AssertCertificateOperationText("finished sample wildcard certificate generation successfully").
		AssertDnsChallengeResult("_acme-challenge.localhost", certificates.SampleWildcardKeyAuth)
}

func TestSettingsCertificateReset(t *testing.T) {
	frame := pages.Setup(t)
	frame.Client.Certificates.Reset()
	defer frame.Client.Test.ResetTestState()

	beforeResetBundleBytes := frame.Client.Certificates.DownloadBundleBytes()
	beforeResetDownloadedLeafDerBytes := component.ExtractLeafCertificateDerBytesFromBundle(t, beforeResetBundleBytes)
	component.AssertServerUsesCertificateBundle(t, beforeResetBundleBytes)

	frame.GoToSettingsPage().
		ResetCertificateAndAssertSuccessSnackbar()

	afterResetBundleBytes := frame.Client.Certificates.DownloadBundleBytes()
	afterResetDownloadedLeafDerBytes := component.ExtractLeafCertificateDerBytesFromBundle(t, afterResetBundleBytes)
	component.AssertServerUsesCertificateBundle(t, afterResetBundleBytes)

	assert.NotEqual(t, beforeResetBundleBytes, afterResetBundleBytes)
	assert.NotEqual(t, beforeResetDownloadedLeafDerBytes, afterResetDownloadedLeafDerBytes)
}

func TestSettingsBackupServerFlow(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()
	expected := backup_server.GetSampleRemoteRepo()

	page := frame.GoToSettingsPage().
		AssertBackupServerFormValues(&tools.BackupServerConfigs{}).
		EnterBackupServerHostAndPort(expected.Host, expected.SshPort).
		AssertBackupServerKnownHostsValue("").
		GetKnownHostsAndAssertValid()

	expected.SshKnownHosts = page.GetBackupServerKnownHosts()
	err := validation.Validate("SshKnownHosts", validation.FieldKnownHosts, expected.SshKnownHosts)
	assert.Nil(t, err)

	page.
		EnterBackupServerSshUserAndPassword(expected.SshUser, expected.SshPassword).
		TestBackupServerConnectionAndAssertSuccess().
		EnterBackupServerEncryptionPassword(expected.EncryptionPassword).
		SetBackupServerEnabled(true).
		SaveBackupServerAndAssertSuccessSnackbar()

	frame.AssertAppOperationFinished()
	page.WaitUntilBackupServerConfigMatches(expected)

	frame.ReloadPage()
	page.AssertBackupServerFormValues(expected)

	emptyConfig := &tools.BackupServerConfigs{}
	frame.SettingsPage.
		ResetBackupServerAndAssertSuccessSnackbar().
		AssertBackupServerFormValues(emptyConfig)

	frame.ReloadPage()
	page.AssertBackupServerFormValues(emptyConfig)
}

func TestSettingsBackupServerPurgeFlow(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	expected := backup_server.GetSampleRemoteRepo()
	expected.SshKnownHosts = frame.Client.Settings.GetKnownHosts(expected)
	assert.Nil(t, frame.Client.Settings.SaveSshConfigs(expected))

	page := frame.GoToSettingsPage()
	page.AssertBackupServerFormValues(expected).
		EnterBackupServerEncryptionPassword("newpassword").
		SetBackupServerEnabled(true).
		SaveBackupServerAndAssertErrorSnackbar(backup_server.WrongEncryptionPasswordErr)

	_, err := frame.Client.Parent.DoRequest(tools.Paths.BackendSettingsSshConfigsReset, nil)
	assert.Nil(t, err)

	frame.ReloadPage()
	emptyConfig := &tools.BackupServerConfigs{}
	frame.SettingsPage.AssertBackupServerFormValues(emptyConfig)

	frame.SettingsPage.
		EnterBackupServerHostAndPort(expected.Host, expected.SshPort).
		AssertBackupServerKnownHostsValue("").
		GetKnownHostsAndAssertValid()

	expected.SshKnownHosts = frame.SettingsPage.GetBackupServerKnownHosts()
	err = validation.Validate("SshKnownHosts", validation.FieldKnownHosts, expected.SshKnownHosts)
	assert.Nil(t, err)

	page.EnterBackupServerSshUserAndPassword(expected.SshUser, expected.SshPassword).
		PurgeBackupServerAndAssertSuccessSnackbar().
		WaitUntilBackupServerConfigMatches(emptyConfig).
		EnterBackupServerEncryptionPassword("newpassword").
		SetBackupServerEnabled(true).
		SaveBackupServerAndAssertSuccessSnackbar()

	// Due to the known 'NETWORK_CHANGED' case, the success snackbar appears while the underlying request is still ongoing. This makes deferred cleanup run before that tail work has finished, the SSH settings is written back after cleanup and leave dirty state, making subsequent tests potentially fail. For this reason, we have to wait until the operation is finished.
	frame.AssertAppOperationFinished()

	savedConfig, err := frame.Client.Settings.ReadSshConfigs()
	assert.Nil(t, err)
	expected.EncryptionPassword = "newpassword"
	assert.Equal(t, *expected, *savedConfig)
}

func TestSettingsBackupServerPasswordVisibility(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.GoToSettingsPage().
		EnterBackupServerSshPassword("sshpassword").
		AssertBackupServerSshPasswordVisibility(false).
		AssertBackupServerSshPasswordValue("sshpassword").
		ToggleBackupServerSshPasswordVisibility().
		AssertBackupServerSshPasswordVisibility(true).
		AssertBackupServerSshPasswordValue("sshpassword").
		ToggleBackupServerSshPasswordVisibility().
		AssertBackupServerSshPasswordVisibility(false).
		AssertBackupServerSshPasswordValue("sshpassword").
		EnterBackupServerEncryptionPassword("restic-password").
		AssertBackupServerEncryptionPasswordVisibility(false).
		AssertBackupServerEncryptionPasswordValue("restic-password").
		ToggleBackupServerEncryptionPasswordVisibility().
		AssertBackupServerEncryptionPasswordVisibility(true).
		AssertBackupServerEncryptionPasswordValue("restic-password").
		ToggleBackupServerEncryptionPasswordVisibility().
		AssertBackupServerEncryptionPasswordVisibility(false).
		AssertBackupServerEncryptionPasswordValue("restic-password")
}

func TestSettingsMaintenanceScheduleFlow(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.GoToSettingsPage().
		EnterMaintenanceTimezone("Europe/").
		AssertMaintenanceTimezoneValue("Europe/").
		AssertMaintenanceTimezoneOptionPresent("Europe/London").
		AssertMaintenanceTimezoneOptionPresent("Europe/Berlin").
		EnterMaintenanceTimezone("Europe/London").
		AssertMaintenanceTimezoneValue("Europe/London").
		AssertMaintenanceWindowOptionCount(24).
		SelectMaintenanceWindow("10:00-11:00").
		AssertSelectedMaintenanceWindow("10:00-11:00").
		SaveMaintenanceConfigAndAssertSuccessSnackbar().
		AssertNextMaintenanceExecutionHour(10)

	frame.ReloadPage()
	frame.SettingsPage.
		AssertMaintenanceTimezoneValue("Europe/London").
		AssertSelectedMaintenanceWindow("10:00-11:00").
		AssertNextMaintenanceExecutionHour(10)
}

func TestSettingsRetentionPolicyFlow(t *testing.T) {
	frame := pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	expected := &retention.RetentionPolicy{
		KeepPreUpdate: 10,
		KeepDaily:     11,
		KeepWeekly:    12,
		KeepMonthly:   13,
		KeepYearly:    14,
	}

	frame.GoToSettingsPage().
		EnterRetentionPolicyValues(expected).
		SaveRetentionPolicyAndAssertSuccessSnackbar()

	frame.ReloadPage()
	frame.SettingsPage.AssertRetentionPolicyValues(expected)
}
