package tools

import "os"

const DevProfile = "DEV"

func NewGlobalConfigFromEnv() *GlobalConfig {
	if os.Getenv("PROFILE") == DevProfile {
		return newDevGlobalConfig()
	}
	return newProdGlobalConfig()
}

func newDevGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		OpenTestStateResetEndpoint:            true,
		PrintCommandOutput:                    true,
		ShouldRunMaintenanceAgent:             false,
		DeployOfficialDatabaseWithPortExposed: true,
		DatabaseHostName:                      OfficialDatabaseAppContainerName,
		ShowReloadFrontendTemplatesButton:     true,
		ShowUnofficialAppsSearch:              true,
		CreateDatabaseSnapshotOnStartup:       true,
		PruneDockerSystemDuringMaintenance:    false,
		UseDevelopmentLogger:                  true,
		UseLocalAppStoreClient:                true,
		UseLocalTestingAuthorizedKey:          true,
		UseStrictEmailClientStub:              true,
		UseMockWildcardCertificateService:     true,
	}
}

func newProdGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		OpenTestStateResetEndpoint:            false,
		PrintCommandOutput:                    false,
		ShouldRunMaintenanceAgent:             true,
		DeployOfficialDatabaseWithPortExposed: false,
		DatabaseHostName:                      OfficialDatabaseAppContainerName,
		ShowReloadFrontendTemplatesButton:     false,
		ShowUnofficialAppsSearch:              false,
		CreateDatabaseSnapshotOnStartup:       false,
		PruneDockerSystemDuringMaintenance:    true,
		UseDevelopmentLogger:                  false,
		UseLocalAppStoreClient:                false,
		UseLocalTestingAuthorizedKey:          false,
		UseStrictEmailClientStub:              false,
		UseMockWildcardCertificateService:     false,
	}
}
