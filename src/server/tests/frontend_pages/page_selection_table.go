package frontend_pages

import (
	"strings"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
)

type SelectionTableHelper struct {
	Frame          *FrameType
	TableBody      string
	RowSelector    func(string) string
	RowCheckboxCSS string
}

func (h SelectionTableHelper) SetFilter(selector, value string) {
	h.Frame.Controls.SetInputValue(selector, value)
}

func (h SelectionTableHelper) AssertRowPresent(value string, expected bool) {
	tableBody := h.Frame.Controls.GetRequiredElement(h.TableBody)
	hasRow, _, err := tableBody.Has(h.RowSelector(value))
	assert.Nil(h.Frame.T, err)
	assert.Equal(h.Frame.T, expected, hasRow)
}

func (h SelectionTableHelper) AssertRowVisible(value string, expectedVisible bool) {
	row := h.getRequiredRow(value)
	style, err := row.Attribute("style")
	assert.Nil(h.Frame.T, err)
	isVisible := style == nil || !strings.Contains(strings.ToLower(strings.TrimSpace(*style)), "display: none")
	assert.Equal(h.Frame.T, expectedVisible, isVisible)
}

func (h SelectionTableHelper) AssertCheckboxChecked(selector string, expected bool) {
	assert.Equal(h.Frame.T, expected, h.Frame.Controls.GetCheckboxValue(selector))
}

func (h SelectionTableHelper) SetCheckbox(selector string, checked bool) {
	h.Frame.Controls.SetCheckboxValue(selector, checked)
}

func (h SelectionTableHelper) AssertRowCheckboxChecked(value string, expected bool) {
	row := h.getRequiredRow(value)
	checkbox, err := row.Element(h.RowCheckboxCSS)
	assert.Nil(h.Frame.T, err)
	assert.Equal(h.Frame.T, expected, isCheckboxChecked(h.Frame, checkbox))
}

func (h SelectionTableHelper) SetRowCheckbox(value string, checked bool) {
	row := h.getRequiredRow(value)
	checkbox, err := row.Element(h.RowCheckboxCSS)
	assert.Nil(h.Frame.T, err)
	currentChecked := isCheckboxChecked(h.Frame, checkbox)
	if currentChecked != checked {
		checkbox.MustClick()
	}
}

func (h SelectionTableHelper) getRequiredRow(value string) *rod.Element {
	tableBody := h.Frame.Controls.GetRequiredElement(h.TableBody)
	row, err := tableBody.Element(h.RowSelector(value))
	assert.Nil(h.Frame.T, err)
	return row
}

func isCheckboxChecked(frame *FrameType, checkbox *rod.Element) bool {
	checkedProperty, err := checkbox.Property("checked")
	assert.Nil(frame.T, err)
	return checkedProperty.Bool()
}
