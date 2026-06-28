package app_store

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"server/apps_basic"
	"server/tools"
	"strings"
	"time"

	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

type AppStoreClientLean interface {
	InitializeOnStartup() error
	ReloadLocalApps() error
	SearchForApps(maintainerSearchTerm, appSearchTerm string, searchForUnofficialApps bool) ([]store.AppWithLatestVersion, error)
	ListVersions(userName, appName string) ([]store.LeanVersionDto, error)
	DownloadVersion(userName, appName, versionName string) (*store.Version, error)
}

type AppStoreClientImpl struct {
	store.AppStoreClientImpl
}

func (h *AppStoreClientImpl) InitializeOnStartup() error {
	return nil
}

func (h *AppStoreClientImpl) ReloadLocalApps() error {
	return nil
}

type AppStoreClientMock struct {
	Apps                       []store.AppWithLatestVersion
	Versions                   []store.Version
	DirectoryProvider          tools.DirectoryProvider
	Config                     *tools.GlobalConfig
	VersionValidator           validation.VersionValidator
	AppRepository              apps_basic.AppRepository
	ClientCredentialsGenerator apps_basic.ClientCredentialsGenerator
	AppServiceHelper           apps_basic.AppServiceHelper
	VersionSigningService      store.VersionSigningService
}

type publishedAppDefinition struct {
	AppName string
	Content []byte
}

func (h *AppStoreClientMock) InitializeOnStartup() error {
	return h.InitializeSampleApp()
}

func (h *AppStoreClientMock) ReloadLocalApps() error {
	return h.InitializeApps()
}

func (h *AppStoreClientMock) InitializeApps() error {
	h.Apps = h.Apps[:0]
	h.Versions = h.Versions[:0]
	if err := h.InitializeSampleApp(); err != nil {
		return err
	}
	if err := h.initializePublishedApps(); err != nil {
		return err
	}
	return nil
}

func (h *AppStoreClientMock) InitializeSampleApp() error {
	h.Apps = append(h.Apps, store.AppWithLatestVersion{
		Maintainer:                     tools.SampleMaintainer,
		AppName:                        tools.SampleApp,
		LatestVersionName:              tools.SampleAppVersion2Name,
		LatestVersionCreationTimestamp: tools.SampleAppVersion2CreationTimestamp,
	})

	appVersion0Content := []byte(tools.SampleAppVersion0ComposeYAML)
	if err := h.addVersion(tools.SampleMaintainer, tools.SampleApp, tools.SampleAppVersion0Name, appVersion0Content, tools.SampleAppVersion0CreationTimestamp); err != nil {
		return err
	}

	appVersion1Content := []byte(tools.SampleAppVersion1ComposeYAML)
	if err := h.addVersion(tools.SampleMaintainer, tools.SampleApp, tools.SampleAppVersion1Name, appVersion1Content, tools.SampleAppCreationTimestamp.Add(-time.Hour)); err != nil {
		return err
	}

	appVersion2Content := []byte(tools.SampleAppVersion2ComposeYAML)
	if err := h.addVersion(tools.SampleMaintainer, tools.SampleApp, tools.SampleAppVersion2Name, appVersion2Content, tools.SampleAppCreationTimestamp.Add(+time.Hour)); err != nil {
		return err
	}
	if err := h.addInvalidSignedVersion(tools.SampleMaintainer, tools.SampleApp, "1.5", appVersion2Content, tools.SampleAppCreationTimestamp); err != nil {
		return err
	}

	return nil
}

func (h *AppStoreClientMock) initializePublishedApps() error {
	publishedAppsDir := h.DirectoryProvider.GetPublishedAppsDir()
	appDefinitions, err := loadPublishedAppDefinitions(publishedAppsDir)
	if err != nil {
		return err
	}
	testMaintainer := "quollix"
	testVersionName := "1.0"

	for _, appDefinition := range appDefinitions {
		if err := h.VersionValidator.Validate(appDefinition.Content, testMaintainer, appDefinition.AppName); err != nil {
			return err
		}

		exists, existsErr := h.AppRepository.DoesAppExist(appDefinition.AppName)
		if existsErr != nil {
			return existsErr
		}

		if exists {
			if err := h.updateExistingTestApp(appDefinition.AppName, appDefinition.Content); err != nil {
				return err
			}
		} else {
			if err := h.createNewTestApp(testMaintainer, appDefinition.AppName, testVersionName, appDefinition.Content); err != nil {
				return err
			}
		}
	}

	return nil
}

func loadPublishedAppDefinitions(publishedAppsDir string) ([]publishedAppDefinition, error) {
	entries, err := os.ReadDir(publishedAppsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, u.Logger.NewError(PublishedAppsDirectoryDoesNotExistError, tools.SourcePathField, publishedAppsDir)
		}
		return nil, u.Logger.NewError(err.Error(), tools.SourcePathField, publishedAppsDir)
	}

	appDefinitions := make([]publishedAppDefinition, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yml" {
			continue
		}

		appComposePath := filepath.Join(publishedAppsDir, entry.Name())
		appContent, appContentErr := loadDockerComposeYaml(appComposePath)
		if appContentErr != nil {
			return nil, appContentErr
		}

		appDefinitions = append(appDefinitions, publishedAppDefinition{
			AppName: strings.TrimSuffix(entry.Name(), ".yml"),
			Content: appContent,
		})
	}
	if len(appDefinitions) == 0 {
		return nil, u.Logger.NewError(PublishedAppsDirectoryIsEmptyError, tools.SourcePathField, publishedAppsDir)
	}

	return appDefinitions, nil
}

func loadDockerComposeYaml(appComposePath string) ([]byte, error) {
	appContent, err := os.ReadFile(appComposePath) // #nosec G304 (CWE-22): Potential file inclusion via variable
	if err != nil {
		return nil, u.Logger.NewError(err.Error(), tools.SourcePathField, appComposePath)
	}
	return appContent, nil
}

func (h *AppStoreClientMock) createNewTestApp(
	testMaintainer string,
	appName string,
	testVersionName string,
	appContent []byte,
) error {
	clientId, clientSecret, err := h.ClientCredentialsGenerator.Generate()
	if err != nil {
		return err
	}

	port, err := h.AppServiceHelper.GetPortFromComposeYaml(appContent, appName)
	if err != nil {
		return err
	}

	newApp := apps_basic.NewRepoApp(
		testMaintainer,
		appName,
		testVersionName,
		tools.Policies.AdminOnlyAccessPolicy,
		port,
		clientId,
		clientSecret,
		time.Now(),
		appContent,
		false,
		false,
		false,
	)

	_, err = h.AppRepository.CreateApp(newApp)
	return err
}

func (h *AppStoreClientMock) updateExistingTestApp(appName string, appContent []byte) error {
	app, err := h.AppRepository.GetAppByName(appName)
	if err != nil {
		return err
	}
	app.VersionContent = appContent
	return h.AppRepository.UpdateApp(app)
}

func (h *AppStoreClientMock) ListVersions(userName, appName string) ([]store.LeanVersionDto, error) {
	versions := make([]store.LeanVersionDto, 0)
	for _, version := range h.Versions {
		if version.Maintainer != userName || version.AppName != appName {
			continue
		}
		versions = append(versions, store.LeanVersionDto{
			Name:              version.VersionName,
			CreationTimestamp: version.VersionCreationTimestamp,
			SizeInBytes:       int64(len(version.Content)),
		})
	}
	return versions, nil
}

func (h *AppStoreClientMock) DownloadVersion(userName, appName, versionName string) (*store.Version, error) {
	for index := range h.Versions {
		version := &h.Versions[index]
		if version.Maintainer == userName && version.AppName == appName && version.VersionName == versionName {
			return version, nil
		}
	}
	return nil, u.Logger.NewError("version not found")
}

func (h *AppStoreClientMock) addVersion(maintainer, appName, versionName string, content []byte, versionCreationTimestamp time.Time) error {
	privateKey, err := decodeTestingPrivateKey()
	if err != nil {
		return err
	}
	version := &store.Version{
		Maintainer:               maintainer,
		AppName:                  appName,
		VersionName:              versionName,
		Content:                  content,
		VersionCreationTimestamp: versionCreationTimestamp,
		MaintainerPublicKeyRaw:   privateKey.Public().(ed25519.PublicKey),
	}

	signature, err := h.VersionSigningService.SignVersion(privateKey, version)
	if err != nil {
		return err
	}
	version.Signature = signature

	h.Versions = append(h.Versions, *version)
	return nil
}

func (h *AppStoreClientMock) addInvalidSignedVersion(maintainer, appName, versionName string, content []byte, versionCreationTimestamp time.Time) error {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}

	version := &store.Version{
		Maintainer:               maintainer,
		AppName:                  appName,
		VersionName:              versionName,
		Content:                  content,
		VersionCreationTimestamp: versionCreationTimestamp,
		MaintainerPublicKeyRaw:   publicKey,
	}
	signature, err := h.VersionSigningService.SignVersion(privateKey, version)
	if err != nil {
		return err
	}
	version.Signature = signature

	h.Versions = append(h.Versions, *version)
	return nil
}

func (h *AppStoreClientMock) SearchForApps(maintainerSearchTerm string, appSearchTerm string, showUnofficialApps bool) ([]store.AppWithLatestVersion, error) {
	results := make([]store.AppWithLatestVersion, 0)
	for _, app := range h.Apps {
		isOfficialApp := app.Maintainer == u.OfficialMaintainer
		if !showUnofficialApps && !isOfficialApp {
			continue
		}
		if maintainerSearchTerm != "" && !strings.Contains(app.Maintainer, maintainerSearchTerm) {
			continue
		}
		if appSearchTerm != "" && !strings.Contains(app.AppName, appSearchTerm) {
			continue
		}
		results = append(results, app)
	}
	return results, nil
}

func decodeTestingPrivateKey() (ed25519.PrivateKey, error) {
	return u.DecodeEd25519PrivateKeyOpenSSH(u.GetLocalTestingPrivateKeyBytes())
}
