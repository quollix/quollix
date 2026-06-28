package frontend_pages

import (
	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
)

type FrameControls struct {
	Frame *FrameType
}

func (c *FrameControls) GetRequiredElement(selector string) *rod.Element {
	element, err := c.Frame.Page.Element(selector)
	assert.Nil(c.Frame.T, err)
	return element
}

func (c *FrameControls) SetInputValue(selector, value string) {
	c.GetRequiredElement(selector).MustSelectAllText().MustInput(value)
}

func (c *FrameControls) GetInputValue(selector string) string {
	value, err := c.GetRequiredElement(selector).Property("value")
	assert.Nil(c.Frame.T, err)
	return value.String()
}

func (c *FrameControls) GetCheckboxValue(selector string) bool {
	checked, err := c.GetRequiredElement(selector).Property("checked")
	assert.Nil(c.Frame.T, err)
	return checked.Bool()
}

func (c *FrameControls) SetCheckboxValue(selector string, checked bool) {
	if c.GetCheckboxValue(selector) != checked {
		c.GetRequiredElement(selector).MustClick()
	}
}

func (c *FrameControls) GetInputType(selector string) string {
	value, err := c.GetRequiredElement(selector).Property("type")
	assert.Nil(c.Frame.T, err)
	return value.String()
}
