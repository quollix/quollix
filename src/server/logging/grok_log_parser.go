package logging

import (
	"strings"

	"github.com/elastic/go-grok"
	"github.com/elastic/go-grok/patterns"
)

var (
	parseComponentMessageLog = grokParser(`\[%{COMPONENT:component}\] %{GREEDYDATA:message}`)
	parseApacheDiagnostic    = grokParser(`%{APACHECODE:error.code}: %{APACHEMODULE:component}: %{GREEDYDATA:message}`)
	parseLevelMessageLog     = grokParser(`%{KNOWNLEVEL:level}: %{GREEDYDATA:message}`)
	parseS6Log               = grokParser(`(?:s6-rc|/package/admin/s6-overlay/libexec/preinit): %{LEVEL:level}: %{GREEDYDATA:message}`)
	parseRabbitmqTextLog     = grokParser(`%{STAMP:timestamp}\s+\[%{LEVEL:level}\]\s+%{ERLANGPID:pid}\s+%{GREEDYDATA:message}`)
	parseRabbitClose         = grokParser(`closing AMQP connection %{ERLANGPID} \(%{IP:source.address}:%{NONNEGINT:source.port:int} -> %{IP:destination.address}:%{NONNEGINT:destination.port:int}, vhost: '%{DATA:rabbitmq.vhost}', user: '%{DATA:user.name}', duration: '%{DATA:connection.duration}'\):`)
	parseForgejoLog          = grokParser(`%{STAMP:timestamp}\s+%{NOTSPACE:source} \[%{LEVEL:level}\]\s+%{GREEDYDATA:message}`)
	parseJitsiLog            = grokParser(`%{JITSIAPP:app} %{STAMP:timestamp}\s+%{LEVEL:level}: \[%{NONNEGINT:thread_id:int}\] (?:\[%{BRACKET:context}\] )?%{NOTSPACE:source}: %{GREEDYDATA:message}`)
	parseProsodyLog          = grokParser(`%{STAMP:timestamp}\s+%{NOTSPACE:component}\s+%{LEVEL:level}\t%{GREEDYDATA:message}`)
	parseNginxLog            = grokParser(`%{STAMP:timestamp}\s+\[%{LEVEL:level}\] %{NONNEGINT:pid:int}#%{NONNEGINT:thread_id:int}: %{GREEDYDATA:message}`)
	parseRedisLog            = grokParser(`%{NONNEGINT:pid:int}:%{LEVEL:role}\s+%{REDISSTAMP:timestamp}\s+%{REDISLEVEL:level}\s+(?:<%{COMPONENT:component}> %{GREEDYDATA:message}|%{REDISBODY:message})`)
	parseMariadbLog          = grokParser(`%{STAMP:timestamp}\s+(?:%{NONNEGINT:thread_id:int} \[%{LEVEL:level}\] |\[%{LEVEL:level}\] \[%{COMPONENT:component}\]: )%{GREEDYDATA:message}`)
	parseVaultwardenLog      = grokParser(`\[%{BRACKETVALUE:timestamp}\]\[%{BRACKET:component}\]\[%{LEVEL:level}\] %{GREEDYDATA:message}`)
	parseWgeasyLog           = grokParser(`%{STAMP:timestamp}\s+%{WGCOMPONENT:component}\s+%{GREEDYDATA:message}`)
	parseTimestampLevelLog   = grokParser(`%{STAMP:timestamp}\s+%{KNOWNLEVEL:level}:?\s+%{GREEDYDATA:message}`)
	parseAdguardLog          = grokParser(`%{STAMP:timestamp}\s+\[%{LEVEL:level}\] %{GREEDYDATA:message}`)
	parseHTTPAccessFields    = grokParser(`%{HTTPD_COMBINEDLOG}|%{HTTPD_COMMONLOG}`)

	// HTTPD_ERRORLOG does not cover trace1 levels, space-padded days or empty messages.
	parseApacheLog = grokParser(
		`\[%{BRACKETVALUE:timestamp}\] \[%{APACHEMODULE:apache.error.module}:%{LEVEL:log.level}\] ` +
			`\[pid %{NONNEGINT:process.pid:int}(?::tid %{NONNEGINT:process.thread.id:int})?\]` +
			`(?: (?:\[client %{CLIENT:source.address}(?::%{NONNEGINT:source.port:int})?\](?: |$))?` +
			`(?:%{APACHECODE:error.code}: ?)?%{GREEDYDATA:message})?`)
)

func grokParser(pattern string) func(string) (map[string]any, bool) {
	parser, err := grok.NewWithPatterns(patterns.Httpd, map[string]string{
		// Preserve unusual request lines that the built-in pattern otherwise discards.
		"HTTPD_COMMONLOG": strings.ReplaceAll(patterns.Httpd["HTTPD_COMMONLOG"], "%{DATA}", "%{DATA:http.request.line}"),
		"STAMP":           `\d{4}[-/]\d{2}[-/]\d{2}(?: +|T)\S+`,
		"DECIMAL":         `[0-9]+(?:\.[0-9]+)?`,
		"LEVEL":           `[A-Za-z][A-Za-z0-9]*`,
		"ERLANGPID":       `<[0-9]+\.[0-9]+\.[0-9]+>`,
		"KNOWNLEVEL":      `(?i:trace|debug|info|warn|warning|error|fatal|panic|notice|log|http|verbose|silly)`,
		"BRACKETVALUE":    `[^\]]+`,
		"BRACKET":         `[^\]]*`,
		"COMPONENT":       `[^\]>]+`,
		"JITSIAPP":        `JVB|Jicofo`,
		"REDISSTAMP":      `\d{1,2} \w{3} \d{4} \S+`,
		"REDISLEVEL":      `[.*#-]`,
		"REDISBODY":       `(?:[^<](?s:.*))?`,
		"WGCOMPONENT":     `Database|WireGuard|CMD|Firewall`,
		"APACHEMODULE":    `[^:\]]+`,
		"CLIENT":          `[^:\] ]+`,
		"APACHECODE":      `AH[0-9]+`,
		"LSN":             `[0-9A-Fa-f]{1,8}/[0-9A-Fa-f]{1,8}`,
	})
	if err == nil {
		// Match the entire record so malformed suffixes cannot silently lose data.
		err = parser.Compile(`(?s)\A(?:`+pattern+`)\z`, true)
	}
	return func(line string) (map[string]any, bool) {
		if err != nil {
			return nil, false
		}
		fields, err := parser.ParseTypedString(line)
		return fields, err == nil && len(fields) > 0
	}
}

func parseHTTPAccessLog(line string) (map[string]any, bool) {
	fields, ok := parseHTTPAccessFields(line)
	if !ok {
		return nil, false
	}
	fields["message"] = "HTTP request"
	return fields, true
}
