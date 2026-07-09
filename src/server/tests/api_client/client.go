package api_client

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"server/tools"

	u "github.com/quollix/common/utils"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // #nosec G402: tests intentionally connect to the local test certificate
		},
	},
}

type QuollixClient struct {
	Parent *u.ComponentClient

	Auth          *AuthClient
	Users         *UsersClient
	Apps          *QuollixAppsClient
	Backups       *QuollixBackupsClient
	Settings      *QuollixSettingsClient
	Maintenance   *QuollixMaintenanceClient
	Certificates  *QuollixCertificatesClient
	Frontend      *QuollixFrontendClient
	Test          *QuollixTestClient
	AppAccess     *AppAccessClient
	Email         *EmailClient
	Groups        *GroupsClient
	OidcProviders *OidcAuthProvidersClient
	OidcClients   *OidcRelyingPartiesClient
}

func NewQuollixClient() *QuollixClient {
	return NewQuollixClientForRootUrl(fmt.Sprintf("https://%slocalhost", tools.BrandAppDomainPrefix))
}

func NewQuollixClientForRootUrl(rootUrl string) *QuollixClient {
	parent := &u.ComponentClient{
		Cookie:            nil,
		SetCookieHeader:   true,
		RootUrl:           rootUrl,
		VerifyCertificate: false,
	}
	return NewQuollixClientWithParent(parent)
}

func NewQuollixClientWithParent(parent *u.ComponentClient) *QuollixClient {
	client := &QuollixClient{Parent: parent}
	client.Auth = &AuthClient{quollix: client}
	client.Users = &UsersClient{quollix: client}
	client.Apps = &QuollixAppsClient{quollix: client}
	client.Backups = &QuollixBackupsClient{quollix: client}
	client.Settings = &QuollixSettingsClient{quollix: client}
	client.Maintenance = &QuollixMaintenanceClient{quollix: client}
	client.Certificates = &QuollixCertificatesClient{quollix: client}
	client.Frontend = &QuollixFrontendClient{quollix: client}
	client.Test = &QuollixTestClient{quollix: client}
	client.AppAccess = &AppAccessClient{quollix: client}
	client.Email = &EmailClient{quollix: client}
	client.Groups = &GroupsClient{quollix: client}
	client.OidcProviders = &OidcAuthProvidersClient{quollix: client}
	client.OidcClients = &OidcRelyingPartiesClient{quollix: client}
	return client
}
