package frontend_pages

import (
	"os"
	"server/tests/api_client"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/quollix/common/assert"
)

const browserTimeout = 10 * time.Second

func LaunchBrowser() *rod.Browser {
	headless := os.Getenv("HEADFUL") != "true"
	noSandbox := os.Getenv("CI") == "true"
	launcherUrl := launcher.New().Headless(headless).NoSandbox(noSandbox).MustLaunch()
	return rod.New().ControlURL(launcherUrl).MustConnect().MustIgnoreCertErrors(true)
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
