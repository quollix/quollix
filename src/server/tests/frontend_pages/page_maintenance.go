package frontend_pages

import (
	"server/tests/component"
	"server/tools"
	"strings"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

type MaintenancePage struct {
	Frame *FrameType
}

type MaintenanceAppEntry struct {
	Maintainer string
	AppName    string

	AutomaticUpdatesCheckboxPresent bool
	AutomaticUpdatesEnabled         bool

	AutomaticBackupsCheckboxPresent bool
	AutomaticBackupsEnabled         bool
}

func (m *MaintenancePage) ListApps() []MaintenanceAppEntry {
	rows := m.Frame.Page.MustElements("tr.maintenance-row")
	out := make([]MaintenanceAppEntry, 0, len(rows))

	for _, row := range rows {
		entry := m.readAppEntry(row)
		out = append(out, *entry)
	}
	return out
}

func (m *MaintenancePage) GetApp(appName string) *MaintenanceAppEntry {
	apps := m.ListApps()
	for _, app := range apps {
		if app.AppName == appName {
			appCopy := app
			return &appCopy
		}
	}
	return nil
}

func (m *MaintenancePage) GetRequiredApp(appName string) *MaintenanceAppEntry {
	app := m.GetApp(appName)
	assert.NotNil(m.Frame.T, app)
	return app
}

func (m *MaintenancePage) SetAutomaticUpdatesEnabled(appName string, enabled bool) *MaintenancePage {
	row, err := m.findRowByAppName(appName)
	assert.Nil(m.Frame.T, err)
	m.setCheckboxEnabled(row, ".auto-update-cell input[type='checkbox']", enabled)
	return m
}

func (m *MaintenancePage) SetAutomaticBackupsEnabled(appName string, enabled bool) *MaintenancePage {
	row, err := m.findRowByAppName(appName)
	assert.Nil(m.Frame.T, err)
	m.setCheckboxEnabled(row, ".auto-backup-cell input[type='checkbox']", enabled)
	return m
}

func (m *MaintenancePage) setCheckboxEnabled(row *rod.Element, selector string, expectedChecked bool) {
	checkbox, err := row.Element(selector)
	assert.Nil(m.Frame.T, err)

	checkedAttr, err := checkbox.Attribute("checked")
	assert.Nil(m.Frame.T, err)
	isChecked := checkedAttr != nil
	if isChecked == expectedChecked {
		return
	}
	checkbox.MustClick()
}

func (m *MaintenancePage) readAppEntry(row *rod.Element) *MaintenanceAppEntry {
	maintainerCell, err := row.Element(".maintenance-maintainer-cell")
	assert.Nil(m.Frame.T, err)
	maintainer, err := maintainerCell.Text()
	assert.Nil(m.Frame.T, err)

	appNameCell, err := row.Element(".maintenance-app-name-cell")
	assert.Nil(m.Frame.T, err)
	appName, err := appNameCell.Text()
	assert.Nil(m.Frame.T, err)

	autoUpdatesPresent, autoUpdatesChecked := m.readCheckboxState(row, ".auto-update-cell input[type='checkbox']")
	autoBackupsPresent, autoBackupsChecked := m.readCheckboxState(row, ".auto-backup-cell input[type='checkbox']")

	return &MaintenanceAppEntry{
		Maintainer: strings.TrimSpace(maintainer),
		AppName:    strings.TrimSpace(appName),

		AutomaticUpdatesCheckboxPresent: autoUpdatesPresent,
		AutomaticUpdatesEnabled:         autoUpdatesChecked,

		AutomaticBackupsCheckboxPresent: autoBackupsPresent,
		AutomaticBackupsEnabled:         autoBackupsChecked,
	}
}

func (m *MaintenancePage) readCheckboxState(row *rod.Element, selector string) (present bool, checked bool) {
	present, _, err := row.Has(selector)
	assert.Nil(m.Frame.T, err)
	if !present {
		return false, false
	}
	checkbox, err := row.Element(selector)
	assert.Nil(m.Frame.T, err)
	checkedAttr, err := checkbox.Attribute("checked")
	assert.Nil(m.Frame.T, err)
	return true, checkedAttr != nil
}

func (m *MaintenancePage) findRowByAppName(appName string) (*rod.Element, error) {
	rows, err := m.Frame.Page.Elements("tr.maintenance-row")
	assert.Nil(m.Frame.T, err)
	for _, row := range rows {
		appNameAttr, err := row.Attribute("data-app-name")
		if err == nil && appNameAttr != nil && strings.TrimSpace(*appNameAttr) == appName {
			return row, nil
		}
	}
	return nil, u.Logger.NewError("maintenance row not found", "app_name", appName)
}

func (m *MaintenancePage) WaitForSampleMaintenanceState(expectedUpdatesEnabled, expectedBackupsEnabled bool) {
	err := tools.Eventually(func() error {
		updatedSample := component.GetInstalledSample(m.Frame.T, m.Frame.Client)
		if updatedSample.AutomaticUpdatesEnabled != expectedUpdatesEnabled {
			return u.Logger.NewError(
				"unexpected automatic updates state",
				"expected", expectedUpdatesEnabled,
				"actual", updatedSample.AutomaticUpdatesEnabled,
			)
		}
		if updatedSample.AutomaticBackupsEnabled != expectedBackupsEnabled {
			return u.Logger.NewError(
				"unexpected automatic backups state",
				"expected", expectedBackupsEnabled,
				"actual", updatedSample.AutomaticBackupsEnabled,
			)
		}
		return nil
	})
	assert.Nil(m.Frame.T, err)
}
