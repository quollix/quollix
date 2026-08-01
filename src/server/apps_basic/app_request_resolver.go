package apps_basic

import (
	"net/http"
	"server/configs"
)

type AppRequestResolver interface {
	ResolveAppRequest(r *http.Request) (*ResolvedAppRequest, error)
}

type ResolvedAppRequest struct {
	BaseDomain string
	AppName    string
	App        *AppRequestData
	AppExists  bool
}

type AppRequestResolverImpl struct {
	ConfigsService   configs.ConfigsService
	AppRepo          AppRepository
	AppRequestParser AppRequestParser
}

func (a *AppRequestResolverImpl) ResolveAppRequest(r *http.Request) (*ResolvedAppRequest, error) {
	requestHost := a.AppRequestParser.GetHostFromRequestHost(r.Host)
	baseDomain, err := a.ConfigsService.GetBaseDomain()
	if err != nil {
		return nil, err
	}

	appName, err := a.AppRequestParser.GetAppNameFromRequestHost(requestHost, baseDomain)
	if err != nil {
		return nil, err
	}

	appExists, err := a.AppRepo.DoesAppExist(appName)
	if err != nil {
		return nil, err
	}
	if !appExists {
		return &ResolvedAppRequest{
			BaseDomain: baseDomain,
			AppName:    appName,
			AppExists:  false,
		}, nil
	}

	app, err := a.AppRepo.GetAppRequestData(appName)
	if err != nil {
		return nil, err
	}

	return &ResolvedAppRequest{
		BaseDomain: baseDomain,
		AppName:    appName,
		App:        app,
		AppExists:  true,
	}, nil
}
