package assets

import (
	"fmt"
	"html/template"
	"server/tools"
	"strings"
)

type AssetTagBuilder interface {
	BuildHeadAssets() template.HTML
	BuildPageAssets(pageHtmlPath string) template.HTML
}

type AssetTagBuilderImpl struct {
	AssetStore AssetStore
}

func (b *AssetTagBuilderImpl) BuildHeadAssets() template.HTML {
	globalFiles := []string{"global/frame", "global/global"}
	return b.buildAssetsForBasePaths(globalFiles)
}

func (b *AssetTagBuilderImpl) BuildPageAssets(pageHtmlPath string) template.HTML {
	basePath := strings.TrimSuffix(pageHtmlPath, ".html")
	return b.buildAssetsForBasePaths([]string{basePath})
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
