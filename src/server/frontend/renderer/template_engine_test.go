package renderer

import (
	"bytes"
	"server/frontend/assets"
	"testing"

	"github.com/quollix/common/assert"
)

// We render "<" to ensure that html/template and text/template cannot be accidentally swapped: html/template escapes it ("\u003c"), text/template does not.

var templateEngine = &TemplateEngineImpl{}

func TestTemplateEngineImpl_CompileHtml_EscapesInScriptContext(t *testing.T) {
	frameTemplateText := `{{define "frame"}}<h1>hello</h1>{{template "content" .}}{{end}}`
	pageTemplateText := `{{define "content"}}<script>const lessThan = "{{ "<" }}";</script>{{end}}`

	compiledTemplate, err := templateEngine.CompileHtml(frameTemplateText, pageTemplateText)
	assert.Nil(t, err)

	var outputBuffer bytes.Buffer
	err = compiledTemplate.ExecuteTemplate(&outputBuffer, "frame", nil)
	assert.Nil(t, err)
	assert.Equal(t, `<h1>hello</h1><script>const lessThan = "\u003c";</script>`, outputBuffer.String())
}

func TestTemplateEngineImpl_CompileText_DoesNotEscapeInScriptContext(t *testing.T) {
	templateText := `<script>{{ with . }}const loginPath = "{{ .Paths.FrontendSignIn }}"; const lessThan = "<";{{ end }}</script>`
	outputBytes, err := templateEngine.CompileText("sample", templateText)
	assert.Nil(t, err)
	assert.Equal(t, `<script>const loginPath = "/sign-in"; const lessThan = "<";</script>`, string(outputBytes))
}

func TestTemplateHelperImpl_PreprocessPageTemplate(t *testing.T) {
	assetTagBuilderMock := assets.NewAssetTagBuilderMock(t)
	engine := &TemplateEngineImpl{
		AssetTagBuilder: assetTagBuilderMock,
	}

	pageHtmlPath := "apps/installed.html"

	assetTagBuilderMock.EXPECT().BuildPageAssets(pageHtmlPath).Return(`<script type="module" src="/mock.js"></script>` + "\n")
	actual := engine.PreprocessPageTemplate("<h1>Installed</h1>", pageHtmlPath)

	expected := `{{define "content"}}
<script type="module" src="/mock.js"></script>
<h1>Installed</h1>
{{end}}
`

	assert.Equal(t, expected, actual)
	assetTagBuilderMock.AssertExpectations(t)
}
