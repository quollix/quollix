//go:build acceptance

package pages

import (
	"testing"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
)

func GetRequiredElement(t *testing.T, page *rod.Page, selector string) *rod.Element {
	element, err := page.Element(selector)
	assert.Nil(t, err)
	return element
}

func SetInputValue(t *testing.T, page *rod.Page, selector, value string) {
	GetRequiredElement(t, page, selector).MustSelectAllText().MustInput(value)
}

func GetInputValue(t *testing.T, page *rod.Page, selector string) string {
	value, err := GetRequiredElement(t, page, selector).Property("value")
	assert.Nil(t, err)
	return value.String()
}

func GetCheckboxValue(t *testing.T, page *rod.Page, selector string) bool {
	checked, err := GetRequiredElement(t, page, selector).Property("checked")
	assert.Nil(t, err)
	return checked.Bool()
}

func SetCheckboxValue(t *testing.T, page *rod.Page, selector string, checked bool) {
	if GetCheckboxValue(t, page, selector) != checked {
		GetRequiredElement(t, page, selector).MustClick()
	}
}

func GetInputType(t *testing.T, page *rod.Page, selector string) string {
	value, err := GetRequiredElement(t, page, selector).Property("type")
	assert.Nil(t, err)
	return value.String()
}
