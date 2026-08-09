//go:build frontend

package frontend

import (
	"testing"

	"server/tests/component"
	"server/tests/frontend_pages"
	"server/tools"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

func TestAppSecretsPage(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	_, err := component.InstallSample(t, frame.Client, "2.0")
	assert.Nil(t, err)

	oldSampleApp := component.GetInstalledSample(t, frame.Client)
	oldSharedSecret := oldSampleApp.Secrets["SECRET_SAMPLE_SHARED"]
	assert.Equal(t, 64, len(oldSharedSecret))

	page := frame.Pages.OpenAppsWithSecretsPage()

	page.
		AssertAppCount(1).
		AssertAppPresent(tools.SampleMaintainer, tools.SampleApp, "2.0")

	detailPage := page.OpenApp(tools.SampleApp)
	detailPage.AssertMetadata(tools.SampleMaintainer, tools.SampleApp, "2.0")
	detailPage.
		AssertSecretCount(3).
		AssertSecretMasked("SECRET_SAMPLE_SHARED").
		AssertSecretMasked("SECRET_SAMPLE_MIGRATED_PASSWORD").
		AssertSecretMasked("SECRET_SAMPLE_VERSION_TWO")

	detailPage.RegenerateSecret("SECRET_SAMPLE_SHARED")

	err = u.Eventually(func() error {
		newSampleApp := component.GetInstalledSample(t, frame.Client)
		newSharedSecret := newSampleApp.Secrets["SECRET_SAMPLE_SHARED"]
		if newSharedSecret == oldSharedSecret {
			return u.Logger.NewError("app secret was not regenerated")
		}
		if len(newSharedSecret) != 64 {
			return u.Logger.NewError("app secret has unexpected length", "actual_length", len(newSharedSecret))
		}
		return nil
	})
	assert.Nil(t, err)

	detailPage.AssertSecretMasked("SECRET_SAMPLE_SHARED")
}
