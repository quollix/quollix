package logging

import "time"

var serviceLogFormats = map[string]string{
	"jicofo":          "jitsi_component",
	"jvb":             "jitsi_component",
	"victorialogs":    "victoria",
	"victoriametrics": "victoria",
	"wikijs":          "json",
	"wordpress":       "apache_container",
	"sampleapp":       "json",
}

/* Parser spec:
Parsers recognize recurring log formats, not individual message wording. A service may combine formats when each covers a family of log entries.
Leave isolated banners, setup messages and traceback fragments unparsed instead of adding special cases.
Match complete lines. Do not reconstruct multiline events here.
Tests should use one representative complete line per format, with additional cases only for distinct behavior such as timestamp variants.
*/

var logFormats = map[string]logParser{
	"jitsi":            parseJitsiContainerLog,
	"victoria":         jsonLogParser("ts", time.RFC3339Nano),
	"vector":           newLogParser(parseVectorLog, time.RFC3339Nano),
	"json":             jsonLogParser("timestamp", time.RFC3339Nano),
	"quollix":          jsonLogParser("time", time.RFC3339Nano),
	"grist":            jsonLogParser("timestamp", time.DateTime),
	"rabbitmq":         parseRabbitmqLog,
	"vikunja":          parseVikunjaLog,
	"grafana":          jsonLogParser("t", time.RFC3339Nano),
	"postgres":         parsePostgresRecord,
	"forgejo":          newLogParser(parseForgejoLog, slashTimestampLayout),
	"jitsi_component":  newLogParser(parseJitsiFields, time.DateTime),
	"prosody":          newLogParser(parseProsodyFields, time.DateTime),
	"nginx":            nginxRecordParser,
	"apache":           apacheRecordParser,
	"http_access":      accessRecordParser,
	"redis":            newLogParser(parseRedisLog, redisTimestampLayout),
	"mariadb":          newLogParser(parseMariadbLog, ""),
	"vaultwarden":      newLogParser(parseVaultwardenLog, time.DateTime),
	"wgeasy":           newLogParser(parseWgeasyLog, time.RFC3339Nano),
	"timestamp_level":  newLogParser(parseTimestampLevelLog, ""),
	"hedgedoc":         newLogParser(parseHedgedocLog, ""),
	"zulip":            newLogParser(parseZulipFields, time.DateTime),
	"adguard":          newLogParser(parseAdguardFields, slashTimestampLayout),
	"apache_container": parseApacheContainerLog,
	"nginx_container":  parseNginxContainerLog,
	"nextcloud":        parseNextcloudLog,
}
