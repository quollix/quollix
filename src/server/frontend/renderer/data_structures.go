package renderer

import (
	"html/template"
	"server/tools"
)

type CompiledPageTemplate struct {
	PageName  string
	PagePath  string
	FramePath string
	Template  *template.Template
}

type Auth struct {
	Name    string
	IsAdmin bool
}

type GlobalPageData struct {
	Paths               tools.PathsType
	Links               tools.LinksType
	Policies            any
	Auth                Auth
	Data                any
	InfoIconRedirectUrl string
	Title               string
	Config              *tools.GlobalConfig
	HeadAssets          template.HTML
	ApplicationVersion  string
	Host                string
}
