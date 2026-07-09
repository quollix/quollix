package tools

import (
	"fmt"
	"time"
)

var (
	SampleAppDockerNetwork = fmt.Sprintf("%s_%s", SampleMaintainer, SampleApp)
	SampleAppDockerVolume  = fmt.Sprintf("%s_%s_data", SampleMaintainer, SampleApp)
	SampleAppContainerName = fmt.Sprintf("%s_%s_%s", SampleMaintainer, SampleApp, SampleApp)

	SampleAppCreationTimestamp         = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	SampleAppVersion0CreationTimestamp = SampleAppCreationTimestamp.Add(-24 * time.Hour)
	SampleAppVersion1CreationTimestamp = SampleAppCreationTimestamp.Add(-time.Hour)
	SampleAppVersion2CreationTimestamp = SampleAppCreationTimestamp.Add(+time.Hour)
)

const (
	SampleMaintainer = "samplemaintainer"
	SampleApp        = "sampleapp"

	SampleAppVersion0Name = "0.0"
	SampleAppVersion1Name = "1.0"
	SampleAppVersion2Name = "2.0"

	TestSshServerHost             = "dummy_backup_server"
	TestSshServerPort             = "2222"
	TestSshServerBackupsDirectory = "/config/backups" // the actual backups are store also in the config folder, which is a little odd. maybe improve naming? where is the /config path coming from?

	SampleAppVersion0ComposeYAML = `services:
  sampleapp:
    image: sampleapp:local
    container_name: samplemaintainer_sampleapp_sampleapp
    environment:
      - VERSION=1.0
      - SERVER_URL=https://sampleapp.${BASE_DOMAIN}
      - OIDC_CLIENT_ID=${CLIENT_ID}
      - OIDC_CLIENT_SECRET=${CLIENT_SECRET}
      - IANA_TIMEZONE=${IANA_TIMEZONE}
      - PORT=3000
    labels:
      quollix.port: 3000
`

	SampleAppVersion1ComposeYAML = `services:
  sampleapp:
    image: sampleapp:local
    container_name: samplemaintainer_sampleapp_sampleapp
    environment:
      - VERSION=1.0
      - SERVER_URL=https://sampleapp.${BASE_DOMAIN}
      - OIDC_CLIENT_ID=${CLIENT_ID}
      - OIDC_CLIENT_SECRET=${CLIENT_SECRET}
      - IANA_TIMEZONE=${IANA_TIMEZONE}
      - PORT=3000
    volumes:
      - samplemaintainer_sampleapp_data:/data
    labels:
      quollix.port: 3000

volumes:
  samplemaintainer_sampleapp_data:
`

	SampleAppVersion2ComposeYAML = `services:
  sampleapp:
    image: sampleapp:local
    container_name: samplemaintainer_sampleapp_sampleapp
    environment:
      - VERSION=2.0
      - SERVER_URL=https://sampleapp.${BASE_DOMAIN}
      - OIDC_CLIENT_ID=${CLIENT_ID}
      - OIDC_CLIENT_SECRET=${CLIENT_SECRET}
      - IANA_TIMEZONE=${IANA_TIMEZONE}
      - PORT=3001
    volumes:
      - samplemaintainer_sampleapp_data:/data
    labels:
      quollix.port: 3001

volumes:
  samplemaintainer_sampleapp_data:
`
)
