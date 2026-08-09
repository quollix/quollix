package apps_basic

import (
	"server/tools"

	u "github.com/quollix/common/utils"
)

type legacySecretKey struct {
	maintainer string
	appName    string
	secretName string
}

// Deprecated: compatibility seeds for app definitions that used static credentials before SECRET_* metadata existed.
// Remove after supported old installs have a real credential rotation migration.
var legacyStaticSecretValues = map[legacySecretKey]string{
	{u.OfficialMaintainer, "forgejo", "SECRET_POSTGRES_PASSWORD"}:                "forgejo",
	{u.OfficialMaintainer, "hedgedoc", "SECRET_POSTGRES_PASSWORD"}:               "password",
	{u.OfficialMaintainer, "jitsi", "SECRET_JICOFO_AUTH_PASSWORD"}:               "password",
	{u.OfficialMaintainer, "jitsi", "SECRET_JVB_AUTH_PASSWORD"}:                  "password",
	{u.OfficialMaintainer, "jitsi", "SECRET_JICOFO_COMPONENT_SECRET"}:            "password",
	{u.OfficialMaintainer, "nextcloud", "SECRET_MARIADB_PASSWORD"}:               "nextcloud",
	{u.OfficialMaintainer, "nextcloud", "SECRET_MARIADB_ROOT_PASSWORD"}:          "nextcloud",
	{u.OfficialMaintainer, "vaultwarden", "SECRET_ADMIN_TOKEN"}:                  "quollix",
	{u.OfficialMaintainer, "wikijs", "SECRET_POSTGRES_PASSWORD"}:                 "password",
	{u.OfficialMaintainer, "wordpress", "SECRET_MARIADB_PASSWORD"}:               "wordpress",
	{u.OfficialMaintainer, "wordpress", "SECRET_MARIADB_ROOT_PASSWORD"}:          "wordpress",
	{u.OfficialMaintainer, "zulip", "SECRET_MEMCACHED_PASSWORD"}:                 "password",
	{u.OfficialMaintainer, "zulip", "SECRET_POSTGRES_PASSWORD"}:                  "password",
	{u.OfficialMaintainer, "zulip", "SECRET_RABBITMQ_PASSWORD"}:                  "password",
	{u.OfficialMaintainer, "zulip", "SECRET_REDIS_PASSWORD"}:                     "password",
	{tools.SampleMaintainer, tools.SampleApp, "SECRET_SAMPLE_MIGRATED_PASSWORD"}: "password",
}

func legacySecretValue(existingApp *RepoApp, secretName string) (string, bool) {
	if existingApp == nil {
		return "", false
	}

	key := legacySecretKey{
		maintainer: existingApp.Maintainer,
		appName:    existingApp.AppName,
		secretName: secretName,
	}
	if value, exists := legacyStaticSecretValues[key]; exists {
		return value, true
	}
	return "", false
}
