package frontend_pages

import (
	"strings"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/browsertest"
	u "github.com/quollix/common/utils"
)

// Only retry GET navigation, never the action that triggered it.
func navigateBackupPage(frame *FrameType, destination, loadedSelector string, action func() error) {
	page := frame.Page.Timeout(backupOperationTimeout)
	defer page.CancelTimeout()
	navigationErr := action()
	err := u.EventuallyWithTimeout(backupOperationTimeout, 100*time.Millisecond, func() error {
		if navigationErr != nil {
			if !browsertest.IsNetworkChangedError(navigationErr) {
				frame.T.Fatal(navigationErr)
			}
			navigationErr = page.Navigate(destination)
			if navigationErr != nil {
				return navigationErr
			}
		}
		info, err := page.Info()
		if err != nil {
			return err
		}
		isErrorPage, _, err := page.Has("#main-frame-error")
		if err != nil {
			return err
		}
		if isErrorPage {
			body, err := page.BodyText()
			if err != nil {
				return err
			}
			if !strings.Contains(body, "ERR_NETWORK_CHANGED") {
				frame.T.Fatalf("Unexpected browser error at %s: %s", info.URL, body)
			}
			navigationErr = page.Navigate(destination)
			return u.Logger.NewError("retrying backup page after network change")
		}
		loaded, _, err := page.Has(loadedSelector)
		if err != nil {
			return err
		}
		if info.URL != destination || !loaded {
			return u.Logger.NewError("backup page is not ready", "expected_url", destination, "actual_url", info.URL)
		}
		return nil
	})
	assert.Nil(frame.T, err)
}
