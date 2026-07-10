package frontend

import (
	frontendpages "server/frontend/pages"
	"server/tools"

	"github.com/quollix/common/frontend"
)

func NewFrontendEngine() (frontend.EngineService, error) {
	return frontend.NewEngine(frontend.Config{
		FrontendFolderPath: "frontend",
		Version:            tools.ApplicationVersion,
		Static: frontendpages.StaticTemplateGlobals{
			Paths:    tools.Paths,
			Links:    tools.Links,
			Policies: tools.Policies,
		},
	})
}
