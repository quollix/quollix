package app_store

import (
	"errors"
	"server/apps_basic"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/store"
)

var testDownloadedVersion = &store.Version{
	Maintainer:               "samplemaintainer",
	AppName:                  "sampleapp",
	VersionName:              "1.0",
	Content:                  []byte("services:\n  sampleapp:\n    labels:\n      quollix.port: \"8080\"\n"),
	MaintainerPublicKeyRaw:   []byte("public-key"),
	Signature:                []byte("signature"),
	VersionCreationTimestamp: time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC),
}

type appStoreServiceTestDependencies struct {
	service                    *AppStoreServiceImpl
	appStoreClient             *AppStoreClientLeanMock
	versionValidator           *VersionValidatorMock
	versionVerifier            *VersionVerifierMock
	clientCredentialsGenerator *apps_basic.ClientCredentialsGeneratorMock
	appServiceHelper           *apps_basic.AppServiceHelperMock
}

func setupAppStoreServiceTestDependencies(t *testing.T) appStoreServiceTestDependencies {
	appStoreClient := NewAppStoreClientLeanMock(t)
	versionValidator := NewVersionValidatorMock(t)
	versionVerifier := NewVersionVerifierMock(t)
	clientCredentialsGenerator := apps_basic.NewClientCredentialsGeneratorMock(t)
	appServiceHelper := apps_basic.NewAppServiceHelperMock(t)

	service := &AppStoreServiceImpl{
		AppStoreClientLean:         appStoreClient,
		ClientCredentialsGenerator: clientCredentialsGenerator,
		AppServiceHelper:           appServiceHelper,
		VersionValidator:           versionValidator,
		VersionVerifier:            versionVerifier,
	}

	return appStoreServiceTestDependencies{
		service:                    service,
		appStoreClient:             appStoreClient,
		versionValidator:           versionValidator,
		versionVerifier:            versionVerifier,
		clientCredentialsGenerator: clientCredentialsGenerator,
		appServiceHelper:           appServiceHelper,
	}
}

func TestDownloadVersion_ValidationFailsBeforeTrustedKeyVerification(t *testing.T) {
	testDependencies := setupAppStoreServiceTestDependencies(t)
	validatorErr := errors.New("invalid downloaded version")

	testDependencies.appStoreClient.EXPECT().
		DownloadVersion(testDownloadedVersion.Maintainer, testDownloadedVersion.AppName, testDownloadedVersion.VersionName).
		Return(testDownloadedVersion, nil)
	testDependencies.versionValidator.EXPECT().
		Validate(testDownloadedVersion.Content, testDownloadedVersion.Maintainer, testDownloadedVersion.AppName).
		Return(validatorErr)

	repoApp, err := testDependencies.service.DownloadVersion(testDownloadedVersion.Maintainer, testDownloadedVersion.AppName, testDownloadedVersion.VersionName)

	assert.Nil(t, repoApp)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, validatorErr))
	assert.Equal(t, "version validation failed: invalid downloaded version", err.Error())
}

func TestDownloadVersion_CreatesRepoAppAfterValidation(t *testing.T) {
	testDependencies := setupAppStoreServiceTestDependencies(t)

	testDependencies.appStoreClient.EXPECT().
		DownloadVersion(testDownloadedVersion.Maintainer, testDownloadedVersion.AppName, testDownloadedVersion.VersionName).
		Return(testDownloadedVersion, nil)
	testDependencies.versionValidator.EXPECT().
		Validate(testDownloadedVersion.Content, testDownloadedVersion.Maintainer, testDownloadedVersion.AppName).
		Return(nil)
	testDependencies.clientCredentialsGenerator.EXPECT().Generate().Return("client-id", "client-secret", nil)
	testDependencies.appServiceHelper.EXPECT().GetPortFromComposeYaml(testDownloadedVersion.Content, testDownloadedVersion.AppName).Return("8080", nil)
	testDependencies.versionVerifier.EXPECT().Verify(testDownloadedVersion).Return(nil)

	repoApp, err := testDependencies.service.DownloadVersion(testDownloadedVersion.Maintainer, testDownloadedVersion.AppName, testDownloadedVersion.VersionName)

	assert.Nil(t, err)
	assert.NotNil(t, repoApp)
	assert.Equal(t, testDownloadedVersion.Maintainer, repoApp.Maintainer)
	assert.Equal(t, testDownloadedVersion.AppName, repoApp.AppName)
	assert.Equal(t, testDownloadedVersion.VersionName, repoApp.VersionName)
	assert.Equal(t, "8080", repoApp.Port)
	assert.Equal(t, "client-id", repoApp.ClientId)
	assert.Equal(t, "client-secret", repoApp.ClientSecret)
	assert.Equal(t, testDownloadedVersion.Content, repoApp.VersionContent)
}
