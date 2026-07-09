package tools

import "os"

const DevProfile = "DEV"

const (
	DatabaseHostNameEnvVar    = "DATABASE_HOST_NAME"
	RedirectHttpToHttpsEnvVar = "REDIRECT_HTTP_TO_HTTPS"
	AppForwardedProtoEnvVar   = "APP_FORWARDED_PROTO"
)

const (
	AppForwardedProtoHttps = "https"
	AppForwardedProtoHttp  = "http"
)

func NewGlobalConfigFromEnv() *GlobalConfig {
	config := newProdGlobalConfig()
	if os.Getenv("PROFILE") == DevProfile {
		config = newDevGlobalConfig()
	}
	databaseHostName := os.Getenv(DatabaseHostNameEnvVar)
	if databaseHostName != "" {
		config.DatabaseHostName = databaseHostName
		config.UseExternalDatabase = true
	}
	if os.Getenv(RedirectHttpToHttpsEnvVar) == "false" {
		config.RedirectHttpToHttps = false
	}
	if os.Getenv(AppForwardedProtoEnvVar) == AppForwardedProtoHttp {
		config.AppForwardedProto = AppForwardedProtoHttp
	}
	return config
}

func newDevGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		PrintCommandOutput:                    true,
		ShouldRunMaintenanceAgent:             false,
		DeployOfficialDatabaseWithPortExposed: true,
		RedirectHttpToHttps:                   true,
		AppForwardedProto:                     AppForwardedProtoHttps,
		DatabaseHostName:                      OfficialDatabaseAppContainerName,
		ExposeDevelopmentRoutes:               true,
		ShowUnofficialAppsSearch:              true,
		CreateDatabaseSnapshotOnStartup:       true,
		PruneDockerSystemDuringMaintenance:    false,
		UseDevelopmentLogger:                  true,
		UseLocalAppStoreClient:                true,
		UseLocalTestingAuthorizedKey:          true,
		UseStrictEmailClientStub:              true,
		UseMockWildcardCertificateService:     true,
		UseExternalDatabase:                   false,
		AllowInsecureOidcProviderTls:          true,
	}
}

func newProdGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		PrintCommandOutput:                    false,
		ShouldRunMaintenanceAgent:             true,
		DeployOfficialDatabaseWithPortExposed: false,
		RedirectHttpToHttps:                   true,
		AppForwardedProto:                     AppForwardedProtoHttps,
		DatabaseHostName:                      OfficialDatabaseAppContainerName,
		ExposeDevelopmentRoutes:               false,
		ShowUnofficialAppsSearch:              false,
		CreateDatabaseSnapshotOnStartup:       false,
		PruneDockerSystemDuringMaintenance:    true,
		UseDevelopmentLogger:                  false,
		UseLocalAppStoreClient:                false,
		UseLocalTestingAuthorizedKey:          false,
		UseStrictEmailClientStub:              false,
		UseMockWildcardCertificateService:     false,
		UseExternalDatabase:                   false,
		AllowInsecureOidcProviderTls:          false,
	}
}
