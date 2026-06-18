//go:build acceptance

package pages

import (
	"strings"
	"testing"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
)

type SelectionTableHelper struct {
	T              *testing.T
	Page           *rod.Page
	TableBody      string
	RowSelector    func(string) string
	RowCheckboxCSS string
}

func (h SelectionTableHelper) SetFilter(selector, value string) {
	SetInputValue(h.T, h.Page, selector, value)
}

func (h SelectionTableHelper) AssertRowPresent(value string, expected bool) {
	tableBody := GetRequiredElement(h.T, h.Page, h.TableBody)
	hasRow, _, err := tableBody.Has(h.RowSelector(value))
	assert.Nil(h.T, err)
	assert.Equal(h.T, expected, hasRow)
}

func (h SelectionTableHelper) AssertRowVisible(value string, expectedVisible bool) {
	row := h.getRequiredRow(value)
	style, err := row.Attribute("style")
	assert.Nil(h.T, err)
	isVisible := style == nil || !strings.Contains(strings.ToLower(strings.TrimSpace(*style)), "display: none")
	assert.Equal(h.T, expectedVisible, isVisible)
}

func (h SelectionTableHelper) AssertCheckboxChecked(selector string, expected bool) {
	assert.Equal(h.T, expected, GetCheckboxValue(h.T, h.Page, selector))
}

func (h SelectionTableHelper) SetCheckbox(selector string, checked bool) {
	SetCheckboxValue(h.T, h.Page, selector, checked)
}

func (h SelectionTableHelper) AssertRowCheckboxChecked(value string, expected bool) {
	row := h.getRequiredRow(value)
	checkbox, err := row.Element(h.RowCheckboxCSS)
	assert.Nil(h.T, err)
	assert.Equal(h.T, expected, isCheckboxChecked(h.T, checkbox))
}

func (h SelectionTableHelper) SetRowCheckbox(value string, checked bool) {
	row := h.getRequiredRow(value)
	checkbox, err := row.Element(h.RowCheckboxCSS)
	assert.Nil(h.T, err)
	currentChecked := isCheckboxChecked(h.T, checkbox)
	if currentChecked != checked {
		checkbox.MustClick()
	}
}

func (h SelectionTableHelper) getRequiredRow(value string) *rod.Element {
	tableBody := GetRequiredElement(h.T, h.Page, h.TableBody)
	row, err := tableBody.Element(h.RowSelector(value))
	assert.Nil(h.T, err)
	return row
}

func isCheckboxChecked(t *testing.T, checkbox *rod.Element) bool {
	checkedProperty, err := checkbox.Property("checked")
	assert.Nil(t, err)
	return checkedProperty.Bool()
}
