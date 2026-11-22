//go:build frontend

package frontend

import (
	"testing"

	"github.com/quollix/common/assert"

	"server/tests/frontend_pages"
)

func TestBlogNewsIconShowsUnreadPostsAndUpdatesAfterClick(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	frame.Pages.GoToInstalledAppsPage()
	frame.Assert.BlogNewsUnread()

	waitForBlogTab := frame.Page.WaitOpen()
	frame.Browser.ClickBlogNewsIcon()
	blogTab, err := waitForBlogTab()
	assert.Nil(t, err)
	assert.Nil(t, blogTab.Close())
	frame.Assert.BlogNewsRead()
}
