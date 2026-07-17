package frontend_pages

import (
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/browsertest"
	"github.com/quollix/common/quollix/api_client"
)

const browserTimeout = 10 * time.Second

func LaunchBrowser() *browsertest.Browser {
	return browsertest.MustLaunchBrowser()
}

func NewBrowserFrame(t *testing.T, baseUrl string, client *api_client.QuollixClient) *FrameType {
	browser := LaunchBrowser()
	t.Cleanup(func() {
		assert.Nil(t, browser.Close())
	})
	page := browser.MustPage()
	t.Cleanup(func() {
		assert.Nil(t, page.Close())
	})
	return NewFrameType(t, baseUrl, page, client)
}
