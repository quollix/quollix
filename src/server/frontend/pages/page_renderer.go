package pages

import (
	"net/http"

	"server/apps_basic"
	"server/configs"
	"server/tools"

	"github.com/quollix/common/frontend"
	u "github.com/quollix/common/utils"
)

type PageRenderer interface {
	RenderPage(request PageRenderRequest)
	PageCreationFailed(w http.ResponseWriter, err error)
}

type PageRenderRequest struct {
	ResponseWriter       http.ResponseWriter
	Request              *http.Request
	PageName             string
	Content              any
	InfoIconRedirectPath string
	PageTitle            string
}

type PageRendererImpl struct {
	Config            *tools.GlobalConfig
	ConfigsService    configs.ConfigsService
	FrontendEngine    frontend.EngineService
	OperationRegistry apps_basic.OperationRegistry
}

type Auth struct {
	Name    string
	IsAdmin bool
}

type PageGlobals struct {
	Auth                Auth
	Config              *tools.GlobalConfig
	Host                string
	InfoIconRedirectUrl string
}

type StaticTemplateGlobals struct {
	Paths    tools.PathsType
	Links    tools.LinksType
	Policies any
}

func (p *PageRendererImpl) PageCreationFailed(w http.ResponseWriter, err error) {
	u.Logger.Error(err)
	err = tools.WritePageCouldNotBeLoaded(w, http.StatusBadRequest)
	if err != nil {
		u.Logger.Error(err)
	}
}

func (p *PageRendererImpl) RenderPage(request PageRenderRequest) {
	auth := Auth{}
	user, err := getAuthFromContext(request.Request)
	if err == nil {
		auth.Name = user.Username
		auth.IsAdmin = user.IsAdmin
	}

	p.OperationRegistry.ClearFinishedAppOperations()

	host, err := p.ConfigsService.GetBaseDomain()
	if err != nil {
		p.PageCreationFailed(request.ResponseWriter, err)
		return
	}

	renderedBytes, err := p.FrontendEngine.Render(frontend.RenderRequest{
		PageName:  request.PageName,
		PageTitle: request.PageTitle,
		Globals: PageGlobals{
			Auth:                auth,
			Config:              p.Config,
			Host:                host,
			InfoIconRedirectUrl: infoIconRedirectUrl(request.InfoIconRedirectPath, auth.IsAdmin),
		},
		Page: request.Content,
	})
	if err != nil {
		p.PageCreationFailed(request.ResponseWriter, err)
		return
	}

	request.ResponseWriter.Header().Set("Cache-Control", "no-cache")
	request.ResponseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = request.ResponseWriter.Write(renderedBytes) // #nosec G705: renderedBytes come from server-side templates, not raw user-controlled HTML injection
	if err != nil {
		u.Logger.Error(err)
	}
}

func infoIconRedirectUrl(infoIconRedirectPath string, isAdmin bool) string {
	if infoIconRedirectPath != "" && isAdmin {
		return infoIconRedirectPath
	}
	return ""
}

func getAuthFromContext(r *http.Request) (*tools.User, error) {
	if r.Context() == nil {
		return nil, u.Logger.NewError("request context is nil")
	}

	val := r.Context().Value(tools.AuthKey)
	if val == nil {
		return nil, u.Logger.NewError("auth not found in context")
	}

	user, ok := val.(tools.User)
	if !ok {
		return nil, u.Logger.NewError("auth context value is of invalid type")
	}

	return &user, nil
}
