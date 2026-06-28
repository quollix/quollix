package frontend

import (
	"testing"

	"server/apps_basic"
	"server/configs"
	"server/groups"
	"server/tools"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

type featureFrontendBuilderTestObjects struct {
	Builder     *FrontendPageDataBuilderImpl
	AppService  *apps_basic.AppServiceMock
	AppRepo     *apps_basic.AppRepositoryMock
	ConfigsRepo *configs.ConfigsRepositoryMock
	EmailRepo   *configs.EmailRepositoryMock
	GroupRepo   *groups.GroupRepositoryMock
}

func getFeatureTestObjects(t *testing.T) featureFrontendBuilderTestObjects {
	appService := apps_basic.NewAppServiceMock(t)
	appRepo := apps_basic.NewAppRepositoryMock(t)
	configsRepo := configs.NewConfigsRepositoryMock(t)
	emailRepo := configs.NewEmailRepositoryMock(t)
	groupRepo := groups.NewGroupRepositoryMock(t)
	oidcEmailService := &configs.OidcEmailExposureServiceImpl{ConfigsRepo: configsRepo}

	return featureFrontendBuilderTestObjects{
		Builder: &FrontendPageDataBuilderImpl{
			AppService:       appService,
			AppRepo:          appRepo,
			ConfigsRepo:      configsRepo,
			OidcEmailService: oidcEmailService,
			EmailRepository:  emailRepo,
			GroupRepo:        groupRepo,
		},
		AppService:  appService,
		AppRepo:     appRepo,
		ConfigsRepo: configsRepo,
		EmailRepo:   emailRepo,
		GroupRepo:   groupRepo,
	}
}

func TestBuildEmailPage_ReturnsEmailConfigAndInvitationTemplate(t *testing.T) {
	testObjects := getFeatureTestObjects(t)
	expectedConfig := &u.EmailConfig{IsEnabled: true}

	testObjects.EmailRepo.EXPECT().ReadEmailConfig().Return(expectedConfig, nil)
	testObjects.ConfigsRepo.EXPECT().GetConfig(configs.ConfigKeys.InvitationEmailTemplate).Return("invitation template", nil)
	testObjects.ConfigsRepo.EXPECT().GetConfig(configs.ConfigKeys.ExposeRealEmailInOidcToken).Return("true", nil)

	pageContent, err := testObjects.Builder.BuildEmailPage()
	assert.Nil(t, err)
	assert.Equal(t, expectedConfig, pageContent.EmailConfig)
	assert.True(t, pageContent.ExposeRealEmailInOidcToken)
	assert.Equal(t, "invitation template", pageContent.InvitationEmailTemplate)
}

func TestBuildTerminalAppsPage_FiltersAndSortsAndAppendsOfficial(t *testing.T) {
	testObjects := getFeatureTestObjects(t)

	testObjects.AppService.EXPECT().
		ListAppsForAdmin().
		Return([]apps_basic.AppDto{
			{Maintainer: "b-maintainer", AppName: "b-app", IsRunning: true},
			{Maintainer: "a-maintainer", AppName: "z-app", IsRunning: true},
			{Maintainer: "a-maintainer", AppName: "a-app", IsRunning: false},
		}, nil)

	pageContent, err := testObjects.Builder.BuildTerminalAppsPage()
	assert.Nil(t, err)

	assert.Equal(t, 3, len(pageContent.Apps))
	assert.Equal(t, "a-maintainer", pageContent.Apps[0].Maintainer)
	assert.Equal(t, "z-app", pageContent.Apps[0].AppName)
	assert.Equal(t, "b-maintainer", pageContent.Apps[1].Maintainer)
	assert.Equal(t, "b-app", pageContent.Apps[1].AppName)
	assert.Equal(t, u.OfficialMaintainer, pageContent.Apps[2].Maintainer)
	assert.Equal(t, u.OfficialBrandAppName, pageContent.Apps[2].AppName)
}

func TestBuildTerminalServicesPage_WhenOfficialAppSelected_ReturnsOfficialService(t *testing.T) {
	testObjects := getFeatureTestObjects(t)

	pageContent, err := testObjects.Builder.BuildTerminalServicesPage(u.OfficialMaintainer, u.OfficialBrandAppName)
	assert.Nil(t, err)
	assert.Equal(t, u.OfficialMaintainer, pageContent.Maintainer)
	assert.Equal(t, u.OfficialBrandAppName, pageContent.AppName)
	assert.Equal(t, []string{tools.BrandAppService}, pageContent.ServiceNames)
}

func TestBuildTerminalServicesPage_ExtractsAndSortsServiceNamesFromCompose(t *testing.T) {
	testObjects := getFeatureTestObjects(t)
	composeFileBytes := []byte(`
services:
  zeta: {}
  alpha: {}
`)

	testObjects.AppRepo.EXPECT().GetAppByName("b-app").Return(&apps_basic.RepoApp{VersionContent: composeFileBytes}, nil)

	pageContent, err := testObjects.Builder.BuildTerminalServicesPage("b-maintainer", "b-app")
	assert.Nil(t, err)
	assert.Equal(t, "b-maintainer", pageContent.Maintainer)
	assert.Equal(t, "b-app", pageContent.AppName)
	assert.Equal(t, []string{"alpha", "zeta"}, pageContent.ServiceNames)
}

func TestBuildTerminalViewPage_PassesThroughSelection(t *testing.T) {
	testObjects := getFeatureTestObjects(t)

	pageContent, err := testObjects.Builder.BuildTerminalViewPage("b-maintainer", "b-app", "alpha")
	assert.Nil(t, err)
	assert.Equal(t, "b-maintainer", pageContent.Maintainer)
	assert.Equal(t, "b-app", pageContent.AppName)
	assert.Equal(t, "alpha", pageContent.ServiceName)
}

func TestBuildGroupsPage_MapsGroupsToDtos(t *testing.T) {
	testObjects := getFeatureTestObjects(t)

	testObjects.GroupRepo.EXPECT().
		ListAllGroups().
		Return([]groups.Group{
			{Id: 3, Name: "devs"},
			{Id: 12, Name: "admins"},
		}, nil)

	pageContent, err := testObjects.Builder.BuildGroupsPage()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(pageContent.Groups))
	assert.Equal(t, "12", pageContent.Groups[0].Id)
	assert.Equal(t, "admins", pageContent.Groups[0].Name)
	assert.Equal(t, "3", pageContent.Groups[1].Id)
	assert.Equal(t, "devs", pageContent.Groups[1].Name)
}

func TestBuildGroupMembersPage_MapsUsersAndAddsGroupInfo_AndSortsByName(t *testing.T) {
	testObjects := getFeatureTestObjects(t)

	testObjects.GroupRepo.EXPECT().
		ListUsersByGroupMembership(7).
		Return(&groups.UsersByGroupMembership{
			In:    []groups.Member{{Id: 2, Name: "bob"}, {Id: 10, Name: "alice"}},
			NotIn: []groups.Member{{Id: 7, Name: "zara"}, {Id: 5, Name: "carol"}},
		}, nil)
	testObjects.GroupRepo.EXPECT().GetGroupById(7).Return(groups.Group{Id: 7, Name: "devs"}, nil)

	pageContent, err := testObjects.Builder.BuildGroupMembersPage(7)
	assert.Nil(t, err)
	assert.Equal(t, "7", pageContent.GroupId)
	assert.Equal(t, "devs", pageContent.GroupName)
	assert.Equal(t, "10", pageContent.In[0].Id)
	assert.Equal(t, "alice", pageContent.In[0].Name)
	assert.Equal(t, "2", pageContent.In[1].Id)
	assert.Equal(t, "bob", pageContent.In[1].Name)
	assert.Equal(t, "5", pageContent.NotIn[0].Id)
	assert.Equal(t, "carol", pageContent.NotIn[0].Name)
	assert.Equal(t, "7", pageContent.NotIn[1].Id)
	assert.Equal(t, "zara", pageContent.NotIn[1].Name)
}

func TestBuildGroupAppsPage_ReturnsAccessListsAndGroupInfo_AndSortsByName(t *testing.T) {
	testObjects := getFeatureTestObjects(t)

	testObjects.GroupRepo.EXPECT().
		ListAppsAccessByGroup(7).
		Return(&groups.AppsAccessByGroup{
			Granted:    []string{"z-app", "a-app"},
			NotGranted: []string{"d-app", "b-app"},
		}, nil)
	testObjects.GroupRepo.EXPECT().GetGroupById(7).Return(groups.Group{Id: 7, Name: "devs"}, nil)

	pageContent, err := testObjects.Builder.BuildGroupAppsPage(7)
	assert.Nil(t, err)
	assert.Equal(t, "7", pageContent.GroupId)
	assert.Equal(t, "devs", pageContent.GroupName)
	assert.Equal(t, []string{"a-app", "z-app"}, pageContent.AccessGrantedApps)
	assert.Equal(t, []string{"b-app", "d-app"}, pageContent.AccessNotGrantedApps)
}
