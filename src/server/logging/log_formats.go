package logging

import "time"

var serviceLogFormats = map[string]string{
	"jitsi":           "nginx_container",
	"jicofo":          "jitsi_component",
	"jvb":             "jitsi_component",
	"victorialogs":    "victoria",
	"victoriametrics": "victoria",
	"wikijs":          "json",
	"wordpress":       "apache_container",
	"sampleapp":       "json",
}

var logFormats = map[string]logParser{
	"victoria":         jsonLogParser("ts", time.RFC3339Nano),
	"vector":           newLogParser(parseVectorLog, time.RFC3339Nano),
	"json":             jsonLogParser("timestamp", time.RFC3339Nano),
	"grist":            jsonLogParser("timestamp", time.DateTime),
	"rabbitmq":         jsonLogParser("time", dateTimeOffsetLayout),
	"vikunja":          jsonLogParser("time", time.RFC3339Nano),
	"grafana":          jsonLogParser("t", time.RFC3339Nano),
	"postgres":         parsePostgresRecord,
	"forgejo":          newLogParser(parseForgejoLog, slashTimestampLayout),
	"jitsi_component":  newLogParser(parseJitsiFields, time.DateTime),
	"prosody":          newLogParser(parseProsodyLog, time.DateTime),
	"nginx":            nginxRecordParser,
	"apache":           apacheRecordParser,
	"http_access":      accessRecordParser,
	"redis":            newLogParser(parseRedisLog, redisTimestampLayout),
	"mariadb":          newLogParser(parseMariadbLog, ""),
	"vaultwarden":      newLogParser(parseVaultwardenLog, time.DateTime),
	"wgeasy":           newLogParser(parseWgeasyLog, time.RFC3339Nano),
	"timestamp_level":  newLogParser(parseTimestampLevelLog, ""),
	"hedgedoc":         newLogParser(parseHedgedocLog, ""),
	"zulip":            newLogParser(parseTimestampLevelLog, time.DateTime),
	"adguard":          newLogParser(parseAdguardFields, slashTimestampLayout),
	"apache_container": parseApacheContainerLog,
	"nginx_container":  parseNginxContainerLog,
	"nextcloud":        parseNextcloudLog,
}
