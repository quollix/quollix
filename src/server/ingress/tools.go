package ingress

import (
	"server/tools"
	"strings"

	u "github.com/quollix/common/utils"
)

const InvalidOriginHeader = "invalid origin header"

type CertificateTools interface {
	GetHostFromOriginHeaderValue(originHeaderValue string) (string, error)
}

type CertificateToolsImpl struct{}

func (c *CertificateToolsImpl) GetHostFromOriginHeaderValue(originHeaderValue string) (string, error) {
	var originHost string
	// Browsers set the origin header to "null" in some cases, which must be considered
	if originHeaderValue == "" || originHeaderValue == "null" {
		return "", nil
	} else {
		originParts := strings.Split(originHeaderValue, "://")
		if len(originParts) < 2 {
			return "", u.Logger.NewError(InvalidOriginHeader, tools.RequestOriginHostField, originHeaderValue)
		} else {
			originHost = originParts[1]
		}
		return originHost, nil
	}
}
