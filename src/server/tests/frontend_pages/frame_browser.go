package frontend_pages

import (
	"strings"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
)

type FrameBrowser struct {
	Frame *FrameType
}

func (b *FrameBrowser) FlushLogs() *FrameType {
	assert.Nil(b.Frame.T, b.Frame.Page.ClickElementWithin("#flush-logs-button", browserTimeout))
	return b.Frame
}

func (b *FrameBrowser) CollectMetrics() *FrameType {
	assert.Nil(b.Frame.T, b.Frame.Page.ClickElementWithin("#collect-metrics-button", browserTimeout))
	return b.Frame
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
	assert.Nil(b.Frame.T, b.Frame.Page.ClickElementWithin("#confirm-dialog[open] #confirm-button", browserTimeout))
	return b.Frame
}

func (b *FrameBrowser) WipeLogs() *FrameType {
	b.wipeAndWait("#wipe-logs-button", "Logs emptied successfully.")
	return b.Frame
}

func (b *FrameBrowser) WipeMetrics() *FrameType {
	b.wipeAndWait("#wipe-metrics-button", "Metrics emptied successfully.")
	return b.Frame
}

// Repeating a wipe is safe for this test's disposable data. Only retry a failed
// fetch; a server error or a missing response must still fail the test.
func (b *FrameBrowser) wipeAndWait(selector, successMessage string) {
	assert.Nil(b.Frame.T, b.Frame.Page.ClickElementWithin(selector, browserTimeout))
	err := u.EventuallyWithTimeout(browserTimeout, 250*time.Millisecond, func() error {
		snackbars, err := b.Frame.Page.Elements(`.snackbar[data-visible="true"]`)
		if err != nil {
			return err
		}
		for _, snackbar := range snackbars {
			text, err := snackbar.Text()
			if err != nil {
				return err
			}
			switch strings.TrimSpace(text) {
			case successMessage:
				return nil
			case "Failed to fetch":
				snackbar.MustElement(".snackbar-close").MustClick()
				assert.Nil(b.Frame.T, b.Frame.Page.ClickElementWithin(selector, browserTimeout))
				return u.Logger.NewError("retrying wipe after failed fetch")
			}
		}
		return u.Logger.NewError("waiting for wipe success", "expected", successMessage)
	})
	assert.Nil(b.Frame.T, err)
}
