package frontend_pages

import (
	"testing"
	"time"

	"server/tests/api_client"

	"github.com/go-rod/rod"
)

const defaultTimeout = 3 * time.Second
const backupOperationTimeout = 1 * time.Minute

type FrameType struct {
	T        *testing.T
	BaseUrl  string
	Page     *rod.Page
	Client   *api_client.QuollixClient
	Controls *FrameControls
	Pages    *FramePages
	Assert   *FrameAssertions
	Session  *FrameSession
	Browser  *FrameBrowser
}

func NewFrameType(t *testing.T, baseUrl string, page *rod.Page, client *api_client.QuollixClient) *FrameType {
	frame := &FrameType{
		T:       t,
		BaseUrl: baseUrl,
		Page:    page,
		Client:  client,
	}
	frame.Controls = &FrameControls{Frame: frame}
	frame.Pages = newFramePages(frame)
	frame.Assert = &FrameAssertions{Frame: frame}
	frame.Session = &FrameSession{Frame: frame}
	frame.Browser = &FrameBrowser{Frame: frame}
	return frame
}
