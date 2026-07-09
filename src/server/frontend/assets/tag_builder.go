package assets

import (
	"fmt"
	"html/template"
	"server/tools"
	"strings"
)

type AssetTagBuilder interface {
	BuildHeadAssets(pageHtmlPath string) template.HTML
}

type AssetTagBuilderImpl struct {
	AssetStore AssetStore
}

func (b *AssetTagBuilderImpl) BuildHeadAssets(pageHtmlPath string) template.HTML {
	basePaths := []string{"global/frame", "global/global"}
	basePath := strings.TrimSuffix(pageHtmlPath, ".html")
	basePaths = append(basePaths, basePath)
	return b.buildAssetsForBasePaths(basePaths)
}

func (b *AssetTagBuilderImpl) buildAssetsForBasePaths(files []string) template.HTML {
	var builderString strings.Builder

	for _, file := range files {
		builderString.WriteString(b.getCssLinkTag(file))
		builderString.WriteString(b.getJsScriptTag(file))
	}

	return template.HTML(builderString.String()) // #nosec G203 (CWE-79): The used method does not auto-escape HTML. This can potentially lead to 'Cross-site Scripting' vulnerabilities, in case the attacker controls the input.
}

func (b *AssetTagBuilderImpl) getJsScriptTag(file string) string {
	jsInjectedPath := b.AssetStore.GetVersionedInjectedWebResourcePath(tools.FrontendResourcesPathWithLeadingSlash, file, "js")
	if b.AssetStore.Has(jsInjectedPath) {
		return fmt.Sprintf(`<script type="module" src="%s"></script>`+"\n", jsInjectedPath)
	}
	return ""
}

func (b *AssetTagBuilderImpl) getCssLinkTag(file string) string {
	cssInjectedPath := b.AssetStore.GetVersionedInjectedWebResourcePath(tools.FrontendResourcesPathWithLeadingSlash, file, "css")
	if b.AssetStore.Has(cssInjectedPath) {
		return fmt.Sprintf(`<link rel="stylesheet" href="%s">`+"\n", cssInjectedPath)
	}
	return ""
}
