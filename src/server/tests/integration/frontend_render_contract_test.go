//go:build integration

package integration

import (
	"fmt"
	"os"
	"server/configs"
	"server/di"
	"server/frontend/assets"
	"server/frontend/renderer"
	"server/tools"
	"strings"
	"testing"

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
	testObjects.configsRepoMock.EXPECT().GetConfig(configs.ConfigKeys.ServerHost).Return("example.test", nil)

	err := testObjects.templateService.ReloadTemplateFromFileSystem()
	assert.Nil(t, err)

	renderedBytes, err := testObjects.templateService.RenderPage(
		"render-contract",
		renderContractPageData{Message: "sample-message"},
		renderer.Auth{Name: "alice", IsAdmin: true},
		tools.Links.UsageDocs.Settings,
	)
	assert.Nil(t, err)

	rendered := string(renderedBytes)

	assert.True(t, strings.Contains(rendered, `<h3 id="render-contract-marker">Render contract sample</h3>`))
	assert.True(t, strings.Contains(rendered, `<p id="render-contract-message">sample-message</p>`))
	assert.True(t, strings.Contains(rendered, `<p id="render-contract-auth">alice</p>`))
	assert.True(t, strings.Contains(rendered, `<p id="render-contract-host">example.test</p>`))
	assert.True(t, strings.Contains(rendered, tools.Paths.FrontendLogin))
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

func TestRenderContractPage_WhenPageMissing_ReturnsError(t *testing.T) {
	withServerWorkingDirectory(t)

	testObjects := createFrontendRenderContractTestObjects(t)

	err := testObjects.templateService.ReloadTemplateFromFileSystem()
	assert.Nil(t, err)

	renderedBytes, err := testObjects.templateService.RenderPage(
		"missing-page",
		nil,
		renderer.Auth{Name: "alice", IsAdmin: true},
		tools.Links.UsageDocs.Settings,
	)

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
	testObjects.configsRepoMock.EXPECT().GetConfig(configs.ConfigKeys.ServerHost).Return("example.test", nil)

	err := testObjects.templateService.ReloadTemplateFromFileSystem()
	assert.Nil(t, err)

	renderedBytes, err := testObjects.templateService.RenderPage(
		"render-contract-no-assets",
		nil,
		renderer.Auth{Name: "alice", IsAdmin: true},
		"",
	)
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
	testObjects.configsRepoMock.EXPECT().GetConfig(configs.ConfigKeys.ServerHost).Return("example.test", nil)

	err := testObjects.templateService.ReloadTemplateFromFileSystem()
	assert.Nil(t, err)

	renderedBytes, err := testObjects.templateService.RenderPage(
		"render-contract-no-assets",
		nil,
		renderer.Auth{Name: "alice", IsAdmin: false},
		tools.Links.UsageDocs.Settings,
	)
	assert.Nil(t, err)

	rendered := string(renderedBytes)

	assert.True(t, strings.Contains(rendered, `<p id="render-contract-no-assets-auth">alice</p>`))
	assert.False(t, strings.Contains(rendered, tools.Links.UsageDocs.Settings))
}

type frontendRenderContractTestObjects struct {
	templateService *renderer.TemplateServiceImpl
	assetStore      *assets.AssetStoreImpl
	configsRepoMock *configs.ConfigsRepositoryMock
}

func createFrontendRenderContractTestObjects(t *testing.T) frontendRenderContractTestObjects {
	assetStore := &assets.AssetStoreImpl{AssetBytes: map[string][]byte{}}
	assetTagBuilder := &assets.AssetTagBuilderImpl{AssetStore: assetStore}
	templateEngine := &renderer.TemplateEngineImpl{AssetTagBuilder: assetTagBuilder}

	configsRepoMock := configs.NewConfigsRepositoryMock(t)

	templateService := &renderer.TemplateServiceImpl{
		Config:            di.NewGlobalConfig(),
		AssetTagBuilder:   assetTagBuilder,
		CompiledPages:     map[string]*renderer.CompiledPageTemplate{},
		TemplateEngine:    templateEngine,
		AssetStore:        assetStore,
		FileSystem:        tools.FrontendResourceFilesystem,
		ConfigsRepository: configsRepoMock,
	}

	return frontendRenderContractTestObjects{
		templateService: templateService,
		assetStore:      assetStore,
		configsRepoMock: configsRepoMock,
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
