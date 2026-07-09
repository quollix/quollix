//go:build acceptance

package frontend_pages

import (
	"server/tests/api_client"
	"testing"

	"github.com/go-rod/rod"
	"github.com/quollix/common/assert"
)

var (
	browser                          *rod.Browser
	wasFrontendReloadedDuringThisRun = false
)

func Setup(t *testing.T) *FrameType {
	if browser == nil {
		// Initialize the shared root Rod client once. It owns the underlying Chrome process connection.
		browser = LaunchBrowser()
	}

	// Create a per-test incognito browser-context client.
	// This is a lightweight Rod handle for isolation; it does not start a new Chrome process.
	testBrowser, err := browser.Incognito()
	assert.Nil(t, err)

	page := testBrowser.MustPage()
	t.Cleanup(func() {
		assert.Nil(t, page.Close())
		assert.Nil(t, testBrowser.Close())
	})

	frame := NewFrameType(t, "https://quollix.localhost", page, api_client.NewQuollixClient())
	frame.Session.SignInAsAdminViaClient()

	if !wasFrontendReloadedDuringThisRun {
		// this means, we can make changes to frontend and simply re-run the acceptance tests with latest changes, without having to re-redeploy the quollix container
		frame.Client.Frontend.Reload()
		wasFrontendReloadedDuringThisRun = true
	}
	return frame
}

func CloseBrowser() {
	if browser == nil {
		return
	}
	// This closes the shared root Rod client and therefore terminates the underlying Chrome process.
	if err := browser.Close(); err != nil {
		panic(err.Error())
	}
	browser = nil
}
