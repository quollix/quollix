//go:build integration

package integration

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"server/configs"
	"server/di"
	"server/frontend/assets"
	"server/frontend/renderer"
	"server/tools"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

type renderContractPageData struct {
	Message string
}

func TestRenderContractPage_UsesProductionFrontendResources(t *testing.T) {
	withServerWorkingDirectory(t)

	testObjects := createFrontendRenderContractTestObjects(t)
	globalFrameCssPath := versionedAssetPath("global/frame", "css")
	globalGlobalJsPath := versionedAssetPath("global/global", "js")
	pageCssPath := versionedAssetPath("pages/render-contract/render-contract", "css")
	pageJsPath := versionedAssetPath("pages/render-contract/render-contract", "js")
	testObjects.configsServiceMock.EXPECT().GetBaseDomain().Return("example.test", nil)

	err := testObjects.templateService.ReloadTemplateFromFileSystem()
	assert.Nil(t, err)

	renderedBytes, err := testObjects.templateService.RenderPage(renderer.PageRenderRequest{
		PageName:             "render-contract",
		Content:              renderContractPageData{Message: "sample-message"},
		Auth:                 renderer.Auth{Name: "alice", IsAdmin: true},
		InfoIconRedirectPath: tools.Links.UsageDocs.Settings,
	})
	assert.Nil(t, err)

	rendered := string(renderedBytes)

	assert.True(t, strings.Contains(rendered, `<title>Render contract</title>`))
	assert.True(t, strings.Contains(rendered, `<h2 id="page-title">Render contract</h2>`))
	assert.True(t, strings.Contains(rendered, `<h3 id="render-contract-marker">Render contract sample</h3>`))
	assert.True(t, strings.Contains(rendered, `<p id="render-contract-message">sample-message</p>`))
	assert.True(t, strings.Contains(rendered, `<p id="render-contract-auth">alice</p>`))
	assert.True(t, strings.Contains(rendered, `<p id="render-contract-host">example.test</p>`))
	assert.True(t, strings.Contains(rendered, tools.Paths.FrontendSignIn))
	assert.True(t, strings.Contains(rendered, tools.Links.UsageDocs.Settings))
	assert.True(t, strings.Contains(rendered, globalFrameCssPath))
	assert.True(t, strings.Contains(rendered, globalGlobalJsPath))
	assert.True(t, strings.Contains(rendered, pageCssPath))
	assert.True(t, strings.Contains(rendered, pageJsPath))

	assert.True(t, testObjects.assetStore.Has(pageCssPath))
	assert.True(t, testObjects.assetStore.Has(pageJsPath))

	renderedCssBytes, cssFound := testObjects.assetStore.Get(pageCssPath)
	assert.True(t, cssFound)
	assert.True(t, strings.Contains(string(renderedCssBytes), "render-contract-css "+tools.ApplicationVersion))

	renderedJsBytes, jsFound := testObjects.assetStore.Get(pageJsPath)
	assert.True(t, jsFound)
	assert.True(t, strings.Contains(string(renderedJsBytes), `window.renderContractVersion = "`+tools.ApplicationVersion+`"`))
}

func TestRenderContractPage_UsesExplicitPageTitle(t *testing.T) {
	withServerWorkingDirectory(t)

	testObjects := createFrontendRenderContractTestObjects(t)
	testObjects.configsServiceMock.EXPECT().GetBaseDomain().Return("example.test", nil)

	err := testObjects.templateService.ReloadTemplateFromFileSystem()
	assert.Nil(t, err)

	renderedBytes, err := testObjects.templateService.RenderPage(renderer.PageRenderRequest{
		PageName:  "render-contract",
		PageTitle: "Contract Title",
		Content:   renderContractPageData{Message: "sample-message"},
		Auth:      renderer.Auth{Name: "alice", IsAdmin: true},
	})
	assert.Nil(t, err)

	rendered := string(renderedBytes)

	assert.True(t, strings.Contains(rendered, `<title>Contract Title</title>`))
	assert.True(t, strings.Contains(rendered, `<h2 id="page-title">Contract Title</h2>`))
}

func TestRenderContractPage_WhenPageMissing_ReturnsError(t *testing.T) {
	withServerWorkingDirectory(t)

	testObjects := createFrontendRenderContractTestObjects(t)

	err := testObjects.templateService.ReloadTemplateFromFileSystem()
	assert.Nil(t, err)

	renderedBytes, err := testObjects.templateService.RenderPage(renderer.PageRenderRequest{
		PageName:             "missing-page",
		Auth:                 renderer.Auth{Name: "alice", IsAdmin: true},
		InfoIconRedirectPath: tools.Links.UsageDocs.Settings,
	})

	assert.NotNil(t, err)
	assert.Nil(t, renderedBytes)
}

func TestRenderContractNoAssetsPage_DoesNotIncludePageLocalAssets_AndDoesNotShowDocsLinkForAdminWithoutPath(t *testing.T) {
	withServerWorkingDirectory(t)

	testObjects := createFrontendRenderContractTestObjects(t)
	globalFrameCssPath := versionedAssetPath("global/frame", "css")
	globalGlobalJsPath := versionedAssetPath("global/global", "js")
	pageCssPath := versionedAssetPath("pages/render-contract-no-assets/render-contract-no-assets", "css")
	pageJsPath := versionedAssetPath("pages/render-contract-no-assets/render-contract-no-assets", "js")
	testObjects.configsServiceMock.EXPECT().GetBaseDomain().Return("example.test", nil)

	err := testObjects.templateService.ReloadTemplateFromFileSystem()
	assert.Nil(t, err)

	renderedBytes, err := testObjects.templateService.RenderPage(renderer.PageRenderRequest{
		PageName: "render-contract-no-assets",
		Auth:     renderer.Auth{Name: "alice", IsAdmin: true},
	})
	assert.Nil(t, err)

	rendered := string(renderedBytes)

	assert.True(t, strings.Contains(rendered, `<h3 id="render-contract-no-assets-marker">Render contract no assets sample</h3>`))
	assert.True(t, strings.Contains(rendered, globalFrameCssPath))
	assert.True(t, strings.Contains(rendered, globalGlobalJsPath))
	assert.False(t, strings.Contains(rendered, pageCssPath))
	assert.False(t, strings.Contains(rendered, pageJsPath))
	assert.False(t, strings.Contains(rendered, tools.UsageDocsBaseUrl))
	assert.False(t, testObjects.assetStore.Has(pageCssPath))
	assert.False(t, testObjects.assetStore.Has(pageJsPath))
}

func TestRenderContractNoAssetsPage_DoesNotShowDocsLinkForNonAdminEvenWhenPathIsSet(t *testing.T) {
	withServerWorkingDirectory(t)

	testObjects := createFrontendRenderContractTestObjects(t)
	testObjects.configsServiceMock.EXPECT().GetBaseDomain().Return("example.test", nil)

	err := testObjects.templateService.ReloadTemplateFromFileSystem()
	assert.Nil(t, err)

	renderedBytes, err := testObjects.templateService.RenderPage(renderer.PageRenderRequest{
		PageName:             "render-contract-no-assets",
		Auth:                 renderer.Auth{Name: "alice", IsAdmin: false},
		InfoIconRedirectPath: tools.Links.UsageDocs.Settings,
	})
	assert.Nil(t, err)

	rendered := string(renderedBytes)

	assert.True(t, strings.Contains(rendered, `<p id="render-contract-no-assets-auth">alice</p>`))
	assert.False(t, strings.Contains(rendered, tools.Links.UsageDocs.Settings))
}

type frontendRenderContractTestObjects struct {
	templateService    *renderer.TemplateServiceImpl
	assetStore         *assets.AssetStoreImpl
	configsServiceMock *configs.ConfigsServiceMock
}

func createFrontendRenderContractTestObjects(t *testing.T) frontendRenderContractTestObjects {
	assetStore := &assets.AssetStoreImpl{AssetBytes: map[string][]byte{}}
	assetTagBuilder := &assets.AssetTagBuilderImpl{AssetStore: assetStore}
	templateEngine := &renderer.TemplateEngineImpl{AssetTagBuilder: assetTagBuilder}

	configsServiceMock := configs.NewConfigsServiceMock(t)

	templateService := &renderer.TemplateServiceImpl{
		Config:          di.NewGlobalConfig(),
		AssetTagBuilder: assetTagBuilder,
		CompiledPages:   map[string]*renderer.CompiledPageTemplate{},
		TemplateEngine:  templateEngine,
		AssetStore:      assetStore,
		FileSystem:      tools.FrontendResourceFilesystem,
		ConfigsService:  configsServiceMock,
	}

	return frontendRenderContractTestObjects{
		templateService:    templateService,
		assetStore:         assetStore,
		configsServiceMock: configsServiceMock,
	}
}

func withServerWorkingDirectory(t *testing.T) {
	serverDir, err := u.FindDir("server")
	assert.Nil(t, err)

	workingDir, err := os.Getwd()
	assert.Nil(t, err)

	err = os.Chdir(serverDir)
	assert.Nil(t, err)

	t.Cleanup(func() {
		chdirErr := os.Chdir(workingDir)
		assert.Nil(t, chdirErr)
	})
}

func versionedAssetPath(basePath, extension string) string {
	return fmt.Sprintf("%s/%s.%s.%s", tools.Paths.WebResourcesPath, basePath, tools.ApplicationVersion, extension)
}
