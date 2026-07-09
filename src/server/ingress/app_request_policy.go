package ingress

import (
	"fmt"
	"regexp"
	"server/tools"
	"strings"

	u "github.com/quollix/common/utils"
)

var (
	crossRequestsToAppsOnlyFromBrandAppOriginErrorMessage = fmt.Sprintf("cross request is only allowed from %s origin", u.OfficialBrandAppName)
	CrossRequestsToBrandAppNotAllowedErrorMessage         = fmt.Sprintf("cross request to %s is not allowed", u.OfficialBrandAppName)
)

type AppRequestPolicy interface {
	ValidateRequestOrigin(requestHost, originHost, baseDomain string) error
	IsRequestAddressedToAnApp(requestHost, baseDomain string) bool
}

type AppRequestPolicyImpl struct{}

func (p *AppRequestPolicyImpl) ValidateRequestOrigin(requestHost, originHost, baseDomain string) error {
	isCrossRequest := p.isCrossOriginRequest(requestHost, originHost)

	if p.IsRequestAddressedToAnApp(requestHost, baseDomain) {
		if isCrossRequest && !p.isOriginHostAllowed(originHost, baseDomain) {
			return u.Logger.NewError(crossRequestsToAppsOnlyFromBrandAppOriginErrorMessage)
		}
	} else if isCrossRequest {
		return u.Logger.NewError(CrossRequestsToBrandAppNotAllowedErrorMessage)
	}
	return nil
}

func (p *AppRequestPolicyImpl) isOriginHostAllowed(originHost string, baseDomain string) bool {
	return originHost == tools.BrandAppDomainPrefix+baseDomain
}

func (p *AppRequestPolicyImpl) isCrossOriginRequest(requestHost, originHost string) bool {
	return originHost != "" && originHost != requestHost
}

func (p *AppRequestPolicyImpl) IsRequestAddressedToAnApp(requestHost, baseDomain string) bool {
	u.Logger.Debug("checking if request is addressed to an app", tools.BaseDomainField, baseDomain, tools.RequestHostField, requestHost)
	pattern := fmt.Sprintf(`^.*\.%s(:\d+)?$`, baseDomain)
	re := regexp.MustCompile(pattern)
	isRequestAddressedToAnApp := re.MatchString(requestHost) && !strings.HasPrefix(requestHost, "quollix.")
	u.Logger.Debug("is request addressed to an app", tools.IsRequestAddressedToAnAppField, isRequestAddressedToAnApp)
	return isRequestAddressedToAnApp
}
