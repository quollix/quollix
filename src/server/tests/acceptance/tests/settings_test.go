//go:build acceptance

package acceptance

import (
	"server/backup_server"
	"server/certificates"
	"server/maintenance/retention"
	"server/tests/component"
	"server/tests/frontend_pages"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/validation"
)

func TestSettingsBaseDomain(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.Pages.GoToSettingsPage().
		AssertBaseDomainValue("localhost").
		ChangeBaseDomain("localhost2").
		SaveBaseDomainAndAssertSuccessSnackbar()

	host, err := frame.Client.Settings.GetBaseDomainValue()
	assert.Nil(t, err)
	assert.Equal(t, "localhost2", host)
}

func TestSettingsCertificateDnsChallenge(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.Pages.GoToSettingsPage().
		StartDnsChallenge().
		AssertCertificateOperationText("finished sample wildcard certificate generation successfully").
		AssertDnsChallengeResult("_acme-challenge.localhost", certificates.SampleWildcardKeyAuth)
}

func TestSettingsCertificateReset(t *testing.T) {
	frame := frontend_pages.Setup(t)
	assert.Nil(t, frame.Client.Certificates.Reset())
	defer frame.Client.Test.ResetTestState()

	beforeResetBundleBytes, err := frame.Client.Certificates.DownloadBundleBytes()
	assert.Nil(t, err)
	beforeResetDownloadedLeafDerBytes := component.ExtractLeafCertificateDerBytesFromBundle(t, beforeResetBundleBytes)
	component.AssertServerUsesCertificateBundle(t, beforeResetBundleBytes)

	frame.Pages.GoToSettingsPage().
		ResetCertificateAndAssertSuccessSnackbar()

	afterResetBundleBytes, err := frame.Client.Certificates.DownloadBundleBytes()
	assert.Nil(t, err)
	afterResetDownloadedLeafDerBytes := component.ExtractLeafCertificateDerBytesFromBundle(t, afterResetBundleBytes)
	component.AssertServerUsesCertificateBundle(t, afterResetBundleBytes)

	assert.NotEqual(t, beforeResetBundleBytes, afterResetBundleBytes)
	assert.NotEqual(t, beforeResetDownloadedLeafDerBytes, afterResetDownloadedLeafDerBytes)
}

func TestSettingsBackupServerFlow(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()
	expected := backup_server.GetSampleRemoteRepo()

	page := frame.Pages.GoToSettingsPage().
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

	frame.Assert.AppOperationFinished()
	page.WaitUntilBackupServerConfigMatches(expected)

	frame.Browser.ReloadPage()
	page.AssertBackupServerFormValues(expected)

	emptyConfig := &tools.BackupServerConfigs{}
	frame.Pages.SettingsPage.
		ResetBackupServerAndAssertSuccessSnackbar(). // TODO acceptance test seems to fail here sometimes. To be investigated. Maybe: snackbars that come quite fast might not override the predecessor snackbar, which makes their success hidden. error trace confirms that snackbar seems to be the issue.
		AssertBackupServerFormValues(emptyConfig)

	frame.Browser.ReloadPage()
	page.AssertBackupServerFormValues(emptyConfig)
}

func TestSettingsBackupServerPurgeFlow(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	expected := backup_server.GetSampleRemoteRepo()
	var err error
	expected.SshKnownHosts, err = frame.Client.Settings.GetKnownHosts(expected)
	assert.Nil(t, err)
	assert.Nil(t, frame.Client.Settings.SaveSshConfigs(expected))

	page := frame.Pages.GoToSettingsPage()
	page.AssertBackupServerFormValues(expected).
		EnterBackupServerEncryptionPassword("newpassword").
		SetBackupServerEnabled(true).
		SaveBackupServerAndAssertErrorSnackbar(backup_server.WrongEncryptionPasswordErr)

	_, err = frame.Client.Parent.DoRequest(tools.Paths.BackendSettingsSshConfigsReset, nil)
	assert.Nil(t, err)

	frame.Browser.ReloadPage()
	emptyConfig := &tools.BackupServerConfigs{}
	frame.Pages.SettingsPage.AssertBackupServerFormValues(emptyConfig)

	frame.Pages.SettingsPage.
		EnterBackupServerHostAndPort(expected.Host, expected.SshPort).
		AssertBackupServerKnownHostsValue("").
		GetKnownHostsAndAssertValid()

	expected.SshKnownHosts = frame.Pages.SettingsPage.GetBackupServerKnownHosts()
	err = validation.Validate("SshKnownHosts", validation.FieldKnownHosts, expected.SshKnownHosts)
	assert.Nil(t, err)

	page.EnterBackupServerSshUserAndPassword(expected.SshUser, expected.SshPassword).
		PurgeBackupServerAndAssertSuccessSnackbar().
		WaitUntilBackupServerConfigMatches(emptyConfig).
		EnterBackupServerEncryptionPassword("newpassword").
		SetBackupServerEnabled(true).
		SaveBackupServerAndAssertSuccessSnackbar()

	// Due to the known 'NETWORK_CHANGED' case, the success snackbar appears while the underlying request is still ongoing. This makes deferred cleanup run before that tail work has finished, the SSH settings is written back after cleanup and leave dirty state, making subsequent tests potentially fail. For this reason, we have to wait until the operation is finished.
	frame.Assert.AppOperationFinished()

	savedConfig, err := frame.Client.Settings.ReadSshConfigs()
	assert.Nil(t, err)
	expected.EncryptionPassword = "newpassword"
	assert.Equal(t, *expected, *savedConfig)
}

func TestSettingsBackupServerPasswordVisibility(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.Pages.GoToSettingsPage().
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
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.Pages.GoToSettingsPage().
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

	frame.Browser.ReloadPage()
	frame.Pages.SettingsPage.
		AssertMaintenanceTimezoneValue("Europe/London").
		AssertSelectedMaintenanceWindow("10:00-11:00").
		AssertNextMaintenanceExecutionHour(10)
}

func TestSettingsRetentionPolicyFlow(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	expected := &retention.RetentionPolicy{
		KeepPreUpdate: 10,
		KeepDaily:     11,
		KeepWeekly:    12,
		KeepMonthly:   13,
		KeepYearly:    14,
	}

	frame.Pages.GoToSettingsPage().
		EnterRetentionPolicyValues(expected).
		SaveRetentionPolicyAndAssertSuccessSnackbar()

	frame.Browser.ReloadPage()
	frame.Pages.SettingsPage.AssertRetentionPolicyValues(expected)
}
