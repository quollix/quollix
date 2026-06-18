package tools

import (
	"time"

	u "github.com/quollix/common/utils"
)

const (
	BrandAppDomainPrefix = u.OfficialBrandAppName + "."

	CookieExpirationTimeInDays = 7
	CookieExpirationTime       = CookieExpirationTimeInDays * 24 * time.Hour
	TempDir                    = "/tmp"
	BrandAppAuthCookieName     = u.BrandName + "-auth"
	BrandAppQuerySecretName    = u.BrandName + "-secret"

	DefaultAdminName     = "admin"
	DefaultAdminPassword = "password"
	DefaultAdminEmail    = "admin@example.invalid"

	ResticImageName                 = "restic:local"
	ResticCleanupLabel              = "quollix.cleanup=restic"
	ResticImageMaintenanceKeepLabel = "quollix.maintenance.keep=restic"

	OfficialDatabaseAppMaintainerAndAppName = u.OfficialMaintainer + "_" + u.OfficialDatabaseAppName
	OfficialDatabaseAppNetworkName          = OfficialDatabaseAppMaintainerAndAppName
	OfficialDatabaseAppContainerName        = OfficialDatabaseAppMaintainerAndAppName + "_" + u.OfficialDatabaseAppName

	BrandAppService       = u.BrandName
	BrandAppContainerName = u.OfficialMaintainer + "_" + u.OfficialBrandAppName + "_" + BrandAppService

	QuollogAppName = "quollog"
)

var ComposeEnvVars = struct {
	ServerHost   string
	ClientId     string
	ClientSecret string

	IanaTimeZone string
}{
	ServerHost:   "SERVER_HOST",
	ClientId:     "CLIENT_ID",
	ClientSecret: "CLIENT_SECRET",
	IanaTimeZone: "IANA_TIMEZONE",
}

type GlobalConfig struct {
	OpenTestStateResetEndpoint            bool
	PrintCommandOutput                    bool
	ShouldRunMaintenanceAgent             bool
	DeployOfficialDatabaseWithPortExposed bool
	DatabaseHostName                      string
	ShowReloadFrontendTemplatesButton     bool
	ShowUnofficialAppsSearch              bool
	CreateDatabaseSnapshotOnStartup       bool
	PruneDockerSystemDuringMaintenance    bool
	UseDevelopmentLogger                  bool
	UseLocalAppStoreClient                bool
	UseLocalTestingAuthorizedKey          bool
	UseStrictEmailClientStub              bool
	UseMockWildcardCertificateService     bool
}
