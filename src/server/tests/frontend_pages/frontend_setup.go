//go:build frontend

package frontend_pages

import (
	"testing"

	"github.com/quollix/common/browsertest"
	"github.com/quollix/common/quollix/api_client"
)

var (
	browser                          *browsertest.Browser
	page                             *browsertest.Page
	wasFrontendReloadedDuringThisRun = false
)

func Setup(t *testing.T) *FrameType {
	if browser == nil {
		browser = LaunchBrowser()
		page = browser.InitialPage()
	}

	frame := NewFrameType(t, "https://quollix.localhost", page, api_client.NewQuollixClient())
	frame.Session.SignInAsAdminViaClient()

	if !wasFrontendReloadedDuringThisRun {
		// this means, we can make changes to frontend and simply re-run the frontend tests with latest changes, without having to re-redeploy the quollix container
		frame.Client.Frontend.Reload()
		wasFrontendReloadedDuringThisRun = true
	}
	return frame
}

func CloseBrowser() {
	if browser == nil {
		return
	}
	if err := browser.Close(); err != nil {
		panic(err.Error())
	}
	browser = nil
	page = nil
}
