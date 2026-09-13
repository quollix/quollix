package app_store

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"server/apps_basic"
	"server/tools"

	"github.com/quollix/common/quollix/api"
	"github.com/quollix/common/store"
	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

type AppStoreClientLean interface {
	InitializeOnStartup() error
	ReloadLocalApps() error
	SearchForApps(maintainerSearchTerm, appSearchTerm string, searchForUnofficialApps bool) ([]store.AppWithLatestVersion, error)
	ListVersions(userName, appName string) ([]store.LeanVersionDto, error)
	DownloadVersionByID(versionId int) (*store.Version, error)
	DownloadNextVersionForUpdate(userName, appName string, currentVersionCreationTimestamp time.Time) (*store.NextVersionForUpdateResponse, error)
	GetMaintainerPublicKeyRecord(maintainer string) (*store.MaintainerPublicKeyRecord, error)
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
	MaintainerPublicKeyRecords []store.MaintainerPublicKeyRecord
	DirectoryProvider          tools.DirectoryProvider
	Config                     *tools.GlobalConfig
	VersionValidator           validation.VersionValidator
	AppRepository              apps_basic.AppRepository
	AppService                 apps_basic.AppService
	ClientCredentialsGenerator apps_basic.ClientCredentialsGenerator
	AuthHelper                 u.AuthHelper
	AppServiceHelper           apps_basic.AppServiceHelper
	VersionSigningService      store.VersionSigningService
}

type publishedAppDefinition struct {
	AppName string
	Content []byte
}

const (
	officialTestAppName    = "officialapp"
	officialTestAppVersion = "1.0"
)

var officialTestAppCreationTimestamp = time.Date(2021, 1, 4, 0, 0, 0, 0, time.UTC)

func (h *AppStoreClientMock) InitializeOnStartup() error {
	h.Apps = h.Apps[:0]
	h.Versions = h.Versions[:0]
	h.MaintainerPublicKeyRecords = h.MaintainerPublicKeyRecords[:0]
	return h.InitializeSampleApp()
}

func (h *AppStoreClientMock) ReloadLocalApps() error {
	return h.InitializeApps()
}

func (h *AppStoreClientMock) InitializeApps() error {
	h.Apps = h.Apps[:0]
	h.Versions = h.Versions[:0]
	h.MaintainerPublicKeyRecords = h.MaintainerPublicKeyRecords[:0]
	if err := h.InitializeSampleApp(); err != nil {
		return err
	}
	if err := h.initializePublishedApps(); err != nil {
		return err
	}
	return nil
}

func (h *AppStoreClientMock) InitializeSampleApp() error {
	if err := h.addMaintainerPublicKeyRecord(tools.SampleMaintainer, u.GetOtherLocalTestingPublicKeyRaw()); err != nil {
		return err
	}
	appVersion0Content := []byte(tools.SampleAppVersion0ComposeYAML)
	if _, err := h.addVersion(tools.SampleApp, tools.SampleAppVersion0Name, appVersion0Content, tools.SampleAppVersion0CreationTimestamp); err != nil {
		return err
	}

	appVersion1Content := []byte(tools.SampleAppVersion1ComposeYAML)
	if _, err := h.addVersion(tools.SampleApp, tools.SampleAppVersion1Name, appVersion1Content, tools.SampleAppCreationTimestamp.Add(-time.Hour)); err != nil {
		return err
	}

	appVersion2Content := []byte(tools.SampleAppVersion2ComposeYAML)
	appVersion2, err := h.addVersion(tools.SampleApp, tools.SampleAppVersion2Name, appVersion2Content, tools.SampleAppCreationTimestamp.Add(+time.Hour))
	if err != nil {
		return err
	}
	h.Versions[len(h.Versions)-1].IsMigrationCheckpoint = true
	if _, err := h.addInvalidSignedVersion(tools.SampleMaintainer, tools.SampleApp, "1.5", appVersion2Content, tools.SampleAppCreationTimestamp); err != nil {
		return err
	}

	postgresVersion17Content := []byte(tools.SamplePostgresAppVersion17ComposeYAML)
	if _, err := h.addVersion(tools.SamplePostgresApp, tools.SamplePostgresAppVersion17Name, postgresVersion17Content, tools.SamplePostgresAppVersion17CreationTimestamp); err != nil {
		return err
	}

	postgresVersion18Content := []byte(tools.SamplePostgresAppVersion18ComposeYAML)
	postgresVersion18, err := h.addVersion(tools.SamplePostgresApp, tools.SamplePostgresAppVersion18Name, postgresVersion18Content, tools.SamplePostgresAppVersion18CreationTimestamp)
	if err != nil {
		return err
	}

	rabbitMQVersion311Content := []byte(tools.SampleRabbitMQAppVersion311ComposeYAML)
	if _, err := h.addVersion(tools.SampleRabbitMQApp, tools.SampleRabbitMQAppVersion311Name, rabbitMQVersion311Content, tools.SampleRabbitMQAppVersion311CreationTimestamp); err != nil {
		return err
	}

	rabbitMQVersion312Content := []byte(tools.SampleRabbitMQAppVersion312ComposeYAML)
	rabbitMQVersion312, err := h.addVersion(tools.SampleRabbitMQApp, tools.SampleRabbitMQAppVersion312Name, rabbitMQVersion312Content, tools.SampleRabbitMQAppVersion312CreationTimestamp)
	if err != nil {
		return err
	}

	h.Apps = append(h.Apps, store.AppWithLatestVersion{
		Maintainer:                     tools.SampleMaintainer,
		AppName:                        tools.SampleApp,
		LatestVersionId:                appVersion2.VersionId,
		LatestVersionName:              tools.SampleAppVersion2Name,
		LatestVersionCreationTimestamp: tools.SampleAppVersion2CreationTimestamp,
	})
	h.Apps = append(h.Apps, store.AppWithLatestVersion{
		Maintainer:                     tools.SampleMaintainer,
		AppName:                        tools.SamplePostgresApp,
		LatestVersionId:                postgresVersion18.VersionId,
		LatestVersionName:              tools.SamplePostgresAppVersion18Name,
		LatestVersionCreationTimestamp: tools.SamplePostgresAppVersion18CreationTimestamp,
	})
	h.Apps = append(h.Apps, store.AppWithLatestVersion{
		Maintainer:                     tools.SampleMaintainer,
		AppName:                        tools.SampleRabbitMQApp,
		LatestVersionId:                rabbitMQVersion312.VersionId,
		LatestVersionName:              tools.SampleRabbitMQAppVersion312Name,
		LatestVersionCreationTimestamp: tools.SampleRabbitMQAppVersion312CreationTimestamp,
	})

	officialAppContent := []byte(`services:
  officialapp:
    image: officialapp:local
    container_name: quollix_officialapp_officialapp
    labels:
      quollix.port: 8080
`)
	officialAppVersion, err := h.addOfficialVersion(officialTestAppName, officialTestAppVersion, officialAppContent, officialTestAppCreationTimestamp)
	if err != nil {
		return err
	}
	h.Apps = append(h.Apps, store.AppWithLatestVersion{
		Maintainer:                     u.OfficialMaintainer,
		AppName:                        officialTestAppName,
		LatestVersionId:                officialAppVersion.VersionId,
		LatestVersionName:              officialTestAppVersion,
		LatestVersionCreationTimestamp: officialTestAppCreationTimestamp,
	})

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
	appSecret, err := h.AuthHelper.GenerateSecret()
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
		api.Policies.AdminOnlyAccessPolicy,
		port,
		clientId,
		clientSecret,
		appSecret,
		time.Now(),
		appContent,
		false,
		false,
		false,
	)

	return h.AppService.UpsertAppInDatabase(newApp)
}

func (h *AppStoreClientMock) updateExistingTestApp(appName string, appContent []byte) error {
	app, err := h.AppRepository.GetAppByName(appName)
	if err != nil {
		return err
	}
	app.VersionContent = appContent
	return h.AppService.UpsertAppInDatabase(app)
}

func (h *AppStoreClientMock) ListVersions(userName, appName string) ([]store.LeanVersionDto, error) {
	versions := make([]store.LeanVersionDto, 0)
	for _, version := range h.Versions {
		if version.Maintainer != userName || version.AppName != appName {
			continue
		}
		versions = append(versions, store.LeanVersionDto{
			VersionId:             version.VersionId,
			Name:                  version.VersionName,
			CreationTimestamp:     version.VersionCreationTimestamp,
			SizeInBytes:           int64(len(version.Content)),
			IsMigrationCheckpoint: version.IsMigrationCheckpoint,
		})
	}
	return versions, nil
}

func (h *AppStoreClientMock) DownloadVersionByID(versionId int) (*store.Version, error) {
	for index := range h.Versions {
		version := &h.Versions[index]
		if version.VersionId == versionId {
			return version, nil
		}
	}
	return nil, u.Logger.NewError("version not found", tools.VersionIdField, versionId)
}

func (h *AppStoreClientMock) DownloadNextVersionForUpdate(userName, appName string, currentVersionCreationTimestamp time.Time) (*store.NextVersionForUpdateResponse, error) {
	var latest *store.Version
	for index := range h.Versions {
		version := &h.Versions[index]
		if version.Maintainer != userName || version.AppName != appName {
			continue
		}
		if !version.VersionCreationTimestamp.After(currentVersionCreationTimestamp) {
			continue
		}
		if latest == nil || version.VersionCreationTimestamp.After(latest.VersionCreationTimestamp) {
			latest = version
		}
	}
	if latest == nil {
		return &store.NextVersionForUpdateResponse{UpdateAvailable: false}, nil
	}
	return &store.NextVersionForUpdateResponse{UpdateAvailable: true, Version: latest}, nil
}

func (h *AppStoreClientMock) GetMaintainerPublicKeyRecord(maintainer string) (*store.MaintainerPublicKeyRecord, error) {
	for index := range h.MaintainerPublicKeyRecords {
		record := &h.MaintainerPublicKeyRecords[index]
		if record.Maintainer == maintainer {
			return record, nil
		}
	}
	return nil, u.Logger.NewError("maintainer not found", tools.MaintainerField, maintainer)
}

func (h *AppStoreClientMock) addVersion(appName, versionName string, content []byte, versionCreationTimestamp time.Time) (*store.Version, error) {
	return h.addVersionForMaintainer(tools.SampleMaintainer, []byte(u.OtherLocalTestingPrivateKeyOpenSSH), []byte(u.OtherLocalTestingPrivateKeyPassphrase), appName, versionName, content, versionCreationTimestamp)
}

func (h *AppStoreClientMock) addOfficialVersion(appName, versionName string, content []byte, versionCreationTimestamp time.Time) (*store.Version, error) {
	return h.addVersionForMaintainer(u.OfficialMaintainer, []byte(u.LocalTestingPrivateKeyOpenSSH), []byte(u.LocalTestingPrivateKeyPassphrase), appName, versionName, content, versionCreationTimestamp)
}

func (h *AppStoreClientMock) addVersionForMaintainer(maintainer string, privateKeyOpenSSH, privateKeyPassphrase []byte, appName, versionName string, content []byte, versionCreationTimestamp time.Time) (*store.Version, error) {
	privateKey, err := u.DecodeEd25519PrivateKeyOpenSSH(privateKeyOpenSSH, privateKeyPassphrase)
	if err != nil {
		return nil, err
	}
	version := &store.Version{
		VersionId:                h.nextVersionID(),
		Maintainer:               maintainer,
		AppName:                  appName,
		VersionName:              versionName,
		Content:                  content,
		VersionCreationTimestamp: versionCreationTimestamp,
		MaintainerPublicKeyRaw:   privateKey.Public().(ed25519.PublicKey),
	}

	signature, err := h.VersionSigningService.SignVersion(privateKey, version)
	if err != nil {
		return nil, err
	}
	version.Signature = signature

	h.Versions = append(h.Versions, *version)
	return version, nil
}

func (h *AppStoreClientMock) addInvalidSignedVersion(maintainer, appName, versionName string, content []byte, versionCreationTimestamp time.Time) (*store.Version, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}

	version := &store.Version{
		VersionId:                h.nextVersionID(),
		Maintainer:               maintainer,
		AppName:                  appName,
		VersionName:              versionName,
		Content:                  content,
		VersionCreationTimestamp: versionCreationTimestamp,
		MaintainerPublicKeyRaw:   publicKey,
	}
	signature, err := h.VersionSigningService.SignVersion(privateKey, version)
	if err != nil {
		return nil, err
	}
	version.Signature = signature

	h.Versions = append(h.Versions, *version)
	return version, nil
}

func (h *AppStoreClientMock) addMaintainerPublicKeyRecord(maintainer string, publicKeyRaw []byte) error {
	officialPrivateKey, err := decodeTestingPrivateKey()
	if err != nil {
		return err
	}
	publicKeySignature, err := store.SignMaintainerPublicKey(officialPrivateKey, maintainer, publicKeyRaw)
	if err != nil {
		return err
	}
	h.MaintainerPublicKeyRecords = append(h.MaintainerPublicKeyRecords, store.MaintainerPublicKeyRecord{
		Maintainer:         maintainer,
		PublicKeyRaw:       publicKeyRaw,
		PublicKeySignature: publicKeySignature,
	})
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

func (h *AppStoreClientMock) nextVersionID() int {
	return len(h.Versions) + 1
}

func decodeTestingPrivateKey() (ed25519.PrivateKey, error) {
	return u.DecodeEd25519PrivateKeyOpenSSH([]byte(u.LocalTestingPrivateKeyOpenSSH), []byte(u.LocalTestingPrivateKeyPassphrase))
}
