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
		AssertSecretVisibility("SECRET_SAMPLE_SHARED", false).
		AssertSecretValue("SECRET_SAMPLE_SHARED", oldSharedSecret).
		ToggleSecretVisibility("SECRET_SAMPLE_SHARED").
		AssertSecretVisibility("SECRET_SAMPLE_SHARED", true).
		AssertSecretValue("SECRET_SAMPLE_SHARED", oldSharedSecret).
		ToggleSecretVisibility("SECRET_SAMPLE_SHARED").
		AssertSecretVisibility("SECRET_SAMPLE_SHARED", false)

	updatedSharedSecret := "frontend-updated-secret"
	detailPage.UpdateSecret("SECRET_SAMPLE_SHARED", updatedSharedSecret)

	err = u.Eventually(func() error {
		newSampleApp := component.GetInstalledSample(t, frame.Client)
		if newSampleApp.Secrets["SECRET_SAMPLE_SHARED"] != updatedSharedSecret {
			return u.Logger.NewError("app secret was not updated")
		}
		return nil
	})
	assert.Nil(t, err)

	detailPage.RegenerateSecret("SECRET_SAMPLE_SHARED")

	err = u.Eventually(func() error {
		newSampleApp := component.GetInstalledSample(t, frame.Client)
		newSharedSecret := newSampleApp.Secrets["SECRET_SAMPLE_SHARED"]
		if newSharedSecret == updatedSharedSecret {
			return u.Logger.NewError("app secret was not regenerated")
		}
		if len(newSharedSecret) != 64 {
			return u.Logger.NewError("app secret has unexpected length", "actual_length", len(newSharedSecret))
		}
		return nil
	})
	assert.Nil(t, err)
}

func TestAppSecretsPageDeletesUnusedSecretOnly(t *testing.T) {
	frame := frontend_pages.Setup(t)
	defer frame.Client.Test.ResetTestState()

	appBeforeUpdate, err := component.InstallSample(t, frame.Client, "1.0")
	assert.Nil(t, err)
	assert.Nil(t, frame.Client.Apps.Update(appBeforeUpdate.AppId))

	page := frame.Pages.OpenAppsWithSecretsPage()
	detailPage := page.OpenApp(tools.SampleApp)
	detailPage.
		AssertSecretDeleteButtonPresent("SECRET_SAMPLE_SHARED", false).
		AssertSecretDeleteButtonPresent("SECRET_SAMPLE_VERSION_TWO", false).
		AssertSecretDeleteButtonPresent("SECRET_SAMPLE_VERSION_ONE", true).
		DeleteSecret("SECRET_SAMPLE_VERSION_ONE")

	err = u.Eventually(func() error {
		newSampleApp := component.GetInstalledSample(t, frame.Client)
		if _, exists := newSampleApp.Secrets["SECRET_SAMPLE_VERSION_ONE"]; exists {
			return u.Logger.NewError("app secret was not deleted")
		}
		return nil
	})
	assert.Nil(t, err)
}
