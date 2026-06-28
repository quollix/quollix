package frontend_pages

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"server/tools"

	"github.com/quollix/common/assert"
)

type FrameAssertions struct {
	Frame *FrameType
}

func (a *FrameAssertions) HostEventually(expectedHost string) *FrameType {
	err := tools.EventuallyWithTimeout(browserTimeout, 50*time.Millisecond, func() error {
		currentUrl, err := a.currentPageUrl()
		if err != nil {
			return err
		}
		if currentUrl.Host != expectedHost {
			return fmt.Errorf("expected host %s, got %s", expectedHost, currentUrl.Host)
		}
		return nil
	})
	assert.Nil(a.Frame.T, err)
	return a.Frame
}

func (a *FrameAssertions) PathEventually(expectedPath string) *FrameType {
	err := tools.EventuallyWithTimeout(browserTimeout, 50*time.Millisecond, func() error {
		currentUrl, err := a.currentPageUrl()
		if err != nil {
			return err
		}
		if currentUrl.Path != expectedPath {
			return fmt.Errorf("expected path %s, got %s", expectedPath, currentUrl.Path)
		}
		return nil
	})
	assert.Nil(a.Frame.T, err)
	return a.Frame
}

func (a *FrameAssertions) PageContainsEventually(expectedText string) *FrameType {
	err := tools.EventuallyWithTimeout(browserTimeout, 50*time.Millisecond, func() error {
		body, err := a.Frame.Page.Element("body")
		if err != nil {
			return err
		}
		text, err := body.Text()
		if err != nil {
			return err
		}
		if !strings.Contains(text, expectedText) {
			return fmt.Errorf("expected page text to contain %q", expectedText)
		}
		return nil
	})
	assert.Nil(a.Frame.T, err)
	return a.Frame
}

func (a *FrameAssertions) SnackbarVisibleWithTextEventually(expectedText string) *FrameType {
	return a.SnackbarVisibleWithTextEventuallyWithin(expectedText, defaultTimeout)
}

func (a *FrameAssertions) SnackbarVisibleWithTextEventuallyWithin(expectedText string, timeout time.Duration) *FrameType {
	err := tools.EventuallyWithTimeout(timeout, 50*time.Millisecond, func() error {
		snackbars, findErr := a.Frame.Page.Elements(`.snackbar[data-visible="true"]`)
		if findErr != nil {
			return findErr
		}
		if len(snackbars) == 0 {
			return fmt.Errorf("no visible snackbar yet")
		}

		var visibleTexts []string
		for _, snackbar := range snackbars {
			text, textErr := snackbar.Text()
			if textErr != nil {
				return textErr
			}
			text = strings.TrimSpace(text)
			visibleTexts = append(visibleTexts, text)
			if strings.Contains(text, expectedText) {
				return nil
			}
		}
		return fmt.Errorf("no visible snackbar contains %q, got %q", expectedText, visibleTexts)
	})
	assert.Nil(a.Frame.T, err)
	return a.Frame
}

func (a *FrameAssertions) PagePath(expectedPath string) {
	a.PathEventually(expectedPath)
}

func (a *FrameAssertions) ElementPresent(selector string) *FrameType {
	a.Frame.Browser.WaitForElement(selector)
	return a.Frame
}

func (a *FrameAssertions) ElementNotPresent(selector string) *FrameType {
	exists, _, err := a.Frame.Page.Has(selector)
	assert.Nil(a.Frame.T, err)
	assert.False(a.Frame.T, exists)
	return a.Frame
}

func (a *FrameAssertions) AppOperationStarted() *FrameType {
	err := tools.EventuallyWithTimeout(backupOperationTimeout, 50*time.Millisecond, func() error {
		_, isOngoing, err := a.Frame.Client.Apps.GetCurrentOperations()
		assert.Nil(a.Frame.T, err)
		if !isOngoing {
			return fmt.Errorf("app operation did not start yet")
		}
		return nil
	})
	assert.Nil(a.Frame.T, err)
	return a.Frame
}

func (a *FrameAssertions) AppOperationFinished() *FrameType {
	err := tools.EventuallyWithTimeout(backupOperationTimeout, 50*time.Millisecond, func() error {
		_, isOngoing, err := a.Frame.Client.Apps.GetCurrentOperations()
		assert.Nil(a.Frame.T, err)
		if isOngoing {
			return fmt.Errorf("app operation is still ongoing")
		}
		return nil
	})
	assert.Nil(a.Frame.T, err)
	return a.Frame
}

func (a *FrameAssertions) AppOperationStartedAndFinished() *FrameType {
	a.AppOperationStarted()
	return a.AppOperationFinished()
}

func (a *FrameAssertions) currentPageUrl() (*url.URL, error) {
	info, err := a.Frame.Page.Info()
	if err != nil {
		return nil, err
	}
	return url.Parse(info.URL)
}
