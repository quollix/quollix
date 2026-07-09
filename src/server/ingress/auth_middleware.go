package ingress

import (
	"net/http"
	"server/apps_basic"
	"server/configs"

	"github.com/go-chi/chi/v5"
	u "github.com/quollix/common/utils"
)

type SetupHandler struct {
	AppManager      apps_basic.AppService
	ConfigsRepo     configs.ConfigsRepository
	Router          chi.Router
	AppRequestProxy *apps_basic.AppRequestProxy
	AuthService     AuthService
}

func (s *SetupHandler) ProxyMiddleware(w http.ResponseWriter, r *http.Request) {
	originHeaderValue := r.Header.Get("Origin")
	isAddressedToApp, err := s.AuthService.IsRequestAddressedToAnApp(r.Host, originHeaderValue)
	if err != nil {
		u.WriteResponseError(w, u.MapOf(CrossRequestsToBrandAppNotAllowedErrorMessage, InvalidOriginHeader), err)
		return
	}

	if isAddressedToApp {
		s.AppRequestProxy.ProxyRequestToTheAppsDockerContainer(w, r)
	} else {
		s.Router.ServeHTTP(w, r)
	}
}

type AuthService interface {
	IsRequestAddressedToAnApp(requestHost, originHeaderValue string) (bool, error)
}

type AuthServiceImpl struct {
	ConfigsService   configs.ConfigsService
	AppRequestPolicy AppRequestPolicy
	CertificateTools CertificateTools
}

func (a *AuthServiceImpl) IsRequestAddressedToAnApp(requestHost, originHeaderValue string) (bool, error) {
	baseDomain, err := a.ConfigsService.GetBaseDomain()
	if err != nil {
		return false, err
	}
	requestOriginHost, err := a.CertificateTools.GetHostFromOriginHeaderValue(originHeaderValue)
	if err != nil {
		return false, err
	}

	err = a.AppRequestPolicy.ValidateRequestOrigin(requestHost, requestOriginHost, baseDomain)
	if err != nil {
		return false, err
	}

	return a.AppRequestPolicy.IsRequestAddressedToAnApp(requestHost, baseDomain), nil
}
