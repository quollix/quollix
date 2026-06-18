package renderer

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"server/configs"
	"server/frontend/assets"
	"server/tools"
	"strings"

	u "github.com/quollix/common/utils"
)

const (
	framePageName        = "frame"
	framePathInWebFolder = tools.FrontendFramePathInResources
)

type TemplateService interface {
	ReloadTemplateFromFileSystem() error
	RenderPage(pageName string, content any, auth Auth, infoIconRedirectPath string) ([]byte, error)
}

type TemplateServiceImpl struct {
	Config          *tools.GlobalConfig
	AssetTagBuilder assets.AssetTagBuilder

	CompiledPages     map[string]*CompiledPageTemplate `wire:"-"`
	TemplateEngine    TemplateEngine
	AssetStore        assets.AssetStore
	FileSystem        fs.FS
	ConfigsRepository configs.ConfigsRepository
}

func (s *TemplateServiceImpl) ReloadTemplateFromFileSystem() error {
	s.AssetStore.Clear()
	if err := s.walkWebResourcesAndAddToAssetStore("js"); err != nil {
		return err
	}
	if err := s.walkWebResourcesAndAddToAssetStore("css"); err != nil {
		return err
	}
	if err := s.walkWebResourcesAndAddToAssetStore("woff2"); err != nil {
		return err
	}

	frameBytes, err := fs.ReadFile(s.FileSystem, framePathInWebFolder)
	if err != nil {
		return u.Logger.NewError(err.Error(), "path", framePathInWebFolder)
	}
	frameText := string(frameBytes)

	s.CompiledPages, err = s.walkHtmlTemplatesAndCompile(frameText)
	return err
}

func (s *TemplateServiceImpl) RenderPage(pageName string, content any, auth Auth, infoIconRedirectPath string) ([]byte, error) {
	compiledPage, ok := s.CompiledPages[pageName]
	if !ok {
		return nil, u.Logger.NewError("template not found", "page_name", pageName)
	}

	host, err := s.ConfigsRepository.GetConfig(configs.ConfigKeys.ServerHost)
	if err != nil {
		return nil, err
	}

	pageData := GlobalPageData{
		Paths:              tools.Paths,
		Links:              tools.Links,
		Policies:           tools.Policies,
		Auth:               auth,
		Data:               content,
		Config:             s.Config,
		Title:              deriveTitle(pageName),
		HeadAssets:         s.AssetTagBuilder.BuildHeadAssets(),
		ApplicationVersion: tools.ApplicationVersion,
		Host:               host,
	}

	if shouldSetInfoIconRedirectUrl(infoIconRedirectPath, auth.IsAdmin) {
		pageData.InfoIconRedirectUrl = infoIconRedirectPath
	}

	var outputBuffer bytes.Buffer
	if err := compiledPage.Template.ExecuteTemplate(&outputBuffer, framePageName, pageData); err != nil {
		return nil, u.Logger.NewError(err.Error(), "page_name", pageName)
	}

	return outputBuffer.Bytes(), nil
}

func (s *TemplateServiceImpl) walkWebResourcesAndAddToAssetStore(fileExtension string) error {
	f := func(path string, dirEntry fs.DirEntry, walkError error) error {
		return s.addWebResourceToAssetStore(path, dirEntry, walkError, fileExtension)
	}
	return fs.WalkDir(s.FileSystem, ".", f)
}

func (s *TemplateServiceImpl) walkHtmlTemplatesAndCompile(frameText string) (map[string]*CompiledPageTemplate, error) {
	compiledPages := map[string]*CompiledPageTemplate{}
	f := func(path string, dirEntry fs.DirEntry, walkError error) error {
		return s.compileHtmlTemplate(path, dirEntry, walkError, frameText, compiledPages)
	}
	return compiledPages, fs.WalkDir(s.FileSystem, ".", f)
}

func shouldSetInfoIconRedirectUrl(infoIconRedirectPath string, isAdmin bool) bool {
	return infoIconRedirectPath != "" && isAdmin
}

func deriveTitle(pageName string) string {
	s := strings.TrimSpace(pageName)
	s = strings.TrimSuffix(s, ".html")
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.TrimSpace(s)

	if s == "" {
		u.Logger.Error("derive title of empty page", "page_name", pageName)
		return ""
	}

	lowerTitle := strings.ToLower(s)
	lowerTitle = strings.ReplaceAll(lowerTitle, "oidc", "OIDC")

	return strings.ToUpper(lowerTitle[:1]) + lowerTitle[1:]
}

func (s *TemplateServiceImpl) compileHtmlTemplate(path string, dirEntry fs.DirEntry, walkError error, frameText string, compiledPages map[string]*CompiledPageTemplate) error {
	if walkError != nil {
		return u.Logger.NewError(walkError.Error(), "current_path", path)
	}
	if dirEntry.IsDir() || !strings.HasSuffix(path, ".html") {
		return nil
	}
	if filepath.ToSlash(path) == framePathInWebFolder {
		return nil
	}

	pageBytes, err := fs.ReadFile(s.FileSystem, path)
	if err != nil {
		return u.Logger.NewError(err.Error(), "path", path)
	}

	pageName := strings.ToLower(strings.TrimSuffix(filepath.Base(path), ".html"))
	pageText := s.TemplateEngine.PreprocessPageTemplate(string(pageBytes), path)

	compiledTemplate, err := s.TemplateEngine.CompileHtml(frameText, pageText)
	if err != nil {
		return u.Logger.AddContext(err, "page_name", pageName, "path", path)
	}

	compiledPages[pageName] = &CompiledPageTemplate{
		PageName:  pageName,
		Template:  compiledTemplate,
		PagePath:  filepath.ToSlash(path),
		FramePath: framePathInWebFolder,
	}

	u.Logger.Info("compiled template", "name", pageName, "path", path)
	return nil
}

func (s *TemplateServiceImpl) addWebResourceToAssetStore(path string, dirEntry fs.DirEntry, walkError error, fileExtension string) error {
	if walkError != nil {
		return u.Logger.NewError(walkError.Error(), "current_path", path)
	}

	if dirEntry.IsDir() || !strings.HasSuffix(path, "."+fileExtension) {
		return nil
	}

	fileBytes, err := fs.ReadFile(s.FileSystem, path)
	if err != nil {
		return u.Logger.NewError(err.Error(), "path", path)
	}

	fileNameWithoutExtension := strings.TrimSuffix(filepath.Base(path), "."+fileExtension)
	fileFolder := filepath.Dir(path)
	injectedPath := s.AssetStore.GetVersionedInjectedWebResourcePath(fileFolder, fileNameWithoutExtension, fileExtension)

	var renderedBytes []byte
	if strings.Contains(path, "vendor/") {
		renderedBytes = fileBytes
	} else {
		renderedBytes, err = s.TemplateEngine.CompileText(path, string(fileBytes))
		if err != nil {
			return err
		}
	}

	fullPath := filepath.Join(tools.FrontendResourcesPathWithSlash, injectedPath)
	s.AssetStore.Put(fullPath, renderedBytes)

	u.Logger.Info("generated injected web resource", "path", path)
	return nil
}
