package logging

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

const postgresCheckpointMessage = "checkpoint complete: wrote 123 buffers (4.5%); 1 WAL file(s) added, 2 removed, 3 recycled; write=1.234 s, sync=0.056 s, total=1.300 s; sync files=7, longest=0.020 s, average=0.008 s; distance=8192 kB, estimate=16384 kB; lsn=1/ABCDEF01, redo lsn=1/ABCDEF00"

type expectedLogLine struct {
	Timestamp  *time.Time
	Level      string
	Message    string
	Attributes map[string]any
}

const nextcloudApplicationLog = `[Fri Oct 09 09:23:27.104962 2026] [php:notice] [pid 67:tid 67] [client 172.28.0.2:38156] [nextcloud][no app in context][2] {"reqId":"dVPO4QpQHGJpUfzPD3JD","level":2,"time":"2026-10-09T08:23:27+00:00","remoteAddr":"172.28.0.2","user":"--","app":"no app in context","method":"POST","url":"/login","scriptName":"/index.php","message":"Login failed: asd (Remote IP: 172.28.0.2)","userAgent":"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/155.0.0.0 Safari/537.36","version":"35.0.1.1","data":[]}, referer: https://nextcloud.community.quollix.org/login`

func TestInterpret_ExtractsFormatData(t *testing.T) {
	for _, tc := range []struct {
		line     string
		service  string
		expected *expectedLogLine
	}{
		{`[xorm] [info]  2026/10/17 01:27:32.672307 PING DATABASE pgx`, "vikunja", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-17T01:27:32.672307Z"),
			Level:     "info", Message: "PING DATABASE pgx",
			Attributes: map[string]any{"component": "xorm"},
		}},
		{`time=2026-10-16T06:54:56.971Z level=ERROR msg="Could not connect to db: dial tcp 172.29.0.3:5432: connect: connection refused"`, "vikunja", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-16T06:54:56.971Z"),
			Level:     "error", Message: "Could not connect to db: dial tcp 172.29.0.3:5432: connect: connection refused",
		}},
		{`initdb: warning: enabling "trust" authentication for local connections`, "postgres", &expectedLogLine{
			Level: "warn", Message: `enabling "trust" authentication for local connections`,
			Attributes: map[string]any{"component": "initdb"},
		}},
		{`{"go.version":"go1.26.8","instance.id":"01a16ef5-b934-7af7-bf81-72952e2db4ae","http.response.duration":"2.115026156s","http.response.status":200,"time":"2026-10-24T16:06:06.169569872Z","level":"info","msg":"response completed","version":"3.1.2","vars.name":"library/nginx"}`, "dockerregistry", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-24T16:06:06.169569872Z"),
			Level:     "info", Message: "response completed",
			Attributes: map[string]any{
				"event.duration": int64(2115026156), "http.response.status": float64(200), "vars.name": "library/nginx",
				"go.version": "go1.26.8", "instance.id": "01a16ef5-b934-7af7-bf81-72952e2db4ae", "version": "3.1.2",
			},
		}},
		{`{"go.version":"go1.26.8","instance.id":"01a16ef5-b934-7af7-bf81-72952e2db4ae","level":"info","msg":"listening on [::]:5000","time":"2026-10-24T17:29:20.285325361Z","version":"3.1.2"}`, "dockerregistry", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-24T17:29:20.285325361Z"),
			Level:     "info", Message: "listening on [::]:5000",
			Attributes: map[string]any{
				"go.version": "go1.26.8", "instance.id": "01a16ef5-b934-7af7-bf81-72952e2db4ae", "version": "3.1.2",
			},
		}},
		{`192.168.16.2 - - [24/Oct/2026:16:06:04 +0000] "GET /v2/library/nginx/blobs/sha256:f2be67225471500c788ce88c597eb5ea8336e912ad1ca4815650e0281afe2034?ns=docker.io HTTP/1.1" 200 29843579 "" "docker/29.8.0 go/go1.26.8 git-commit/3ce5872 kernel/6.17.0-1032-oem os/linux arch/amd64 containerd-client/2.3.4+unknown storage-driver/overlayfs UpstreamClient(Docker-Client/29.8.0 \\(linux\\))"`, "dockerregistry", &expectedLogLine{
			Level: "unknown", Message: "HTTP request", Timestamp: testTimestamp(t, "2026-10-24T16:06:04Z"),
			Attributes: map[string]any{
				"source.address": "192.168.16.2", "http.request.method": "GET",
				"url.original": "/v2/library/nginx/blobs/sha256:f2be67225471500c788ce88c597eb5ea8336e912ad1ca4815650e0281afe2034?ns=docker.io",
				"http.version": "1.1", "http.response.status_code": 200, "http.response.body.size": 29843579,
				"user_agent.original": "docker/29.8.0 go/go1.26.8 git-commit/3ce5872 kernel/6.17.0-1032-oem os/linux arch/amd64 containerd-client/2.3.4+unknown storage-driver/overlayfs UpstreamClient(Docker-Client/29.8.0 \\\\(linux\\\\))",
			},
		}},
		{`[register-setup] JVB user registration output: Error: User exists`, "prosody", &expectedLogLine{
			Level: "unknown", Message: "JVB user registration output: Error: User exists",
			Attributes: map[string]any{"component": "register-setup"},
		}},
		{`AH00558: apache2: Could not reliably determine the server's fully qualified domain name, using 172.28.0.5. Set the 'ServerName' directive globally to suppress this message`, "nextcloud", &expectedLogLine{
			Level: "unknown", Message: "Could not reliably determine the server's fully qualified domain name, using 172.28.0.5. Set the 'ServerName' directive globally to suppress this message",
			Attributes: map[string]any{"error.code": "AH00558", "component": "apache2"},
		}},
		{`Warning: /var/www/html/config/s3.config.php differs from the latest version of this image at /usr/src/nextcloud/config/s3.config.php`, "nextcloud", &expectedLogLine{
			Level: "warn", Message: "/var/www/html/config/s3.config.php differs from the latest version of this image at /usr/src/nextcloud/config/s3.config.php",
		}},
		{`Notice: /Stage[main]/Zulip::Supervisor/Service[supervisor]: Triggered 'refresh' from 2 events`, "zulip", &expectedLogLine{
			Level: "info", Message: "/Stage[main]/Zulip::Supervisor/Service[supervisor]: Triggered 'refresh' from 2 events",
		}},
		{`/package/admin/s6-overlay/libexec/preinit: info: /run belongs to uid 0 instead of 1000 - fixing it`, "jitsi", &expectedLogLine{
			Level: "info", Message: "/run belongs to uid 0 instead of 1000 - fixing it",
		}},
		{`s6-rc: warning: service nginx failed to start`, "jitsi", &expectedLogLine{
			Level: "warn", Message: "service nginx failed to start",
		}},
		{`2026-10-17  1:27:19 0 [Note] mariadbd: ready for connections.`, "mariadb", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-17T01:27:19Z"),
			Level:     "info", Message: "mariadbd: ready for connections.",
			Attributes: map[string]any{"thread_id": 0},
		}},
		{`2026-10-17 01:28:04.272591+00:00 [info] <0.1164.0> connection 172.31.0.7:50896 -> 172.31.0.5:5672: user 'zulip' authenticated and granted access to vhost '/'`, "rabbitmq", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-17T01:28:04.272591Z"),
			Level:     "info", Message: "connection 172.31.0.7:50896 -> 172.31.0.5:5672: user 'zulip' authenticated and granted access to vhost '/'",
			Attributes: map[string]any{"pid": "<0.1164.0>"},
		}},
		{`2026-10-21 01:37:58.961821+00:00 [warning] <0.1046.0> closing AMQP connection <0.1046.0> (172.31.0.7:47964 -> 172.31.0.4:5672, vhost: '/', user: 'zulip', duration: '1D, 0H, 10M, 25s'):`, "rabbitmq", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-21T01:37:58.961821Z"),
			Level:     "warn", Message: "closing AMQP connection <0.1046.0> (172.31.0.7:47964 -> 172.31.0.4:5672, vhost: '/', user: 'zulip', duration: '1D, 0H, 10M, 25s'):",
			Attributes: map[string]any{
				"pid": "<0.1046.0>", "source.address": "172.31.0.7", "source.port": 47964,
				"destination.address": "172.31.0.4", "destination.port": 5672,
				"rabbitmq.vhost": "/", "user.name": "zulip",
				"event.duration": int64((24*time.Hour + 10*time.Minute + 25*time.Second).Nanoseconds()),
			},
		}},
		{`{"time":"2026-10-17T01:27:59.965149849Z","level":"INFO","source":{"function":"server/tools.(*DockerServiceImpl).StartAppContainer","file":"server/tools/docker_service.go","line":131},"msg":"Starting app","app":"vaultwarden","compose_file_path":"/tmp/app-compose-1989703575/docker-compose.yml","maintainer":"quollix"}`, "quollix", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-17T01:27:59.965149849Z"), Level: "info", Message: "Starting app",
			Attributes: map[string]any{
				"source": map[string]any{
					"function": "server/tools.(*DockerServiceImpl).StartAppContainer",
					"file":     "server/tools/docker_service.go",
					"line":     float64(131),
				},
				"app":               "vaultwarden",
				"compose_file_path": "/tmp/app-compose-1989703575/docker-compose.yml",
				"maintainer":        "quollix",
			},
		}},
		{`{"timestamp":"2026-10-12T17:36:20.580492Z","level":"INFO","message":"Log level is enabled.","level":"\"debug\"","target":"vector::app"}`, "vector", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-12T17:36:20.580492Z"), Level: "info", Message: "Log level is enabled.",
			Attributes: map[string]any{"target": "vector::app"},
		}},
		{`[nextcloud][updater][1] {"reqId":"oLJjgMVkn1KqIs4t0E9P","level":1,"time":"2026-10-17T14:45:13+00:00","remoteAddr":"","user":"--","app":"updater","method":"","url":"--","scriptName":"/var/www/html/occ","message":"\\OC\\Updater::appUpgrade: Updated <weather_status> to 2.0.0","userAgent":"--","version":"35.0.1.1","occ_command":["/var/www/html/occ","upgrade"],"data":{"app":"updater"}}`, "nextcloud", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-17T14:45:13Z"),
			Level:     "info",
			Message:   `\OC\Updater::appUpgrade: Updated <weather_status> to 2.0.0`,
			Attributes: map[string]any{
				"reqId": "oLJjgMVkn1KqIs4t0E9P", "remoteAddr": "", "user": "--",
				"app": "updater", "method": "", "url": "--", "scriptName": "/var/www/html/occ",
				"userAgent": "--", "version": "35.0.1.1",
				"occ_command": []any{"/var/www/html/occ", "upgrade"}, "data": map[string]any{"app": "updater"},
			},
		}},
		{nextcloudApplicationLog, "nextcloud", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-09T08:23:27Z"),
			Level:     "warn",
			Message:   "Login failed: asd",
			Attributes: map[string]any{
				"apache.error.module": "php", "process.pid": 67, "process.thread.id": 67,
				"source.address": "172.28.0.2", "source.port": 38156,
				"reqId": "dVPO4QpQHGJpUfzPD3JD", "remoteAddr": "172.28.0.2",
				"user": "--", "app": "no app in context", "method": "POST", "url": "/login",
				"scriptName": "/index.php", "version": "35.0.1.1", "data": []any{},
				"userAgent":             "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/155.0.0.0 Safari/537.36",
				"http.request.referrer": "https://nextcloud.community.quollix.org/login",
			},
		}},
		{`{"msg":"","n":1}`, "vikunja", &expectedLogLine{
			Message:    defaultStructuredLogMessage,
			Level:      "unknown",
			Attributes: map[string]any{"n": float64(1)},
		}},
		{`{"msg":"hello","n":1}`, "vikunja", &expectedLogLine{
			Level:   "unknown",
			Message: "hello",
			Attributes: map[string]any{
				"n": float64(1),
			},
		}},
		{`{"label":"MASTER","level":"info","message":"= Wiki.js 2.5.316 =====================","timestamp":"2026-10-03T09:29:08.589Z"}`, "wikijs", &expectedLogLine{
			Level:     "info",
			Message:   "= Wiki.js 2.5.316 =====================",
			Timestamp: testTimestamp(t, "2026-10-03T09:29:08.589000Z"),
			Attributes: map[string]any{
				"label": "MASTER",
			},
		}},
		{`{"level":"info","message":"RestartShell: child ready","timestamp":"2026-10-03 10:29:55.285"}`, "grist", &expectedLogLine{
			Level:     "info",
			Message:   "RestartShell: child ready",
			Timestamp: testTimestamp(t, "2026-10-03T10:29:55.285000Z"),
		}},
		{`{"error_logger":{"tag":"info_report","type":"std_info","report_cb":"fun application_controller:format_log/1"},"line":2125,"pid":"<0.45.0>","time":"2026-10-03 10:30:14.695462+01:00","file":"application_controller.erl","gl":"<0.69.0>","domain":"otp","msg":"Application syslog exited with reason: stopped","level":"notice","mfa":["application_controller","info_exited",3],"report_cb":"fun application_controller:format_log/2"}`, "rabbitmq", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T09:30:14.695462Z"),
			Message:   "Application syslog exited with reason: stopped",
			Level:     "info",
			Attributes: map[string]any{
				"error_logger": map[string]any{
					"tag":       "info_report",
					"type":      "std_info",
					"report_cb": "fun application_controller:format_log/1",
				},
				"line":      float64(2125),
				"pid":       "<0.45.0>",
				"file":      "application_controller.erl",
				"gl":        "<0.69.0>",
				"domain":    "otp",
				"mfa":       []any{"application_controller", "info_exited", float64(3)},
				"report_cb": "fun application_controller:format_log/2",
			},
		}},
		{`{"duration":"1.6096ms","handler":"/api/datasources/uid/:uid/health","level":"info","logger":"context","method":"GET","msg":"Request Completed","orgId":1,"path":"/api/datasources/uid/cg0lndiimru9sa/health","referer":"https://grafana.test.quollix.org/connections/datasources/edit/cg0lndiimru9sa","remote_addr":"172.21.0.2","size":159,"status":400,"status_source":"server","t":"2026-10-08T11:42:24.252037652+01:00","time_ms":1,"uname":"admin","userId":1}`, "grafana", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-08T10:42:24.252037652Z"),
			Level:     "info",
			Message:   "Request Completed",
			Attributes: map[string]any{
				"duration": "1.6096ms", "handler": "/api/datasources/uid/:uid/health", "logger": "context", "method": "GET",
				"orgId": float64(1), "path": "/api/datasources/uid/cg0lndiimru9sa/health",
				"referer": "https://grafana.test.quollix.org/connections/datasources/edit/cg0lndiimru9sa", "remote_addr": "172.21.0.2",
				"size": float64(159), "status": float64(400), "status_source": "server", "time_ms": float64(1), "uname": "admin", "userId": float64(1),
			},
		}},
		{`2026-10-03 10:29:44.176 BST [1] LOG:  database system is ready to accept connections`, "postgres", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T09:29:44.176Z"),
			Level:     "info",
			Message:   "database system is ready to accept connections",
			Attributes: map[string]any{
				"pid": 1,
			},
		}},
		{`2026/10/03 10:29:45 ...s/storage/storage.go:157:initAttachments() [I] Initialising Attachment storage with type: local`, "forgejo", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T10:29:45Z"),
			Level:     "info",
			Message:   "Initialising Attachment storage with type: local",
			Attributes: map[string]any{
				"source": "...s/storage/storage.go:157:initAttachments()",
			},
		}},
		{`Jicofo 2026-10-03 10:31:09.499 INFO: [1] Main.main#49: Starting Jicofo.`, "jicofo", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T10:31:09.499Z"),
			Level:     "info",
			Message:   "Starting Jicofo.",
			Attributes: map[string]any{
				"app":       "Jicofo",
				"thread_id": 1,
				"source":    "Main.main#49",
			},
		}},
		{`JVB 2026-10-03 10:31:14.002 WARNING: [35] [hostname=quollix_jitsi_prosody id=shard0] MucClient.setPresenceExtensions#519: Cannot set presence extension: not connected.`, "jvb", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T10:31:14.002Z"),
			Level:     "warn",
			Message:   "Cannot set presence extension: not connected.",
			Attributes: map[string]any{
				"app":       "JVB",
				"thread_id": 35,
				"context":   map[string]any{"hostname": "quollix_jitsi_prosody", "id": "shard0"},
				"source":    "MucClient.setPresenceExtensions#519",
			},
		}},
		{"2026-10-03 10:31:09 polls.meet.jitsi:tls              info\tCertificates loaded", "prosody", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T10:31:09Z"),
			Level:     "info",
			Message:   "Certificates loaded",
			Attributes: map[string]any{
				"component": "polls.meet.jitsi:tls",
			},
		}},
		{`2026/10/03 10:31:09 [warn] 115#115: could not build optimal variables_hash`, "jitsi", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T10:31:09Z"),
			Level:     "warn",
			Message:   "could not build optimal variables_hash",
			Attributes: map[string]any{
				"pid":       115,
				"thread_id": 115,
			},
		}},
		{`[Sat Oct 03 10:30:41.019843 2026] [mpm_prefork:notice] [pid 1:tid 1] AH00163: Apache configured -- resuming normal operations`, "nextcloud", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T10:30:41.019843Z"),
			Level:     "info",
			Message:   "Apache configured -- resuming normal operations",
			Attributes: map[string]any{
				"apache.error.module": "mpm_prefork",
				"process.pid":         1,
				"process.thread.id":   1,
				"error.code":          "AH00163",
			},
		}},
		{`1:C 03 Oct 2026 10:30:40.487 * Configuration loaded`, "redis", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T10:30:40.487Z"),
			Level:     "info",
			Message:   "Configuration loaded",
			Attributes: map[string]any{
				"pid":  1,
				"role": "C",
			},
		}},
		{`1:M 03 Oct 2026 10:30:40.492 * <bf> RedisBloom version 8.6.6`, "redis", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T10:30:40.492Z"),
			Level:     "info",
			Message:   "RedisBloom version 8.6.6",
			Attributes: map[string]any{
				"pid":       1,
				"role":      "M",
				"component": "bf",
			},
		}},
		{`2026-10-03 10:30:41 0 [Note] Starting MariaDB`, "mariadb", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T10:30:41Z"),
			Level:     "info",
			Message:   "Starting MariaDB",
			Attributes: map[string]any{
				"thread_id": 0,
			},
		}},
		{`2026-10-03 10:30:40+01:00 [Note] [Entrypoint]: Entrypoint script for MariaDB Server started.`, "mariadb", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T09:30:40Z"),
			Level:     "info",
			Message:   "Entrypoint script for MariaDB Server started.",
			Attributes: map[string]any{
				"component": "Entrypoint",
			},
		}},
		{`[2026-10-03 10:31:17.421][start][INFO] Rocket has launched from http://0.0.0.0:80`, "vaultwarden", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T10:31:17.421Z"),
			Level:     "info",
			Message:   "Rocket has launched from http://0.0.0.0:80",
			Attributes: map[string]any{
				"component": "start",
			},
		}},
		{`2026-10-03T09:30:20.167Z WireGuard Starting WireGuard...`, "wgeasy", &expectedLogLine{
			Level:     "unknown",
			Timestamp: testTimestamp(t, "2026-10-03T09:30:20.167Z"),
			Message:   "Starting WireGuard...",
			Attributes: map[string]any{
				"component": "WireGuard",
			},
		}},
		{`2026-10-03T09:30:20.538Z CMD $ wg syncconf wg0 <(wg-quick strip wg0)`, "wgeasy", &expectedLogLine{
			Level:     "unknown",
			Timestamp: testTimestamp(t, "2026-10-03T09:30:20.538Z"),
			Message:   "$ wg syncconf wg0 <(wg-quick strip wg0)",
			Attributes: map[string]any{
				"component": "CMD",
			},
		}},
		{`2025-01-23 01:23:23,123 INFO something happened`, "zulip", &expectedLogLine{
			Timestamp: testTimestamp(t, "2025-01-23T01:23:23.123Z"),
			Level:     "info",
			Message:   "something happened",
		}},
		{"2026-10-03T09:30:55.672Z info: \tAll migrations performed successfully", "hedgedoc", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T09:30:55.672Z"),
			Level:     "info",
			Message:   "All migrations performed successfully",
		}},
		{"2026-10-03T09:31:04.812Z info: \t" + `127.0.0.1 - - [03/Oct/2026:09:31:04 +0000] "GET /_health HTTP/1.1" 200 14 "-" "hedgedoc-container-healthcheck/1.2"`, "hedgedoc", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T09:31:04.812Z"),
			Level:     "info",
			Message:   "HTTP request",
			Attributes: map[string]any{
				"source.address":            "127.0.0.1",
				"http.request.method":       "GET",
				"url.original":              "/_health",
				"http.version":              "1.1",
				"http.response.status_code": 200,
				"http.response.body.size":   14,
				"user_agent.original":       "hedgedoc-container-healthcheck/1.2",
			},
		}},
		{`2026/10/03 10:29:27.226036 [info] starting adguard home version="AdGuard Home, version v0.107.79"`, "adguard", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T10:29:27.226036Z"),
			Level:     "info",
			Message:   "starting adguard home",
			Attributes: map[string]any{
				"context": map[string]any{"version": "AdGuard Home, version v0.107.79"},
			},
		}},
		{`2026/10/03 10:29:27.272821 [info] dnsproxy: upstream mode is set mode=load_balance note="uses two upstreams"`, "adguard", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T10:29:27.272821Z"),
			Level:     "info",
			Message:   "upstream mode is set",
			Attributes: map[string]any{
				"component": "dnsproxy",
				"context":   map[string]any{"mode": "load_balance", "note": "uses two upstreams"},
			},
		}},

		{`127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326`, "nextcloud", &expectedLogLine{
			Level:     "unknown",
			Message:   "HTTP request",
			Timestamp: testTimestamp(t, "2000-10-10T20:55:36Z"),
			Attributes: map[string]any{
				"source.address":            "127.0.0.1",
				"http.request.method":       "GET",
				"url.original":              "/apache_pb.gif",
				"http.version":              "1.0",
				"http.response.status_code": 200,
				"http.response.body.size":   2326,
			},
		}},
		{`172.30.0.2 - - [03/Oct/2026:13:09:31 +0100] "GET / HTTP/1.1" 200 12043 "https://quollix.community.quollix.org/" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/154.0.0.0 Safari/537.36"`, "wordpress", &expectedLogLine{
			Level:     "unknown",
			Message:   "HTTP request",
			Timestamp: testTimestamp(t, "2026-10-03T12:09:31Z"),
			Attributes: map[string]any{
				"source.address":            "172.30.0.2",
				"http.request.method":       "GET",
				"url.original":              "/",
				"http.version":              "1.1",
				"http.response.status_code": 200,
				"http.response.body.size":   12043,
				"http.request.referrer":     "https://quollix.community.quollix.org/",
				"user_agent.original":       "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/154.0.0.0 Safari/537.36",
			},
		}},
		{`[Sat Oct  3 10:30:41 2026] [core:trace1] [pid 12] cache: entry missing`, "apache", &expectedLogLine{
			Timestamp:  testTimestamp(t, "2026-10-03T10:30:41Z"),
			Level:      "trace",
			Message:    "cache: entry missing",
			Attributes: map[string]any{"apache.error.module": "core", "process.pid": 12},
		}},
		{`[Sat Oct 03 10:30:41.123 2026] [core:info] [pid 12] AH00163:`, "apache", &expectedLogLine{
			Message:    defaultStructuredLogMessage,
			Timestamp:  testTimestamp(t, "2026-10-03T10:30:41.123Z"),
			Level:      "info",
			Attributes: map[string]any{"apache.error.module": "core", "process.pid": 12, "error.code": "AH00163"},
		}},
		{`[Sat Oct 03 10:30:41 2026] [core:info] [pid 12]`, "apache", &expectedLogLine{
			Message:    defaultStructuredLogMessage,
			Timestamp:  testTimestamp(t, "2026-10-03T10:30:41Z"),
			Level:      "info",
			Attributes: map[string]any{"apache.error.module": "core", "process.pid": 12},
		}},
		{`::1 ident alice [03/Oct/2026:10:30:41 +0100] "GET / HTTP/1.1" 200 0 "" "-"`, "http_access", &expectedLogLine{
			Level:     "unknown",
			Message:   "HTTP request",
			Timestamp: testTimestamp(t, "2026-10-03T09:30:41Z"),
			Attributes: map[string]any{
				"source.address": "::1", "apache.access.user.identity": "ident", "user.name": "alice",
				"http.request.method": "GET", "url.original": "/", "http.version": "1.1", "http.response.status_code": 200, "http.response.body.size": 0,
			},
		}},
		{`::1 ident alice [03/Oct/2026:10:30:41 +0100] "GET / HTTP/1.1" 200 0`, "http_access", &expectedLogLine{
			Level:     "unknown",
			Message:   "HTTP request",
			Timestamp: testTimestamp(t, "2026-10-03T09:30:41Z"),
			Attributes: map[string]any{
				"source.address": "::1", "apache.access.user.identity": "ident", "user.name": "alice",
				"http.request.method": "GET", "url.original": "/", "http.version": "1.1", "http.response.status_code": 200, "http.response.body.size": 0,
			},
		}},
		{`127.0.0.1 - - [03/Oct/2026:10:30:41 +0000] "GET / HTTP/1.1" 304 - "-" "client \"quoted\""`, "http_access", &expectedLogLine{
			Level:     "unknown",
			Message:   "HTTP request",
			Timestamp: testTimestamp(t, "2026-10-03T10:30:41Z"),
			Attributes: map[string]any{
				"source.address": "127.0.0.1", "http.request.method": "GET", "url.original": "/", "http.version": "1.1",
				"http.response.status_code": 304, "user_agent.original": `client \"quoted\"`,
			},
		}},
		{`127.0.0.1 - - [03/Oct/2026:10:30:41 +0000] "GET / HTTP/1.1" 304 -`, "http_access", &expectedLogLine{
			Level:     "unknown",
			Message:   "HTTP request",
			Timestamp: testTimestamp(t, "2026-10-03T10:30:41Z"),
			Attributes: map[string]any{
				"source.address": "127.0.0.1", "http.request.method": "GET", "url.original": "/", "http.version": "1.1",
				"http.response.status_code": 304,
			},
		}},
		{`127.0.0.1 - - [03/Oct/2026:10:30:41 +0000] "bad request data" 400 12 "-" "-"`, "http_access", &expectedLogLine{
			Level:     "unknown",
			Message:   "HTTP request",
			Timestamp: testTimestamp(t, "2026-10-03T10:30:41Z"),
			Attributes: map[string]any{
				"source.address": "127.0.0.1", "http.request.line": "bad request data", "http.response.status_code": 400, "http.response.body.size": 12,
			},
		}},
		{`127.0.0.1 - - [03/Oct/2026:10:30:41 +0000] "bad request data" 400 12`, "http_access", &expectedLogLine{
			Level:     "unknown",
			Message:   "HTTP request",
			Timestamp: testTimestamp(t, "2026-10-03T10:30:41Z"),
			Attributes: map[string]any{
				"source.address": "127.0.0.1", "http.request.line": "bad request data", "http.response.status_code": 400, "http.response.body.size": 12,
			},
		}},
		{"2026-10-08 08:43:21.242 UTC [50] LOG:  " + postgresCheckpointMessage, "postgres", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-08T08:43:21.242Z"),
			Level:     "info",
			Message:   "checkpoint complete",
			Attributes: map[string]any{
				"pid":                        50,
				"checkpoint.buffers_written": 123, "checkpoint.buffers_percent": 4.5,
				"checkpoint.wal_files_added": 1, "checkpoint.wal_files_removed": 2, "checkpoint.wal_files_recycled": 3,
				"checkpoint.write_seconds": 1.234, "checkpoint.sync_seconds": 0.056, "checkpoint.total_seconds": 1.3,
				"checkpoint.sync_files": 7, "checkpoint.longest_sync_seconds": 0.020, "checkpoint.average_sync_seconds": 0.008,
				"checkpoint.distance_kb": 8192, "checkpoint.estimate_kb": 16384,
				"checkpoint.lsn": "1/ABCDEF01", "checkpoint.redo_lsn": "1/ABCDEF00",
			},
		}},
		{"2026-10-08 08:43:21.242 UTC [50] LOG:  checkpoint starting: shutdown immediate", "postgres", &expectedLogLine{
			Timestamp:  testTimestamp(t, "2026-10-08T08:43:21.242Z"),
			Level:      "info",
			Message:    "checkpoint starting",
			Attributes: map[string]any{"pid": 50, "checkpoint.reason": "shutdown immediate"},
		}},
		{`2026-10-08 08:43:21.242 UTC [50] LOG:  background worker "logical replication launcher" (PID 55) exited with exit code 1`, "postgres", &expectedLogLine{
			Timestamp:  testTimestamp(t, "2026-10-08T08:43:21.242Z"),
			Level:      "info",
			Message:    "background worker exited",
			Attributes: map[string]any{"pid": 50, "worker.name": "logical replication launcher", "worker.pid": 55, "worker.exit_code": 1},
		}},
		{"2026-10-08 08:43:21.242 UTC [50] LOG:  starting PostgreSQL 17.5 (Debian 17.5-1.pgdg130+1) on x86_64-pc-linux-gnu, compiled by gcc (Debian 14.2.0-19) 14.2.0, 64-bit", "postgres", &expectedLogLine{
			Timestamp:  testTimestamp(t, "2026-10-08T08:43:21.242Z"),
			Level:      "info",
			Message:    "starting PostgreSQL 17.5 (Debian 17.5-1.pgdg130+1) on x86_64-pc-linux-gnu, compiled by gcc (Debian 14.2.0-19) 14.2.0, 64-bit",
			Attributes: map[string]any{"pid": 50},
		}},
		{"2026-10-08 08:43:21.242 UTC [50] LOG:  " + strings.Replace(postgresCheckpointMessage, "redo lsn=1/ABCDEF00", "redo lsn=invalid", 1), "postgres", &expectedLogLine{
			Timestamp:  testTimestamp(t, "2026-10-08T08:43:21.242Z"),
			Level:      "info",
			Message:    strings.Replace(postgresCheckpointMessage, "redo lsn=1/ABCDEF00", "redo lsn=invalid", 1),
			Attributes: map[string]any{"pid": 50},
		}},
		{"2026-10-08 08:43:21.242 UTC [50] LOG:  checkpoint starting: ", "postgres", &expectedLogLine{
			Timestamp:  testTimestamp(t, "2026-10-08T08:43:21.242Z"),
			Level:      "info",
			Message:    "checkpoint starting: ",
			Attributes: map[string]any{"pid": 50},
		}},
		{`2026-10-08 08:43:21.242 UTC [50] LOG:  background worker "logical replication launcher" (PID 55) exited with exit code invalid`, "postgres", &expectedLogLine{
			Timestamp:  testTimestamp(t, "2026-10-08T08:43:21.242Z"),
			Level:      "info",
			Message:    `background worker "logical replication launcher" (PID 55) exited with exit code invalid`,
			Attributes: map[string]any{"pid": 50},
		}},
		{`2026/10/03 10:29:27.226036 [info] dnsproxy: ready a=b c="two words" a=last component=other`, "adguard", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T10:29:27.226036Z"), Level: "info", Message: "ready",
			Attributes: map[string]any{"component": "dnsproxy", "context": map[string]any{"a": "last", "c": "two words", "component": "other"}},
		}},
		{`JVB 2026-10-03 10:31:14.002 WARNING: [35] [a=b c="say \"hello\"" app=other] Source: ready`, "jvb", &expectedLogLine{
			Timestamp: testTimestamp(t, "2026-10-03T10:31:14.002Z"), Level: "warn", Message: "ready",
			Attributes: map[string]any{"app": "JVB", "thread_id": 35, "source": "Source", "context": map[string]any{"a": "b", "c": `say "hello"`, "app": "other"}},
		}},
	} {
		t.Run(tc.service+"/"+tc.line, func(t *testing.T) {
			record := NewLogLineInterpreter(tc.service).Parse(tc.line, LogRecordContext{ReadTimestamp: receivedAt})
			assert.Equal(t, &api.LogRecord{
				ReadTimestamp: receivedAt, LogTimestamp: tc.expected.Timestamp,
				Level: tc.expected.Level, Message: tc.expected.Message,
				Attributes: tc.expected.Attributes, Parsed: true,
			}, record)
		})
	}
}

func TestInterpret_InvalidRecordsFallBackToPlainText(t *testing.T) {
	for _, tc := range []struct{ service, line string }{
		{"vikunja", `time=2026-10-16T06:54:56.971Z level=ERROR msg="unterminated`},
		{"vikunja", `prefix time=2026-10-16T06:54:56.971Z level=INFO msg="ready"`},
		{"vikunja", `time=2026-10-16T06:54:56.971Z level=INFO`},
		{"prosody", "[register-setup Module reloaded successfully"},
		{"nextcloud", "AH00558: apache2 missing separator"},
		{"zulip", "Notice missing separator"},
		{"jitsi", "s6-rc: info service nginx: starting"},
		{"jitsi", "/package/admin/s6-overlay/libexec/preinit: info read-only root"},
		{"nextcloud", strings.Replace(nextcloudApplicationLog, `"level":2`, `"level":9`, 1)},
		{"nextcloud", strings.Replace(nextcloudApplicationLog, `"data":[]}`, `"data":[]`, 1)},
		{"nextcloud", strings.Replace(nextcloudApplicationLog, ", referer: ", " unexpected: ", 1)},
		{"apache", `[] [core:info] [pid 12] ready`},
		{"vaultwarden", `[][start][INFO] ready`},
		{"http_access", `127.0.0.1 - - [] "GET / HTTP/1.1" 200 12`},
		{"apache", `[Tue Feb 31 10:30:41 2026] [core:error] [pid 12] AH00163: failed`},
		{"http_access", `127.0.0.1 - - [31/Feb/2026:10:30:41 +0000] "GET / HTTP/1.1" 200 12 "-" "-"`},
		{"http_access", `127.0.0.1 - - [31/Feb/2026:10:30:41 +0000] "GET / HTTP/1.1" 200 12`},
		{"http_access", `127.0.0.1 - - [03/Oct/2026:10:30:41 +0000] "GET / HTTP/1.1" 200 999999999999999999999999999999 "-" "-"`},
		{"http_access", `127.0.0.1 - - [03/Oct/2026:10:30:41 +0000] "GET / HTTP/1.1" 200 999999999999999999999999999999`},
		{"apache", `[Sat Oct 03 10:30:41 2026] [core:info] [pid`},
		{"http_access", `127.0.0.1 - - [03/Oct/2026:10:30:41 +0000] "GET / HTTP/1.1" 200 12 "unterminated`},
		{"http_access", `prefix 127.0.0.1 - - [03/Oct/2026:10:30:41 +0000] "GET / HTTP/1.1" 200 12 "-" "-"`},
		{"http_access", `127.0.0.1 - - [03/Oct/2026:10:30:41 +0000] "GET / HTTP/1.1" 200 12 trailing data`},
		{"json", `{"timestamp":"unresolved","message":"hello"}`},
		{"json", `{"timestamp":"2026-10-03T10:29:55Z","level":"info","message":"hello","count":`},
		{"json", `[]`},
		{"json", `null`},
		{"json", `"hello"`},
		{"json", `123`},
		{"json", `true`},
		{"postgres", `2026-10-03 10:29:44.176 XYZ [1] LOG: ready`},
		{"adguard", `2026/10/03 10:29:27.226036 [info] ready a="unterminated`},
		{"jvb", `JVB 2026-10-03 10:31:14.002 WARNING: [35] [a=b broken] Source: failed`},
		{"postgres", "2026-10-03 10:29:44 UTC [x] LOG: ready"},
		{"postgres", "2026-10-03 10:29:44 UTC [99999999999999999999999] LOG: ready"},
		{"jvb", "JVB 2026-10-03 10:31:14 WARNING: [99999999999999999999999] Source: ready"},
		{"jitsi", "2026/10/03 10:31:09 [warn] 99999999999999999999999#12: ready"},
		{"jitsi", "2026/10/03 10:31:09 [warn] 12#99999999999999999999999: ready"},
		{"redis", "99999999999999999999999:C 03 Oct 2026 10:30:40 * ready"},
		{"mariadb", "2026-10-03 10:30:41 99999999999999999999999 [Note] ready"},
		{"forgejo", "2026/10/03 10:29:45 source [INFO ready"},
		{"jvb", "JVB 2026-10-03 10:31:14 WARNING: [x] Source: ready"},
		{"prosody", "2026-10-03 10:31:09 component info ready"},
		{"jitsi", "2026/10/03 10:31:09 [warn] 115#oops: ready"},
		{"redis", "x:C 03 Oct 2026 10:30:40 * ready"},
		{"redis", "1:C 03 Oct 2026 10:30:40 * <unterminated"},
		{"mariadb", "2026-10-03 10:30:41 x [Note] ready"},
		{"mariadb", "2026-10-03 10:30:41+01:00 [Note] [Entrypoint] ready"},
		{"vaultwarden", "[2026-10-03 10:31:17][start][INFO ready"},
		{"wgeasy", "2026-10-03T09:30:20Z Unknown ready"},
		{"hedgedoc", "2026-10-03T09:30:20Z unknown ready"},
		{"adguard", "2026/10/03 10:29:27 [info] ready a=one bare"},
		{"nextcloud", "[Sat Oct 03 10:30:41 2026] [core:info] [pid 99999999999999999999999] ready"},
		{"wordpress", `127.0.0.1 - - [03/Oct/2026:10:30:41 +0000] "GET / HTTP/1.1" xyz 12`},
	} {
		t.Run(tc.service+"/"+tc.line, func(t *testing.T) {
			record := NewLogLineInterpreter(tc.service).Parse(tc.line, LogRecordContext{ReadTimestamp: receivedAt})
			assert.Equal(t, &api.LogRecord{ReadTimestamp: receivedAt, Level: "unknown", Message: tc.line}, record)
		})
	}
}

func TestInterpret_NextcloudLevelsWithoutReferer(t *testing.T) {
	for level, normalizedLevel := range []string{"debug", "info", "warn", "error", "fatal"} {
		t.Run(normalizedLevel, func(t *testing.T) {
			line, _, _ := strings.Cut(nextcloudApplicationLog, ", referer: ")
			line = strings.Replace(line, "[2]", "["+strconv.Itoa(level)+"]", 1)
			line = strings.Replace(line, `"level":2`, `"level":`+strconv.Itoa(level), 1)
			record := NewLogLineInterpreter("nextcloud").Parse(line, LogRecordContext{ReadTimestamp: receivedAt})
			assert.Equal(t, true, record.Parsed)
			assert.Equal(t, normalizedLevel, record.Level)
			_, hasReferer := record.Attributes["http.request.referrer"]
			assert.Equal(t, false, hasReferer)
		})
	}
}
