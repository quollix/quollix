package apps_basic

import (
	"net/http"
	"net/http/httptest"
	"server/configs"
	"testing"

	"github.com/quollix/common/assert"
)

type appRequestResolverTestDependencies struct {
	configsService   *configs.ConfigsServiceMock
	appRepo          *AppRepositoryMock
	appRequestParser *AppRequestParserMock
	resolver         *AppRequestResolverImpl
	request          *http.Request
	app              *AppRequestData
}

func newAppRequestResolverTestDependencies(t *testing.T, rawURL string) *appRequestResolverTestDependencies {
	configsService := configs.NewConfigsServiceMock(t)
	appRepo := NewAppRepositoryMock(t)
	appRequestParser := NewAppRequestParserMock(t)
	app := &AppRequestData{
		Maintainer: "maintainer",
		AppName:    "sample-app",
	}
	return &appRequestResolverTestDependencies{
		configsService:   configsService,
		appRepo:          appRepo,
		appRequestParser: appRequestParser,
		resolver: &AppRequestResolverImpl{
			ConfigsService:   configsService,
			AppRepo:          appRepo,
			AppRequestParser: appRequestParser,
		},
		request: httptest.NewRequest(http.MethodGet, rawURL, nil),
		app:     app,
	}
}

func TestAppRequestResolver_ResolveAppRequestReturnsKnownApp(t *testing.T) {
	deps := newAppRequestResolverTestDependencies(t, "https://sample-app.example.com/docs/view")
	deps.appRequestParser.EXPECT().GetHostFromRequestHost("sample-app.example.com").Return("sample-app.example.com")
	deps.configsService.EXPECT().GetBaseDomain().Return("example.com", nil)
	deps.appRequestParser.EXPECT().GetAppNameFromRequestHost("sample-app.example.com", "example.com").Return("sample-app", nil)
	deps.appRepo.EXPECT().DoesAppExist("sample-app").Return(true, nil)
	deps.appRepo.EXPECT().GetAppRequestData("sample-app").Return(deps.app, nil)

	resolvedAppRequest, err := deps.resolver.ResolveAppRequest(deps.request)

	assert.Nil(t, err)
	assert.Equal(t, "example.com", resolvedAppRequest.BaseDomain)
	assert.Equal(t, "sample-app", resolvedAppRequest.AppName)
	assert.Equal(t, true, resolvedAppRequest.AppExists)
	assert.Equal(t, deps.app, resolvedAppRequest.App)
}

func TestAppRequestResolver_ResolveAppRequestReturnsUnknownApp(t *testing.T) {
	deps := newAppRequestResolverTestDependencies(t, "https://unknown-app.example.com/docs/view")
	deps.appRequestParser.EXPECT().GetHostFromRequestHost("unknown-app.example.com").Return("unknown-app.example.com")
	deps.configsService.EXPECT().GetBaseDomain().Return("example.com", nil)
	deps.appRequestParser.EXPECT().GetAppNameFromRequestHost("unknown-app.example.com", "example.com").Return("unknown-app", nil)
	deps.appRepo.EXPECT().DoesAppExist("unknown-app").Return(false, nil)

	resolvedAppRequest, err := deps.resolver.ResolveAppRequest(deps.request)

	assert.Nil(t, err)
	assert.Equal(t, "example.com", resolvedAppRequest.BaseDomain)
	assert.Equal(t, "unknown-app", resolvedAppRequest.AppName)
	assert.Equal(t, false, resolvedAppRequest.AppExists)
	assert.Nil(t, resolvedAppRequest.App)
}
