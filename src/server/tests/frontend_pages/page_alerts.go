package frontend_pages

import (
	"strings"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/browsertest"
)

type AlertsPage struct {
	Frame *FrameType
}

type AlertEntry struct {
	Id      string
	Enabled bool
	Status  string
}

func (p *AlertsPage) ListAlerts() []AlertEntry {
	rows := p.Frame.Page.MustElements("tr.alert-row")
	alerts := make([]AlertEntry, 0, len(rows))
	for _, row := range rows {
		alerts = append(alerts, p.readAlert(row))
	}
	return alerts
}

func (p *AlertsPage) GetRequiredAlert(id string) *AlertEntry {
	for _, alert := range p.ListAlerts() {
		if alert.Id == id {
			return &alert
		}
	}
	assert.NotNil(p.Frame.T, nil)
	return nil
}

func (p *AlertsPage) SetAlertEnabled(id string, enabled bool) *AlertsPage {
	p.Frame.Controls.SetCheckboxValue("#alert-"+id+"-enabled", enabled)
	return p
}

func (p *AlertsPage) SaveAndAssertSuccessSnackbar() *AlertsPage {
	p.Frame.Controls.GetRequiredElement("#alert-settings-save-button").MustClick()
	p.Frame.Assert.SnackbarVisibleWithTextEventually("Alert settings saved.")
	return p
}

func (p *AlertsPage) readAlert(row *browsertest.Element) AlertEntry {
	id, err := row.Attribute("data-alert-id")
	assert.Nil(p.Frame.T, err)
	assert.NotNil(p.Frame.T, id)

	checkbox, err := row.Element("input[type='checkbox']")
	assert.Nil(p.Frame.T, err)
	checked, err := checkbox.Attribute("checked")
	assert.Nil(p.Frame.T, err)

	statusElement, err := row.Element(".alert-status")
	assert.Nil(p.Frame.T, err)
	status, err := statusElement.Text()
	assert.Nil(p.Frame.T, err)

	return AlertEntry{
		Id:      strings.TrimSpace(*id),
		Enabled: checked != nil,
		Status:  strings.TrimSpace(status),
	}
}
