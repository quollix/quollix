package component

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

const (
	SampleUsername     = "user"
	SampleUserPassword = "userpassword"
	SampleUserEmail    = "user@example.invalid"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // #nosec G402: component tests intentionally connect to the local test certificate
		},
	},
}

type QuollixClient struct {
	Parent u.ComponentClient
	T      *testing.T

	Auth         *QuollixAuthClient
	Users        *QuollixUsersClient
	Apps         *QuollixAppsClient
	Backups      *QuollixBackupsClient
	Settings     *QuollixSettingsClient
	Maintenance  *QuollixMaintenanceClient
	Certificates *QuollixCertificatesClient
	Frontend     *QuollixFrontendClient
	Test         *QuollixTestClient
	Content      *QuollixContentClient
	Email        *EmailClient
	Groups       *GroupsClient
}

func GetClientAndLogin(t *testing.T) *QuollixClient {
	client := GetQuollixClient(t)
	assert.Nil(t, client.Auth.Login(tools.DefaultAdminName, tools.DefaultAdminPassword))
	return client
}

func GetQuollixClient(t *testing.T) *QuollixClient {
	quollix := &QuollixClient{
		Parent: u.ComponentClient{
			Cookie:            nil,
			SetCookieHeader:   true,
			RootUrl:           fmt.Sprintf("https://%slocalhost", tools.BrandAppDomainPrefix),
			VerifyCertificate: false,
		},
		T: t,
	}
	quollix.Auth = &QuollixAuthClient{quollix: quollix}
	quollix.Users = &QuollixUsersClient{quollix: quollix}
	quollix.Apps = &QuollixAppsClient{quollix: quollix}
	quollix.Backups = &QuollixBackupsClient{quollix: quollix}
	quollix.Settings = &QuollixSettingsClient{quollix: quollix}
	quollix.Maintenance = &QuollixMaintenanceClient{quollix: quollix}
	quollix.Certificates = &QuollixCertificatesClient{quollix: quollix}
	quollix.Frontend = &QuollixFrontendClient{quollix: quollix}
	quollix.Test = &QuollixTestClient{quollix: quollix}
	quollix.Content = &QuollixContentClient{quollix: quollix}
	quollix.Email = &EmailClient{quollix: quollix}
	quollix.Groups = &GroupsClient{quollix: quollix}

	return quollix
}
