package configs

var ConfigKeys = struct {
	BaseDomain                      string
	LogsRetentionDays               string
	LogsMaxDiskSpaceGB              string
	MetricsRetentionDays            string
	HostMetricsIntervalSeconds      string
	ContainerMetricsIntervalSeconds string
	VolumeMetricsIntervalSeconds    string

	OidcPrivateKey             string
	ExposeRealEmailInOidcToken string

	CertificatePemBundle string

	SmtpHost             string
	SmtpPort             string
	SmtpFromEmailAddress string
	SmtpUsername         string
	SmtpPassword         string
	SmtpEnabled          string

	BackupEnabled            string
	BackupServerHost         string
	BackupServerPort         string
	BackupServerUser         string
	BackupServerPassword     string
	BackupServerKnownHosts   string
	BackupEncryptionPassword string
	BackupFolderPath         string

	InvitationEmailTemplate string
	BroadcastEmailTemplate  string

	BackupRetentionPreUpdated string
	BackupRetentionDaily      string
	BackupRetentionWeekly     string
	BackupRetentionMonthly    string
	BackupRetentionYearly     string

	MaintenanceIanaTimeZone    string
	MaintenanceWindowStartHour string
	MaintenanceNextAtUtc       string

	ApplicationVersion   string
	SystemConfigVersion  string
	ResticDockerfileHash string

	// acme library derives the public key from the private key, so we only need to store the private key
	AcmeAccountPrivateKey         string
	AcmeAccountRegistered         string
	AlertCPUUsageEnabled          string
	AlertMemoryUsageEnabled       string
	AlertDiskUsageEnabled         string
	AlertCertificateExpiryEnabled string
	AlertQuollixErrorsEnabled     string
	AlertHostUsageTriggerPercent  string
	AlertHostUsageRecoveryPercent string
	AlertHostUsagePendingMinutes  string
	AlertCPUUsageState            string
	AlertCPUUsagePendingSince     string
	AlertMemoryUsageState         string
	AlertMemoryUsagePendingSince  string
	AlertDiskUsageState           string
	AlertDiskUsagePendingSince    string
	AlertCertificateExpiryState   string
	AlertQuollixErrorsState       string
	LatestBlogArticleDate         string
	LastBlogFetchAttempt          string
}{
	BaseDomain:                      "base_domain",
	LogsRetentionDays:               "logs_retention_days",
	LogsMaxDiskSpaceGB:              "logs_max_disk_space_gb",
	MetricsRetentionDays:            "metrics_retention_days",
	HostMetricsIntervalSeconds:      "host_metrics_interval_seconds",
	ContainerMetricsIntervalSeconds: "container_metrics_interval_seconds",
	VolumeMetricsIntervalSeconds:    "volume_metrics_interval_seconds",

	OidcPrivateKey:             "oidc_private_key",
	ExposeRealEmailInOidcToken: "expose_real_email_in_oidc_token",

	CertificatePemBundle: "certificate_pem_bundle",

	SmtpHost:             "smtp_host",
	SmtpPort:             "smtp_port",
	SmtpFromEmailAddress: "smtp_from_email_address",
	SmtpUsername:         "smtp_username",
	SmtpPassword:         "smtp_password",
	SmtpEnabled:          "smtp_enabled",

	BackupEnabled:            "backup_enabled",
	BackupServerHost:         "backup_server_host",
	BackupServerPort:         "backup_server_port",
	BackupServerUser:         "backup_server_user",
	BackupServerPassword:     "backup_server_password",
	BackupServerKnownHosts:   "backup_server_known_hosts",
	BackupEncryptionPassword: "backup_restic_password",
	BackupFolderPath:         "backup_folder_path",

	InvitationEmailTemplate: "email_template_invitation",
	BroadcastEmailTemplate:  "email_template_broadcast",

	BackupRetentionPreUpdated: "backup_retention_pre_updated",
	BackupRetentionDaily:      "backup_retention_daily",
	BackupRetentionWeekly:     "backup_retention_weekly",
	BackupRetentionMonthly:    "backup_retention_monthly",
	BackupRetentionYearly:     "backup_retention_yearly",

	MaintenanceIanaTimeZone:    "maintenance_iana_time_zone",
	MaintenanceWindowStartHour: "maintenance_window_start_hour",
	MaintenanceNextAtUtc:       "maintenance_next_at_utc",

	ApplicationVersion:   "application_version",
	SystemConfigVersion:  "system_config_version",
	ResticDockerfileHash: "restic_dockerfile_sha256",

	AcmeAccountPrivateKey:         "acme_account_private_key",
	AcmeAccountRegistered:         "acme_account_registered",
	AlertCPUUsageEnabled:          "alert_cpu_usage_enabled",
	AlertMemoryUsageEnabled:       "alert_memory_usage_enabled",
	AlertDiskUsageEnabled:         "alert_disk_usage_enabled",
	AlertCertificateExpiryEnabled: "alert_certificate_expiry_enabled",
	AlertQuollixErrorsEnabled:     "alert_quollix_errors_enabled",
	AlertHostUsageTriggerPercent:  "alert_host_usage_trigger_percent",
	AlertHostUsageRecoveryPercent: "alert_host_usage_recovery_percent",
	AlertHostUsagePendingMinutes:  "alert_host_usage_pending_minutes",
	AlertCPUUsageState:            "alert_cpu_usage_state",
	AlertCPUUsagePendingSince:     "alert_cpu_usage_pending_since",
	AlertMemoryUsageState:         "alert_memory_usage_state",
	AlertMemoryUsagePendingSince:  "alert_memory_usage_pending_since",
	AlertDiskUsageState:           "alert_disk_usage_state",
	AlertDiskUsagePendingSince:    "alert_disk_usage_pending_since",
	AlertCertificateExpiryState:   "alert_certificate_expiry_state",
	AlertQuollixErrorsState:       "alert_quollix_errors_state",
	LatestBlogArticleDate:         "blog_latest_article_date",
	LastBlogFetchAttempt:          "blog_last_fetch_attempt",
}
