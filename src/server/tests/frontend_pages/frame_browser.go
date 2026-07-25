package frontend_pages

import (
	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

type FrameBrowser struct {
	Frame *FrameType
}

func (b *FrameBrowser) DoAndWaitDOMContentLoaded(action func()) {
	assert.Nil(b.Frame.T, b.Frame.Page.DoAndWaitLoad(func() error {
		action()
		return nil
	}))
}

func (b *FrameBrowser) WaitForElement(selector string) *FrameType {
	_, err := b.Frame.Page.WaitElementWithin(selector, browserTimeout)
	assert.Nil(b.Frame.T, err)
	return b.Frame
}

func (b *FrameBrowser) ReloadPage() *FrameType {
	b.DoAndWaitDOMContentLoaded(func() {
		b.Frame.Page.MustReload()
	})
	return b.Frame
}

func (b *FrameBrowser) ClickSidebarLink(groupId, itemId string) {
	assert.Nil(b.Frame.T, b.Frame.Page.ClickElementWithin("#"+groupId+" > summary", browserTimeout))
	assert.Nil(b.Frame.T, b.Frame.Page.ClickElementWithin("#"+itemId, browserTimeout))
}

func (b *FrameBrowser) ClickSidebarUserLink() {
	assert.Nil(b.Frame.T, b.Frame.Page.ClickElementWithin("#sidebar-user-link", browserTimeout))
	b.Frame.Assert.PathEventually(api.Paths.FrontendAccount)
}

func (b *FrameBrowser) ClickSidebarTopLevelLink(itemID string) {
	assert.Nil(b.Frame.T, b.Frame.Page.ClickElementWithin("#"+itemID, browserTimeout))
}

func (b *FrameBrowser) ConfirmDialog() *FrameType {
	assert.Nil(b.Frame.T, b.Frame.Page.ClickElement("#confirm-button"))
	return b.Frame
}
