package pages

import (
	"net/http"
	"server/apps_basic"
	"server/frontend/renderer"
	"server/tools"

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
	TemplateService   renderer.TemplateService
	OperationRegistry apps_basic.OperationRegistry
}

func (p *PageRendererImpl) PageCreationFailed(w http.ResponseWriter, err error) {
	u.Logger.Error(err)
	err = tools.WritePageCouldNotBeLoaded(w, http.StatusBadRequest)
	if err != nil {
		u.Logger.Error(err)
	}
}

func (p *PageRendererImpl) RenderPage(request PageRenderRequest) {
	auth := renderer.Auth{}
	user, err := getAuthFromContext(request.Request)
	if err == nil {
		auth.Name = user.Username
		auth.IsAdmin = user.IsAdmin
	}

	p.OperationRegistry.ClearFinishedAppOperations()

	renderedBytes, err := p.TemplateService.RenderPage(renderer.PageRenderRequest{
		PageName:             request.PageName,
		Content:              request.Content,
		Auth:                 auth,
		InfoIconRedirectPath: request.InfoIconRedirectPath,
		PageTitle:            request.PageTitle,
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
