package frontend_pages

import (
	"fmt"
	"server/tools"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/quollix/common/assert"
)

const (
	expectedShellBanner    = "Shell: /bin/sh"
	expectedTerminalPrompt = "/opt/server #"
)

type TerminalAppRow struct {
	Maintainer string
	AppName    string
}

type TerminalAppsPage struct {
	Frame *FrameType
}

type TerminalServicesPage struct {
	Frame *FrameType
}

type TerminalViewPage struct {
	Frame *FrameType
}

func (p *TerminalAppsPage) ListApps() []TerminalAppRow {
	rows, err := p.Frame.Page.Elements("tr.terminal-app-row")
	assert.Nil(p.Frame.T, err)

	out := make([]TerminalAppRow, 0, len(rows))
	for _, row := range rows {
		maintainer, err := row.Element(".terminal-app-maintainer-cell")
		assert.Nil(p.Frame.T, err)
		maintainerText, err := maintainer.Text()
		assert.Nil(p.Frame.T, err)

		appName, err := row.Element(".terminal-app-name-cell")
		assert.Nil(p.Frame.T, err)
		appNameText, err := appName.Text()
		assert.Nil(p.Frame.T, err)

		out = append(out, TerminalAppRow{
			Maintainer: strings.TrimSpace(maintainerText),
			AppName:    strings.TrimSpace(appNameText),
		})
	}
	return out
}

func (p *TerminalAppsPage) AssertAppPresent(maintainer, appName string) *TerminalAppsPage {
	row := p.getRequiredRow(maintainer, appName)
	assert.NotNil(p.Frame.T, row)
	return p
}

func (p *TerminalAppsPage) OpenServicesPage(maintainer, appName string) *TerminalServicesPage {
	row := p.getRequiredRow(maintainer, appName)
	button, err := row.Element("button.terminal-view-services-button")
	assert.Nil(p.Frame.T, err)
	p.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		button.MustClick()
	})
	p.Frame.Assert.PagePath(tools.Paths.FrontendTerminalServices)
	return p.Frame.Pages.TerminalServicesPage
}

func (p *TerminalAppsPage) getRequiredRow(maintainer, appName string) *rod.Element {
	rows, err := p.Frame.Page.Elements("tr.terminal-app-row")
	assert.Nil(p.Frame.T, err)
	for _, row := range rows {
		rowMaintainer, err := row.Attribute("data-maintainer")
		assert.Nil(p.Frame.T, err)
		rowAppName, err := row.Attribute("data-app-name")
		assert.Nil(p.Frame.T, err)
		if rowMaintainer != nil && rowAppName != nil && *rowMaintainer == maintainer && *rowAppName == appName {
			return row
		}
	}
	p.Frame.T.Fatalf("terminal app row not found maintainer=%s app=%s", maintainer, appName)
	return nil
}

func (p *TerminalServicesPage) ClickBackAndAssertTerminalAppsPage() *TerminalAppsPage {
	button := p.Frame.Controls.GetRequiredElement("#terminal-services-back-button")
	p.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		button.MustClick()
	})
	p.Frame.Assert.PagePath(tools.Paths.FrontendTerminalApps)
	return p.Frame.Pages.TerminalAppsPage
}

func (p *TerminalServicesPage) AssertSelection(maintainer, appName string) *TerminalServicesPage {
	assert.Equal(p.Frame.T, "Maintainer: "+maintainer, strings.TrimSpace(p.Frame.Controls.GetRequiredElement("#terminal-services-maintainer").MustText()))
	assert.Equal(p.Frame.T, "App Name: "+appName, strings.TrimSpace(p.Frame.Controls.GetRequiredElement("#terminal-services-app-name").MustText()))
	return p
}

func (p *TerminalServicesPage) AssertServicePresent(serviceName string) *TerminalServicesPage {
	row := p.getRequiredRow(serviceName)
	assert.NotNil(p.Frame.T, row)
	return p
}

func (p *TerminalServicesPage) OpenTerminal(serviceName string) *TerminalViewPage {
	row := p.getRequiredRow(serviceName)
	button, err := row.Element("button.terminal-open-button")
	assert.Nil(p.Frame.T, err)
	p.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		button.MustClick()
	})
	p.Frame.Assert.PagePath(tools.Paths.FrontendTerminalView)
	return p.Frame.Pages.TerminalViewPage
}

func (p *TerminalServicesPage) getRequiredRow(serviceName string) *rod.Element {
	rows, err := p.Frame.Page.Elements("tr.terminal-service-row")
	assert.Nil(p.Frame.T, err)
	for _, row := range rows {
		rowServiceName, err := row.Attribute("data-service-name")
		assert.Nil(p.Frame.T, err)
		if rowServiceName != nil && *rowServiceName == serviceName {
			return row
		}
	}
	p.Frame.T.Fatalf("terminal service row not found service=%s", serviceName)
	return nil
}

func (p *TerminalViewPage) ClickBackAndAssertTerminalServicesPage() *TerminalServicesPage {
	button := p.Frame.Controls.GetRequiredElement("#terminal-view-back-button")
	p.Frame.Browser.DoAndWaitDOMContentLoaded(func() {
		button.MustClick()
	})
	p.Frame.Assert.PagePath(tools.Paths.FrontendTerminalServices)
	return p.Frame.Pages.TerminalServicesPage
}

func (p *TerminalViewPage) AssertSelection(maintainer, appName, serviceName string) *TerminalViewPage {
	assert.Equal(p.Frame.T, "Maintainer: "+maintainer, strings.TrimSpace(p.Frame.Controls.GetRequiredElement("#terminal-view-maintainer").MustText()))
	assert.Equal(p.Frame.T, "App Name: "+appName, strings.TrimSpace(p.Frame.Controls.GetRequiredElement("#terminal-view-app-name").MustText()))
	assert.Equal(p.Frame.T, "Service: "+serviceName, strings.TrimSpace(p.Frame.Controls.GetRequiredElement("#terminal-view-service-name").MustText()))
	return p
}

func (p *TerminalViewPage) AssertReady() *TerminalViewPage {
	err := tools.Eventually(func() error {
		output := p.readTerminalBuffer()
		if strings.Contains(output, expectedShellBanner+"\n") || strings.Contains(output, expectedTerminalPrompt) {
			return nil
		}
		return fmt.Errorf("terminal not ready yet")
	})
	assert.Nil(p.Frame.T, err)
	return p
}

func (p *TerminalViewPage) RunCommand(command string) *TerminalViewPage {
	textarea := p.Frame.Controls.GetRequiredElement(".xterm-helper-textarea")
	textarea.MustClick()
	keys := make([]input.Key, 0, len(command)+1)
	for _, r := range command {
		keys = append(keys, input.Key(r))
	}
	textarea.MustKeyActions().Type(keys...).Type(input.Enter).MustDo()
	return p
}

func (p *TerminalViewPage) AssertOutputContains(expected string) *TerminalViewPage {
	expected = normalizeTerminalAssertionText(expected)
	err := tools.Eventually(func() error {
		output := normalizeTerminalAssertionText(p.readTerminalBuffer())
		if !strings.Contains(output, expected) {
			return fmt.Errorf("missing terminal output block:\n%s\n\nin output:\n%s", expected, output)
		}
		return nil
	})
	assert.Nil(p.Frame.T, err)
	return p
}

func (p *TerminalViewPage) readTerminalBuffer() string {
	rows, err := p.Frame.Page.Elements("#terminal-output-accessibility [role='listitem']")
	assert.Nil(p.Frame.T, err)

	rowTexts := make([]string, 0, len(rows))
	for _, row := range rows {
		text, textErr := row.Text()
		assert.Nil(p.Frame.T, textErr)
		rowTexts = append(rowTexts, strings.TrimSpace(text))
	}

	return strings.TrimSpace(strings.Join(rowTexts, "\n"))
}

func normalizeTerminalAssertionText(value string) string {
	lines := strings.Split(value, "\n")
	for index, line := range lines {
		lines[index] = strings.TrimRight(line, " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
