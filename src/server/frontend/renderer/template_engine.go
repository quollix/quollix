package renderer

import (
	"bytes"
	htmltemplate "html/template"
	"server/frontend/assets"
	"server/tools"
	"strings"
	texttemplate "text/template"

	u "github.com/quollix/common/utils"
)

var textTemplateData = map[string]any{
	"Paths":              tools.Paths,
	"Links":              tools.Links,
	"Policies":           tools.Policies,
	"ApplicationVersion": tools.ApplicationVersion,
}

type TemplateEngine interface {
	CompileHtml(frameTemplateText, pageTemplateText string) (*htmltemplate.Template, error)
	CompileText(templateName string, templateText string) ([]byte, error)
	PreprocessPageTemplate(content, pageHtmlPath string) string
}

type TemplateEngineImpl struct {
	AssetTagBuilder assets.AssetTagBuilder
}

func (c *TemplateEngineImpl) CompileHtml(frameTemplateText, pageTemplateText string) (*htmltemplate.Template, error) {
	tmpl, err := htmltemplate.New("").Parse(frameTemplateText)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	tmpl, err = tmpl.Parse(pageTemplateText)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	return tmpl, nil
}

func (e *TemplateEngineImpl) CompileText(templateName, templateText string) ([]byte, error) {
	tmpl, err := texttemplate.New(templateName).Option("missingkey=error").Parse(templateText)
	if err != nil {
		return nil, u.Logger.NewError(err.Error(), "template_name", templateName)
	}
	var outputBuffer bytes.Buffer
	if err := tmpl.Execute(&outputBuffer, textTemplateData); err != nil {
		return nil, u.Logger.NewError(err.Error(), "template_name", templateName)
	}
	return outputBuffer.Bytes(), nil
}

func (e *TemplateEngineImpl) PreprocessPageTemplate(content, pageHtmlPath string) string {
	var builder strings.Builder
	builder.WriteString(`{{define "content"}}` + "\n")
	builder.WriteString(string(e.AssetTagBuilder.BuildPageAssets(pageHtmlPath)))
	builder.WriteString(content + "\n")
	builder.WriteString(`{{end}}` + "\n")
	return builder.String()
}
