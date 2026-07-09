package configs

var ConfigKeys = struct {
	BaseDomain string

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

	InvitationEmailTemplate string

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
	AcmeAccountPrivateKey string
	AcmeAccountRegistered string
}{
	BaseDomain: "base_domain",

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

	InvitationEmailTemplate: "email_template_invitation",

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

	AcmeAccountPrivateKey: "acme_account_private_key",
	AcmeAccountRegistered: "acme_account_registered",
}
