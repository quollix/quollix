package frontend_pages

import (
	"server/tools"
	"time"

	"github.com/go-rod/rod/lib/proto"
	"github.com/quollix/common/assert"
)

type FrameBrowser struct {
	Frame *FrameType
}

func (b *FrameBrowser) DoAndWaitDOMContentLoaded(action func()) {
	waitForNavigation := b.Frame.Page.WaitNavigation(proto.PageLifecycleEventNameDOMContentLoaded)
	action()
	waitForNavigation()
}

func (b *FrameBrowser) WaitForElement(selector string) *FrameType {
	err := tools.EventuallyWithTimeout(browserTimeout, 50*time.Millisecond, func() error {
		_, findErr := b.Frame.Page.Element(selector)
		return findErr
	})
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
	b.Frame.Page.MustElement("#" + groupId + " > summary").MustClick()
	b.DoAndWaitDOMContentLoaded(func() {
		b.Frame.Page.MustElement("#" + itemId).MustClick()
	})
}

func (b *FrameBrowser) ClickSidebarUserLink() {
	b.DoAndWaitDOMContentLoaded(func() {
		b.Frame.Page.MustElement("#sidebar-user-link").MustClick()
	})
}

func (b *FrameBrowser) ClickSidebarTopLevelLink(itemID string) {
	b.DoAndWaitDOMContentLoaded(func() {
		b.Frame.Page.MustElement("#" + itemID).MustClick()
	})
}

func (b *FrameBrowser) ConfirmDialog() *FrameType {
	err := tools.Eventually(func() error {
		confirmButton, findErr := b.Frame.Page.Element("#confirm-button")
		if findErr != nil {
			return findErr
		}
		return confirmButton.Click(proto.InputMouseButtonLeft, 1)
	})
	assert.Nil(b.Frame.T, err)
	return b.Frame
}
