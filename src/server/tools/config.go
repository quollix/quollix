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

	DefaultAdminName     = "administrator"
	DefaultAdminPassword = "password"
	DefaultAdminEmail    = DefaultAdminName + "@example.invalid"

	ResticImageName                 = "restic:local"
	ResticCleanupLabel              = "quollix.cleanup=restic"
	ResticImageMaintenanceKeepLabel = "quollix.maintenance.keep=restic"

	OfficialDatabaseAppMaintainerAndAppName = u.OfficialMaintainer + "_" + u.OfficialDatabaseAppName
	OfficialDatabaseAppNetworkName          = OfficialDatabaseAppMaintainerAndAppName
	OfficialDatabaseAppContainerName        = OfficialDatabaseAppMaintainerAndAppName + "_" + u.OfficialDatabaseAppName

	BrandAppService       = u.BrandName
	BrandAppContainerName = u.OfficialMaintainer + "_" + u.OfficialBrandAppName + "_" + BrandAppService

	GrafanaAppName                         = "grafana"
	DefaultLogsRetentionDays               = 30
	DefaultLogsMaxDiskSpaceGB              = 5
	DefaultMetricsRetentionDays            = 30
	DefaultHostMetricsIntervalSeconds      = 300
	DefaultContainerMetricsIntervalSeconds = 300
	DefaultVolumeMetricsIntervalSeconds    = 3600
)

var ComposeEnvVars = struct {
	BaseDomain   string
	ClientId     string
	ClientSecret string

	IanaTimeZone               string
	LogsRetentionDays          string
	LogsMaxDiskSpaceGB         string
	MetricsRetentionDays       string
	HostMetricsIntervalSeconds string
}{
	BaseDomain:                 "BASE_DOMAIN",
	ClientId:                   "CLIENT_ID",
	ClientSecret:               "CLIENT_SECRET",
	IanaTimeZone:               "IANA_TIMEZONE",
	LogsRetentionDays:          "LOGS_RETENTION_DAYS",
	LogsMaxDiskSpaceGB:         "LOGS_MAX_DISK_SPACE_GB",
	MetricsRetentionDays:       "METRICS_RETENTION_DAYS",
	HostMetricsIntervalSeconds: "HOST_METRICS_INTERVAL_SECONDS",
}

type GlobalConfig struct {
	CollectMetricsAutomatically           bool
	PrintCommandOutput                    bool
	ShouldRunMaintenanceAgent             bool
	DeployOfficialDatabaseWithPortExposed bool
	RedirectHttpToHttps                   bool
	AppForwardedProto                     string
	DatabaseHostName                      string
	ExposeDevelopmentRoutes               bool
	CreateDatabaseSnapshotOnStartup       bool
	PruneDockerSystemDuringMaintenance    bool
	UseDevelopmentLogger                  bool
	UseLocalAppStoreClient                bool
	UseLocalTestingAuthorizedKey          bool
	UseStrictEmailClientStub              bool
	UseMockWildcardCertificateService     bool
	UseExternalDatabase                   bool
	AllowInsecureOidcProviderTls          bool
}
