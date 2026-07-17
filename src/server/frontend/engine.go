package frontend

import (
	frontendpages "server/frontend/pages"
	"server/tools"

	"github.com/quollix/common/frontend"
	api "github.com/quollix/common/quollix/api"
)

func NewFrontendEngine() (frontend.EngineService, error) {
	return frontend.NewEngine(frontend.Config{
		FrontendFolderPath: "frontend",
		Version:            tools.ApplicationVersion,
		Static: frontendpages.StaticTemplateGlobals{
			Paths:    api.Paths,
			Links:    tools.Links,
			Policies: api.Policies,
		},
	})
}
