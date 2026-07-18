package frontend_pages

import (
	"server/tools"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

type SignInPage struct {
	Frame *FrameType
}

func (l *SignInPage) SignInAsAdmin() *InstalledAppsPage {
	assert.Nil(l.Frame.T, l.Frame.Quollix.SyncedLoginWithBrowser(tools.DefaultAdminName, tools.DefaultAdminPassword))
	l.Frame.Assert.PagePath(api.Paths.FrontendInstalledApps)
	return l.Frame.Pages.InstalledAppsPage
}
