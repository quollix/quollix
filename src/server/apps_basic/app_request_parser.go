package apps_basic

import (
	"net"
	"net/http"
	"server/tools"
	"strings"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

type AppRequestParser interface {
	GetHostFromRequestHost(rawHost string) string
	GetAppNameFromRequestHost(requestHost, hostFromDatabase string) (string, error)
	GetQuerySecret(r *http.Request) (string, bool, bool)
}

const HostDoesNotEndWithDatabaseHostError = "host in HTTP request does not end with host in database, but it should have"

type AppRequestParserImpl struct{}

func (a *AppRequestParserImpl) GetHostFromRequestHost(rawHost string) string {
	host, port, err := net.SplitHostPort(rawHost)
	if err != nil {
		return rawHost
	}

	if port != "80" && port != "443" {
		return rawHost
	}

	return host
}

func (a *AppRequestParserImpl) GetAppNameFromRequestHost(requestHost, hostFromDatabase string) (string, error) {
	if !strings.HasSuffix(requestHost, hostFromDatabase) {
		return "", u.Logger.NewError(HostDoesNotEndWithDatabaseHostError, tools.RequestHostField, requestHost, tools.DatabaseHostField, hostFromDatabase)
	}

	return strings.TrimSuffix(requestHost, "."+hostFromDatabase), nil
}

func (a *AppRequestParserImpl) GetQuerySecret(r *http.Request) (secret string, isPresent bool, isValidValue bool) {
	secret = r.URL.Query().Get(tools.BrandAppQuerySecretName)
	if secret == "" {
		return "", false, true
	}

	err := validation.Validate("secret", validation.FieldSecret, secret)
	if err != nil {
		return "", false, false
	}
	return secret, true, true
}
