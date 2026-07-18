package frontend_pages

import (
	"testing"
	"time"

	"github.com/quollix/common/quollix/api_client"
	quollixbrowser "github.com/quollix/common/quollix/browser"

	"github.com/quollix/common/browsertest"
)

const defaultTimeout = 3 * time.Second
const backupOperationTimeout = 1 * time.Minute

type FrameType struct {
	T        *testing.T
	BaseUrl  string
	Page     *browsertest.Page
	Client   *api_client.QuollixClient
	Quollix  *quollixbrowser.Browser
	Controls *FrameControls
	Pages    *FramePages
	Assert   *FrameAssertions
	Session  *FrameSession
	Browser  *FrameBrowser
}

func NewFrameType(t *testing.T, baseUrl string, page *browsertest.Page, client *api_client.QuollixClient) *FrameType {
	frame := &FrameType{
		T:       t,
		BaseUrl: baseUrl,
		Page:    page,
		Client:  client,
	}
	frame.Quollix = &quollixbrowser.Browser{
		BaseURL:       baseUrl,
		Client:        client,
		Page:          page,
		InstalledApps: &quollixbrowser.InstalledAppsPageHelpers{Page: page},
	}
	frame.Controls = &FrameControls{Frame: frame}
	frame.Pages = newFramePages(frame)
	frame.Assert = &FrameAssertions{Frame: frame}
	frame.Session = &FrameSession{Frame: frame}
	frame.Browser = &FrameBrowser{Frame: frame}
	return frame
}
