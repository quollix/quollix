package tools

import (
	"os"
)

const DevProfile = "DEV"
const MockBlogPagePath = "/test-blog"

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
		CollectMetricsAutomatically:           false,
		ForceFlushMetricsAfterPush:            true,
		PrintCommandOutput:                    true,
		ShouldRunMaintenanceAgent:             false,
		DeployOfficialDatabaseWithPortExposed: true,
		RedirectHttpToHttps:                   true,
		AppForwardedProto:                     AppForwardedProtoHttps,
		DatabaseHostName:                      OfficialDatabaseAppContainerName,
		ExposeDevelopmentRoutes:               true,
		CreateDatabaseSnapshotOnStartup:       true,
		PruneDockerSystemDuringMaintenance:    false,
		UseDevelopmentLogger:                  true,
		UseLocalAppStoreClient:                true,
		UseLocalTestingAuthorizedKey:          true,
		UseStrictEmailClientStub:              true,
		UseMockWildcardCertificateService:     true,
		MockRSSClient:                         true,
		BlogURL:                               "https://quollix.localhost" + MockBlogPagePath,
		UseExternalDatabase:                   false,
		AllowInsecureOidcProviderTls:          true,
	}
}

func newProdGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		CollectMetricsAutomatically:           true,
		ForceFlushMetricsAfterPush:            false,
		PrintCommandOutput:                    false,
		ShouldRunMaintenanceAgent:             true,
		DeployOfficialDatabaseWithPortExposed: false,
		RedirectHttpToHttps:                   true,
		AppForwardedProto:                     AppForwardedProtoHttps,
		DatabaseHostName:                      OfficialDatabaseAppContainerName,
		ExposeDevelopmentRoutes:               false,
		CreateDatabaseSnapshotOnStartup:       false,
		PruneDockerSystemDuringMaintenance:    true,
		UseDevelopmentLogger:                  false,
		UseLocalAppStoreClient:                false,
		UseLocalTestingAuthorizedKey:          false,
		UseStrictEmailClientStub:              false,
		UseMockWildcardCertificateService:     false,
		MockRSSClient:                         false,
		BlogURL:                               "https://quollix.org/blog/",
		UseExternalDatabase:                   false,
		AllowInsecureOidcProviderTls:          false,
	}
}
