package frontend_pages

import (
	"github.com/quollix/common/assert"
	"github.com/quollix/common/browsertest"
)

type FrameControls struct {
	Frame *FrameType
}

func (c *FrameControls) GetRequiredElement(selector string) *browsertest.Element {
	element, err := c.Frame.Page.Element(selector)
	assert.Nil(c.Frame.T, err)
	return element
}

func (c *FrameControls) GetRequiredElementEventually(selector string) *browsertest.Element {
	element, err := c.Frame.Page.WaitElementWithin(selector, browserTimeout)
	assert.Nil(c.Frame.T, err)
	return element
}

func (c *FrameControls) SetInputValue(selector, value string) {
	assert.Nil(c.Frame.T, c.Frame.Page.SetInputValue(selector, value))
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
	assert.Nil(c.Frame.T, c.Frame.Page.SetCheckboxChecked(selector, checked))
}

func (c *FrameControls) GetInputType(selector string) string {
	value, err := c.GetRequiredElement(selector).Property("type")
	assert.Nil(c.Frame.T, err)
	return value.String()
}
