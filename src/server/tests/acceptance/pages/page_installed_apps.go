//go:build acceptance

package pages

import (
	"fmt"
	"net/url"
	"server/apps_basic"
	"server/tools"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

const (
	sampleAppHost             = "sampleapp.localhost"
	sampleAppPath             = "/"
	sampleAppExpectedBodyText = "this is version 2.0"
)

type InstalledAppsPage struct {
	Frame *FrameType
}

type InstalledAppEntry struct {
	Maintainer        string
	AppName           string
	Version           string
	VersionCreated    string
	Access            string
	Status            string
	IsRunning         bool
	OpenButtonPresent bool
	OpenButtonEnabled bool
}

func (i *InstalledAppsPage) AssertHeaderColumnCount(expected int) *InstalledAppsPage {
	headers, err := i.Frame.page.Elements("#installed-apps-table thead th")
	assert.Nil(i.Frame.t, err)
	assert.Equal(i.Frame.t, expected, len(headers))
	return i
}

func (i *InstalledAppsPage) AssertHasApp(maintainer, appName, versionPrefix string) *InstalledAppsPage {
	apps := i.listAppEntries()
	for _, app := range apps {
		if app.Maintainer == maintainer && app.AppName == appName && strings.HasPrefix(app.Version, versionPrefix) {
			return i
		}
	}
	assert.True(i.Frame.t, false)
	return i
}

func (i *InstalledAppsPage) ListApps() []apps_basic.AppDto {
	entries := i.listAppEntries()
	out := make([]apps_basic.AppDto, 0, len(entries))
	for _, entry := range entries {
		out = append(out, apps_basic.AppDto{
			Maintainer:  entry.Maintainer,
			AppName:     entry.AppName,
			VersionName: entry.Version,
		})
	}
	return out
}

func (i *InstalledAppsPage) listAppEntries() []InstalledAppEntry {
	rows, err := i.Frame.page.Elements(`#installed-apps-tbody tr`)
	assert.Nil(i.Frame.t, err)

	out := make([]InstalledAppEntry, 0, len(rows))
	for _, r := range rows {
		entry, entryErr := i.readAppEntry(r)
		assert.Nil(i.Frame.t, entryErr)
		out = append(out, *entry)
	}

	return out
}

func (i *InstalledAppsPage) GetApp(appName string) *InstalledAppEntry {
	rows := i.listAppEntries()
	for _, entry := range rows {
		if entry.AppName == appName {
			appCopy := entry
			return &appCopy
		}
	}
	return nil
}

func (i *InstalledAppsPage) IsOpenButtonPresent(appName string) bool {
	return i.GetRequiredApp(appName).OpenButtonPresent
}

func (i *InstalledAppsPage) ClickOpenButton(appName string) *InstalledAppsPage {
	row, err := i.findRowByAppName(appName)
	assert.Nil(i.Frame.t, err)
	accessCell, err := row.Element(".app-access")
	assert.Nil(i.Frame.t, err)
	openButton, err := accessCell.Element("button.open-btn")
	assert.Nil(i.Frame.t, err)
	openButton.MustClick()
	return i
}

func (i *InstalledAppsPage) AssertDocsLinkHref(appName, expectedHref string) *InstalledAppsPage {
	row, err := i.findRowByAppName(appName)
	assert.Nil(i.Frame.t, err)
	docsLink, err := row.Element("a.app-homepage-link")
	assert.Nil(i.Frame.t, err)
	href, err := docsLink.Attribute("href")
	assert.Nil(i.Frame.t, err)
	assert.NotNil(i.Frame.t, href)
	assert.Equal(i.Frame.t, expectedHref, strings.TrimSpace(*href))
	return i
}

func (i *InstalledAppsPage) AssertVersionCreatedTooltip(appName, expectedTooltip string) *InstalledAppsPage {
	row, err := i.findRowByAppName(appName)
	assert.Nil(i.Frame.t, err)
	versionCreatedCell, err := row.Element(".app-version-created")
	assert.Nil(i.Frame.t, err)
	title, err := versionCreatedCell.Attribute("title")
	assert.Nil(i.Frame.t, err)
	assert.NotNil(i.Frame.t, title)
	assert.Equal(i.Frame.t, expectedTooltip, strings.TrimSpace(*title))
	return i
}

func (i *InstalledAppsPage) StartAppViaOperations(appName string) *InstalledAppsPage {
	return i.ExecuteOperationAndWait(appName, "Start", false)
}

func (i *InstalledAppsPage) StopAppViaOperations(appName string) *InstalledAppsPage {
	return i.ExecuteOperationAndWait(appName, "Stop", true)
}

func (i *InstalledAppsPage) UpdateAppViaOperations(appName string) *InstalledAppsPage {
	return i.ExecuteOperationAndWait(appName, "Update", false)
}

func (i *InstalledAppsPage) DeleteAppViaOperations(appName string) *InstalledAppsPage {
	return i.ExecuteOperationAndWait(appName, "Delete", true)
}

func (i *InstalledAppsPage) BackupAppViaOperations(appName string) *InstalledAppsPage {
	return i.ExecuteOperation(appName, "Backup", false)
}

func (i *InstalledAppsPage) AssertNoOngoingAppOperation() *InstalledAppsPage {
	_, isOngoing := i.Frame.Client.Apps.GetCurrentOperations()
	assert.False(i.Frame.t, isOngoing)
	return i
}

func (i *InstalledAppsPage) ExecuteOperation(appName, operationLabel string, shouldConfirm bool) *InstalledAppsPage {
	row, err := i.findRowByAppName(appName)
	assert.Nil(i.Frame.t, err)
	operationsSelect, err := row.Element("select.ops-select")
	assert.Nil(i.Frame.t, err)
	operationsSelect.MustSelect(operationLabel)
	selectedOption, err := operationsSelect.Element("option:checked")
	assert.Nil(i.Frame.t, err)
	selectedLabel, err := selectedOption.Text()
	assert.Nil(i.Frame.t, err)
	assert.Equal(i.Frame.t, "Operations", strings.TrimSpace(selectedLabel))
	if shouldConfirm {
		i.Frame.ConfirmDialog()
	}
	return i
}

func (i *InstalledAppsPage) ExecuteOperationAndWait(appName, operationLabel string, shouldConfirm bool) *InstalledAppsPage {
	i.ExecuteOperation(appName, operationLabel, shouldConfirm)
	i.Frame.AssertAppOperationStartedAndFinished().ReloadPage()
	return i
}

func (i *InstalledAppsPage) SetAccessPolicyPublic(appName string) *InstalledAppsPage {
	return i.SetAccessPolicyViaGui(appName, "Public", true)
}

func (i *InstalledAppsPage) SetAccessPolicyAdminOnly(appName string) *InstalledAppsPage {
	return i.SetAccessPolicyViaGui(appName, "Admin only", false)
}

func (i *InstalledAppsPage) SetAccessPolicyAuthenticated(appName string) *InstalledAppsPage {
	return i.SetAccessPolicyViaGui(appName, "Authenticated", true)
}

func (i *InstalledAppsPage) SetAccessPolicyGroupRestricted(appName string) *InstalledAppsPage {
	return i.SetAccessPolicyViaGui(appName, "Group restricted", true)
}

func (i *InstalledAppsPage) SetAccessPolicyViaGui(appName, accessPolicyLabel string, shouldConfirm bool) *InstalledAppsPage {
	row, err := i.findRowByAppName(appName)
	assert.Nil(i.Frame.t, err)
	accessPolicySelect, err := row.Element("select.vis-select")
	assert.Nil(i.Frame.t, err)
	accessPolicySelect.MustSelect(accessPolicyLabel)
	if shouldConfirm {
		i.Frame.ConfirmDialog()
	}
	i.Frame.AssertSnackbarVisibleWithTextEventually("Access policy changed successfully.")
	return i
}

func (i *InstalledAppsPage) AssertNoAccessPolicySelector(appName string) *InstalledAppsPage {
	row, err := i.findRowByAppName(appName)
	assert.Nil(i.Frame.t, err)
	hasSelector, _, hasErr := row.Has("select.vis-select")
	assert.Nil(i.Frame.t, hasErr)
	assert.False(i.Frame.t, hasSelector)
	return i
}

func (i *InstalledAppsPage) AssertAccessPolicyOptionPresent(appName, optionLabel string) *InstalledAppsPage {
	assert.True(i.Frame.t, i.hasAccessPolicyOption(appName, optionLabel))
	return i
}

func (i *InstalledAppsPage) AssertAccessPolicyOptionNotPresent(appName, optionLabel string) *InstalledAppsPage {
	assert.False(i.Frame.t, i.hasAccessPolicyOption(appName, optionLabel))
	return i
}

func (i *InstalledAppsPage) AssertOperationOptionPresent(appName, optionLabel string) *InstalledAppsPage {
	assert.True(i.Frame.t, i.hasOperationOption(appName, optionLabel))
	return i
}

func (i *InstalledAppsPage) AssertOperationOptionNotPresent(appName, optionLabel string) *InstalledAppsPage {
	assert.False(i.Frame.t, i.hasOperationOption(appName, optionLabel))
	return i
}

func (i *InstalledAppsPage) AssertSelectedAccessPolicy(optionAppName, expectedPolicyLabel string) *InstalledAppsPage {
	selected, err := i.getSelectedAccessPolicyLabel(optionAppName)
	assert.Nil(i.Frame.t, err)
	assert.Equal(i.Frame.t, expectedPolicyLabel, selected)
	return i
}

func (i *InstalledAppsPage) AssertSelectedAccessPolicyEventually(optionAppName, expectedPolicyLabel string) *InstalledAppsPage {
	selected, err := i.getSelectedAccessPolicyLabel(optionAppName)
	assert.Nil(i.Frame.t, err)
	assert.Equal(i.Frame.t, expectedPolicyLabel, selected)
	return i
}

func (i *InstalledAppsPage) hasAccessPolicyOption(appName, optionLabel string) bool {
	row, err := i.findRowByAppName(appName)
	assert.Nil(i.Frame.t, err)
	selectElement, err := row.Element("select.vis-select")
	assert.Nil(i.Frame.t, err)
	options, err := selectElement.Elements("option")
	assert.Nil(i.Frame.t, err)
	for _, option := range options {
		text, textErr := option.Text()
		assert.Nil(i.Frame.t, textErr)
		if strings.TrimSpace(text) == optionLabel {
			return true
		}
	}
	return false
}

func (i *InstalledAppsPage) hasOperationOption(appName, optionLabel string) bool {
	row, err := i.findRowByAppName(appName)
	assert.Nil(i.Frame.t, err)
	selectElement, err := row.Element("select.ops-select")
	assert.Nil(i.Frame.t, err)
	options, err := selectElement.Elements("option")
	assert.Nil(i.Frame.t, err)
	for _, option := range options {
		text, textErr := option.Text()
		assert.Nil(i.Frame.t, textErr)
		if strings.TrimSpace(text) == optionLabel {
			return true
		}
	}
	return false
}

func (i *InstalledAppsPage) getSelectedAccessPolicyLabel(appName string) (string, error) {
	row, err := i.findRowByAppName(appName)
	if err != nil {
		return "", err
	}
	selectElement, err := row.Element("select.vis-select")
	if err != nil {
		return "", u.Logger.NewError(err.Error())
	}
	selectedOption, err := selectElement.Element("option:checked")
	if err != nil {
		return "", u.Logger.NewError(err.Error())
	}
	selectedLabel, err := selectedOption.Text()
	if err != nil {
		return "", u.Logger.NewError(err.Error())
	}
	return strings.TrimSpace(selectedLabel), nil
}

func (i *InstalledAppsPage) OpenSampleAppInNewTabAndAssertContent() *InstalledAppsPage {
	const appName = "sampleapp"
	waitForNewTab := i.Frame.page.MustWaitOpen()
	i.ClickOpenButton(appName)
	newTab := waitForNewTab().Timeout(defaultTimeout)
	defer func() {
		assert.Nil(i.Frame.t, newTab.Close())
	}()

	// After app start, Docker/network handover can briefly surface Chrome's transient ERR_NETWORK_CHANGED page
	// even though the backend operation already finished. Retry until the tab shows the actual app content.
	err := tools.EventuallyWithTimeout(backupOperationTimeout, 50*time.Millisecond, func() error {
		newTab.MustWaitLoad()
		info, err := newTab.Info()
		if err != nil {
			return err
		}
		parsedURL, err := url.Parse(info.URL)
		if err != nil {
			return err
		}
		if parsedURL.Host != sampleAppHost {
			return fmt.Errorf("unexpected sample app host: %q", parsedURL.Host)
		}
		if parsedURL.Path != sampleAppPath {
			return fmt.Errorf("unexpected sample app path: %q", parsedURL.Path)
		}

		bodyText, err := newTab.Element("body")
		if err != nil {
			return err
		}
		text, err := bodyText.Text()
		if err != nil {
			return err
		}
		text = strings.TrimSpace(text)
		if strings.Contains(text, "ERR_NETWORK_CHANGED") {
			newTab.MustReload()
			return fmt.Errorf("sample app tab still shows transient ERR_NETWORK_CHANGED page")
		}
		if text != sampleAppExpectedBodyText {
			return fmt.Errorf("unexpected sample app body text: %q", text)
		}
		return nil
	})
	assert.Nil(i.Frame.t, err)
	return i
}

func (i *InstalledAppsPage) AssertAppStatusAndOpenButtonEventually(appName string, expectedIsRunning bool, expectedOpenButtonEnabled bool) *InstalledAppsPage {
	err := tools.Eventually(func() error {
		i.Frame.ReloadPage()

		row, rowErr := i.findRowByAppName(appName)
		if rowErr != nil {
			return rowErr
		}
		app, readErr := i.readAppEntry(row)
		if readErr != nil {
			return readErr
		}
		if app.IsRunning != expectedIsRunning {
			return u.Logger.NewError("unexpected app running state", "app_name", appName, "expected_is_running", expectedIsRunning, "actual_is_running", app.IsRunning)
		}
		if !app.OpenButtonPresent {
			return u.Logger.NewError("open button missing", "app_name", appName)
		}
		if app.OpenButtonEnabled != expectedOpenButtonEnabled {
			return u.Logger.NewError("unexpected open button state", "app_name", appName, "expected_enabled", expectedOpenButtonEnabled, "actual_enabled", app.OpenButtonEnabled)
		}
		return nil
	})
	assert.Nil(i.Frame.t, err)
	return i
}

func (i *InstalledAppsPage) GetRequiredApp(appName string) *InstalledAppEntry {
	app := i.GetApp(appName)
	assert.NotNil(i.Frame.t, app)
	return app
}

func (i *InstalledAppsPage) readAppEntry(row *rod.Element) (*InstalledAppEntry, error) {
	nameCell, err := row.Element(".app-name")
	if err != nil {
		return nil, u.Logger.NewError("unexpected installed apps row")
	}

	maintainer, err := getOptionalCellText(row, ".app-maintainer")
	if err != nil {
		return nil, err
	}

	appName, err := nameCell.Text()
	if err != nil {
		return nil, err
	}

	version, err := getOptionalCellText(row, ".app-version")
	if err != nil {
		return nil, err
	}

	versionCreated, err := getOptionalCellText(row, ".app-version-created")
	if err != nil {
		return nil, err
	}

	accessCell, err := row.Element(".app-access")
	if err != nil {
		return nil, err
	}
	access, err := accessCell.Text()
	if err != nil {
		return nil, err
	}

	openButtonPresent, _, hasOpenButtonErr := accessCell.Has("button.open-btn")
	if hasOpenButtonErr != nil {
		return nil, u.Logger.NewError(hasOpenButtonErr.Error())
	}
	openButtonEnabled := false
	if openButtonPresent {
		openButton, openButtonErr := accessCell.Element("button.open-btn")
		if openButtonErr != nil {
			return nil, u.Logger.NewError(openButtonErr.Error())
		}
		disabledAttribute, disabledErr := openButton.Attribute("disabled")
		if disabledErr != nil {
			return nil, u.Logger.NewError(disabledErr.Error())
		}
		openButtonEnabled = disabledAttribute == nil
	}

	status := ""
	isRunning := false
	statusCellVisible, _, hasStatusCellErr := row.Has(".app-status")
	if hasStatusCellErr != nil {
		return nil, u.Logger.NewError(hasStatusCellErr.Error())
	}
	if statusCellVisible {
		statusCell, statusCellErr := row.Element(".app-status")
		if statusCellErr != nil {
			return nil, u.Logger.NewError(statusCellErr.Error())
		}
		statusText, statusTextErr := statusCell.Text()
		if statusTextErr != nil {
			return nil, u.Logger.NewError(statusTextErr.Error())
		}
		status = statusText
		isRunning = strings.TrimSpace(statusText) == "Running"
	}

	return &InstalledAppEntry{
		Maintainer:        strings.TrimSpace(maintainer),
		AppName:           strings.TrimSpace(appName),
		Version:           strings.TrimSpace(version),
		VersionCreated:    strings.TrimSpace(versionCreated),
		Access:            strings.TrimSpace(access),
		Status:            strings.TrimSpace(status),
		IsRunning:         isRunning,
		OpenButtonPresent: openButtonPresent,
		OpenButtonEnabled: openButtonEnabled,
	}, nil
}

func getOptionalCellText(row *rod.Element, selector string) (string, error) {
	hasCell, _, hasErr := row.Has(selector)
	if hasErr != nil {
		return "", u.Logger.NewError(hasErr.Error())
	}
	if !hasCell {
		return "", nil
	}
	cell, err := row.Element(selector)
	if err != nil {
		return "", u.Logger.NewError(err.Error())
	}
	text, err := cell.Text()
	if err != nil {
		return "", u.Logger.NewError(err.Error())
	}
	return text, nil
}

func (i *InstalledAppsPage) findRowByAppName(appName string) (*rod.Element, error) {
	rows, err := i.Frame.page.Elements(`#installed-apps-tbody tr`)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	for _, row := range rows {
		appNameAttr, err := row.Attribute("data-app-name")
		if err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		if appNameAttr != nil && strings.TrimSpace(*appNameAttr) == appName {
			return row, nil
		}
	}
	return nil, u.Logger.NewError("app row not found", "app_name", appName)
}
