package frontend

import (
	"fmt"
	"server/app_store"
	"server/apps_basic"
	"server/backup_server"
	"server/configs"
	"server/groups"
	"server/maintenance/retention"
	"server/tools"
	"server/users"
	"sort"
	"strconv"
	"time"

	u "github.com/quollix/common/utils"
	"gopkg.in/yaml.v3"
)

type FrontendPageDataBuilder interface {
	BuildSettingsPage() (*SettingsPageContent, error)
	BuildInstalledAppsPage(usersId int, role tools.UserAccessLevel) (*AppsPageContent, error)
	BuildUsersPage() (*UsersPageContent, error)
	BuildEmailPage() (*EmailPageContent, error)
	BuildTerminalAppsPage() (*TerminalAppsPageContent, error)
	BuildTerminalServicesPage(selectedMaintainer, selectedAppName string) (*TerminalServicesPageContent, error)
	BuildTerminalViewPage(selectedMaintainer, selectedAppName, selectedServiceName string) (*TerminalViewPageContent, error)
	BuildGroupsPage() (*GroupsPageContent, error)
	BuildGroupMembersPage(groupId int) (*GroupMembersPageContent, error)
	BuildGroupAppsPage(groupId int) (*GroupAppsPageContent, error)
	BuildStorePage(maintainerName, appName string, showUnofficial, isSearch bool) (*StorePageContent, error)
	BuildVersionsPage(maintainer, app string) (*VersionsPageContent, error)
	BuildBackedUpAppsPage() (*BackedUpAppsPageContent, error)
	BuildBackupsPage(request tools.MaintainerAndApp) (*BackupsPageContent, error)
	BuildSetPasswordPage(token string) (*SetPasswordPageContent, error)
	BuildOidcClientsPage() (*OidcClientsPageContent, error)
	BuildMaintenancePage() (*MaintenancePage, error)
	BuildUserEditPageData(userId string) (*UserEditPage, error)
	BuildAccountPageData(user *tools.User) *AccountPageData
}

type FrontendPageDataBuilderImpl struct {
	AppService                 apps_basic.AppService
	AppRepo                    apps_basic.AppRepository
	ConfigsRepo                configs.ConfigsRepository
	AppStoreClient             app_store.AppStoreClientLean
	UserRepo                   users.UserRepository
	SshRepositoryConfigService backup_server.SshRepositoryService
	SshRepository              backup_server.SshRepository
	MaintenanceRepo            configs.MaintenanceRepository
	RetentionPolicyRepo        retention.RetentionPolicyRepository
	OsWrapper                  u.OsWrapper
	TimezoneProvider           tools.TimezoneProvider
	EmailRepository            configs.EmailRepository
	GroupRepo                  groups.GroupRepository
	GlobalConfig               *tools.GlobalConfig
}

func (b *FrontendPageDataBuilderImpl) BuildSettingsPage() (*SettingsPageContent, error) {
	sshConfigs, err := b.SshRepositoryConfigService.GetRemoteBackupRepository()
	if err != nil {
		return nil, err
	}

	maintenanceConfig, err := b.MaintenanceRepo.GetMaintenanceConfig()
	if err != nil {
		return nil, err
	}

	retentionPolicy, err := b.RetentionPolicyRepo.GetRetentionPolicy()
	if err != nil {
		return nil, err
	}

	location, err := time.LoadLocation(maintenanceConfig.IanaTimezone)
	if err != nil {
		return nil, err
	}

	nextMaintenanceAtString := maintenanceConfig.NextMaintenanceAt.In(location).Format("Mon, 02 Jan 2006 15:04")

	return &SettingsPageContent{
		BackupServer:             sshConfigs,
		MaintenanceConfig:        maintenanceConfig,
		RetentionPolicy:          retentionPolicy,
		MaintenanceWindowOptions: buildMaintenanceWindowOptions(),
		IanaTimezoneOptions:      b.TimezoneProvider.ListIanaTimezones(),
		NextMaintenanceAt:        nextMaintenanceAtString,
	}, nil
}

func buildMaintenanceWindowOptions() []MaintenanceWindowOption {
	options := make([]MaintenanceWindowOption, 0, 24)
	for startHour := range 24 {
		endHour := (startHour + 1) % 24
		options = append(options, MaintenanceWindowOption{
			Value: startHour,
			Label: fmt.Sprintf("%02d:00-%02d:00", startHour, endHour),
		})
	}
	return options
}

func (b *FrontendPageDataBuilderImpl) BuildInstalledAppsPage(usersId int, role tools.UserAccessLevel) (*AppsPageContent, error) {
	appDtos, err := b.AppService.ListAppsForRole(usersId, role)
	if err != nil {
		return nil, err
	}

	isBackupEnabled, err := b.SshRepository.IsRemoteBackupEnabled()
	if err != nil {
		return nil, err
	}

	now := b.OsWrapper.Now()
	for i := range appDtos {
		appDtos[i].VersionCreationTimestampFormatted = u.FormatRelativeDuration(now, appDtos[i].VersionCreationTimestamp)
		appDtos[i].VersionCreationTimestampTooltip = appDtos[i].VersionCreationTimestamp.UTC().Format(tools.PrettyFrontendTimeLayout)
	}

	sort.Slice(appDtos, func(i int, j int) bool {
		if appDtos[i].Maintainer != appDtos[j].Maintainer {
			return appDtos[i].Maintainer < appDtos[j].Maintainer
		}
		return appDtos[i].AppName < appDtos[j].AppName
	})

	return &AppsPageContent{
		Apps:            appDtos,
		IsBackupEnabled: isBackupEnabled,
	}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildUsersPage() (*UsersPageContent, error) {
	databaseUsers, err := b.UserRepo.ListUsers()
	if err != nil {
		return nil, err
	}

	sort.Slice(databaseUsers, func(i, j int) bool {
		return databaseUsers[i].Username < databaseUsers[j].Username
	})

	host, err := b.ConfigsRepo.GetConfig(configs.ConfigKeys.ServerHost)
	if err != nil {
		return nil, err
	}
	emailConfig, err := b.EmailRepository.ReadEmailConfig()
	if err != nil {
		return nil, err
	}
	var frontendUserDtos []UserFrontendDto
	for _, user := range databaseUsers {
		var setPasswordLinkBase string
		if user.SetPasswordToken != "" {
			setPasswordLinkBase = fmt.Sprintf("https://quollix.%s/set-password?token=%s", host, user.SetPasswordToken)
		}

		dto := UserFrontendDto{
			Id:              user.Id,
			Username:        user.Username,
			Email:           user.Email,
			IsAdmin:         user.IsAdmin,
			SetPasswordLink: setPasswordLinkBase,
			CreatedAt:       user.CreationDate.Format(tools.PrettyFrontendTimeLayout),
		}

		if tools.DefaultTime.Equal(user.SetPasswordTokenExpirationDate) {
			dto.SetPasswordTokenExpirationDate = ""
		} else {
			dto.SetPasswordTokenExpirationDate = user.SetPasswordTokenExpirationDate.Format(tools.PrettyFrontendTimeLayout)
		}
		frontendUserDtos = append(frontendUserDtos, dto)

	}

	return &UsersPageContent{
		Users:          frontendUserDtos,
		IsEmailEnabled: emailConfig.IsEnabled,
	}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildEmailPage() (*EmailPageContent, error) {
	emailConfig, err := b.EmailRepository.ReadEmailConfig()
	if err != nil {
		return nil, err
	}

	invitationTemplate, err := b.ConfigsRepo.GetConfig(configs.ConfigKeys.InvitationEmailTemplate)
	if err != nil {
		return nil, err
	}

	return &EmailPageContent{
		EmailConfig:             emailConfig,
		InvitationEmailTemplate: invitationTemplate,
	}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildTerminalAppsPage() (*TerminalAppsPageContent, error) {
	apps, err := b.listRunningAppsForAdminSorted()
	if err != nil {
		return nil, err
	}
	return &TerminalAppsPageContent{Apps: apps}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildTerminalServicesPage(selectedMaintainer, selectedAppName string) (*TerminalServicesPageContent, error) {
	serviceNames, err := b.listServiceNamesSorted(selectedMaintainer, selectedAppName)
	if err != nil {
		return nil, err
	}

	return &TerminalServicesPageContent{
		Maintainer:   selectedMaintainer,
		AppName:      selectedAppName,
		ServiceNames: serviceNames,
	}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildTerminalViewPage(selectedMaintainer, selectedAppName, selectedServiceName string) (*TerminalViewPageContent, error) {
	return &TerminalViewPageContent{
		Maintainer:  selectedMaintainer,
		AppName:     selectedAppName,
		ServiceName: selectedServiceName,
	}, nil
}

func (b *FrontendPageDataBuilderImpl) listRunningAppsForAdminSorted() ([]apps_basic.AppDto, error) {
	appsForAdmin, err := b.AppService.ListAppsForAdmin()
	if err != nil {
		return nil, err
	}

	runningApps := make([]apps_basic.AppDto, 0, len(appsForAdmin)+1)
	for _, app := range appsForAdmin {
		if app.IsRunning {
			runningApps = append(runningApps, app)
		}
	}

	runningApps = append(runningApps, apps_basic.AppDto{
		Maintainer: u.OfficialMaintainer,
		AppName:    u.OfficialBrandAppName,
	})

	sort.Slice(runningApps, func(i, j int) bool {
		if runningApps[i].Maintainer == runningApps[j].Maintainer {
			return runningApps[i].AppName < runningApps[j].AppName
		}
		return runningApps[i].Maintainer < runningApps[j].Maintainer
	})

	return runningApps, nil
}

func (b *FrontendPageDataBuilderImpl) listServiceNamesSorted(selectedMaintainer, selectedAppName string) ([]string, error) {
	if selectedMaintainer == u.OfficialMaintainer && selectedAppName == u.OfficialBrandAppName {
		return []string{tools.BrandAppService}, nil
	}

	repoApp, err := b.AppRepo.GetAppByName(selectedAppName)
	if err != nil {
		return nil, err
	}

	serviceNames, err := extractServiceNamesFromCompose(repoApp.VersionContent)
	if err != nil {
		return nil, err
	}

	sort.Strings(serviceNames)
	return serviceNames, nil
}

func extractServiceNamesFromCompose(composeFileBytes []byte) ([]string, error) {
	var root map[string]any
	if err := yaml.Unmarshal(composeFileBytes, &root); err != nil {
		return nil, err
	}

	servicesAny, ok := root["services"]
	if !ok {
		return nil, u.Logger.NewError("no services section in compose file")
	}

	servicesMap, ok := servicesAny.(map[string]any)
	if !ok {
		return nil, u.Logger.NewError("invalid services section in compose file")
	}

	serviceNames := make([]string, 0, len(servicesMap))
	for serviceName := range servicesMap {
		serviceNames = append(serviceNames, serviceName)
	}

	sort.Strings(serviceNames)
	return serviceNames, nil
}

func (b *FrontendPageDataBuilderImpl) BuildGroupsPage() (*GroupsPageContent, error) {
	allGroups, err := b.GroupRepo.ListAllGroups()
	if err != nil {
		return nil, err
	}

	sort.Slice(allGroups, func(i int, j int) bool {
		return allGroups[i].Name < allGroups[j].Name
	})

	groupDtos := make([]GroupDTO, 0, len(allGroups))
	for _, group := range allGroups {
		groupDtos = append(groupDtos, GroupDTO{
			Id:   strconv.Itoa(group.Id),
			Name: group.Name,
		})
	}

	return &GroupsPageContent{Groups: groupDtos}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildGroupMembersPage(groupId int) (*GroupMembersPageContent, error) {
	usersByGroup, err := b.GroupRepo.ListUsersByGroupMembership(groupId)
	if err != nil {
		return nil, err
	}

	sort.Slice(usersByGroup.In, func(i int, j int) bool {
		return usersByGroup.In[i].Name < usersByGroup.In[j].Name
	})
	sort.Slice(usersByGroup.NotIn, func(i int, j int) bool {
		return usersByGroup.NotIn[i].Name < usersByGroup.NotIn[j].Name
	})

	inDtos := make([]MemberDto, 0, len(usersByGroup.In))
	for _, member := range usersByGroup.In {
		inDtos = append(inDtos, MemberDto{
			Id:   strconv.Itoa(member.Id),
			Name: member.Name,
		})
	}

	notInDtos := make([]MemberDto, 0, len(usersByGroup.NotIn))
	for _, member := range usersByGroup.NotIn {
		notInDtos = append(notInDtos, MemberDto{
			Id:   strconv.Itoa(member.Id),
			Name: member.Name,
		})
	}

	group, err := b.GroupRepo.GetGroupById(groupId)
	if err != nil {
		return nil, err
	}

	return &GroupMembersPageContent{
		In:        inDtos,
		NotIn:     notInDtos,
		GroupId:   strconv.Itoa(groupId),
		GroupName: group.Name,
	}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildGroupAppsPage(groupId int) (*GroupAppsPageContent, error) {
	appsAccessByGroup, err := b.GroupRepo.ListAppsAccessByGroup(groupId)
	if err != nil {
		return nil, err
	}

	appsAccessByGroup.Granted = filterFrontendApps(appsAccessByGroup.Granted)
	appsAccessByGroup.NotGranted = filterFrontendApps(appsAccessByGroup.NotGranted)
	sort.Strings(appsAccessByGroup.Granted)
	sort.Strings(appsAccessByGroup.NotGranted)

	group, err := b.GroupRepo.GetGroupById(groupId)
	if err != nil {
		return nil, err
	}

	return &GroupAppsPageContent{
		AccessGrantedApps:    appsAccessByGroup.Granted,
		AccessNotGrantedApps: appsAccessByGroup.NotGranted,
		GroupId:              strconv.Itoa(groupId),
		GroupName:            group.Name,
	}, nil
}

func filterFrontendApps(appNames []string) []string {
	filteredAppNames := make([]string, 0, len(appNames))
	for _, appName := range appNames {
		if appName != u.OfficialDatabaseAppName {
			filteredAppNames = append(filteredAppNames, appName)
		}
	}
	return filteredAppNames
}

func (b *FrontendPageDataBuilderImpl) BuildStorePage(maintainerName string, appName string, showUnofficial bool, isSearch bool) (*StorePageContent, error) {
	var searchedMaintainer string
	if showUnofficial {
		searchedMaintainer = ""
	} else {
		searchedMaintainer = maintainerName
	}

	appsToDisplay := []StoreAppDto{}
	if isSearch {
		foundApps, err := b.AppStoreClient.SearchForApps(searchedMaintainer, appName, showUnofficial)
		if err != nil {
			return nil, err
		}

		appsToDisplay = make([]StoreAppDto, 0, len(foundApps))
		for _, app := range foundApps {
			appsToDisplay = append(appsToDisplay, StoreAppDto{
				Maintainer:                     app.Maintainer,
				AppName:                        app.AppName,
				LatestVersionName:              app.LatestVersionName,
				LatestVersionCreationTimestamp: app.LatestVersionCreationTimestamp.UTC().Format(tools.PrettyFrontendTimeLayout),
			})
		}
	}

	return &StorePageContent{
		MaintainerSearchTerm: maintainerName,
		AppSearchTerm:        appName,
		ShowUnofficialApps:   showUnofficial,
		ShowUnofficialToggle: b.GlobalConfig.ShowUnofficialAppsSearch,
		Apps:                 appsToDisplay,
	}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildVersionsPage(maintainer string, app string) (*VersionsPageContent, error) {
	versions, err := b.AppStoreClient.ListVersions(maintainer, app)
	if err != nil {
		return nil, err
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].CreationTimestamp.After(versions[j].CreationTimestamp)
	})

	return &VersionsPageContent{
		Maintainer: maintainer,
		App:        app,
		Versions:   versions,
	}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildBackedUpAppsPage() (*BackedUpAppsPageContent, error) {
	isEnabled, err := b.SshRepository.IsRemoteBackupEnabled()
	if err != nil {
		return nil, err
	}
	if !isEnabled {
		return &BackedUpAppsPageContent{
			IsBackupEnabled: false,
			Apps:            nil,
		}, nil
	}

	return &BackedUpAppsPageContent{
		IsBackupEnabled: true,
		IsLoading:       true,
		Apps:            nil,
	}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildBackupsPage(request tools.MaintainerAndApp) (*BackupsPageContent, error) {
	return &BackupsPageContent{
		Maintainer: request.Maintainer,
		AppName:    request.AppName,
		IsLoading:  true,
	}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildSetPasswordPage(token string) (*SetPasswordPageContent, error) {
	user, err := b.UserRepo.GetUserByToken(token)
	if err != nil {
		return nil, err
	}
	return &SetPasswordPageContent{Username: user.Username}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildOidcClientsPage() (*OidcClientsPageContent, error) {
	appsForAdmin, err := b.AppService.ListAppsForAdmin()
	if err != nil {
		return nil, err
	}

	var filteredApps []apps_basic.AppDto
	for _, app := range appsForAdmin {
		if app.AppName != u.OfficialDatabaseAppName {
			filteredApps = append(filteredApps, app)
		}
	}

	sort.Slice(filteredApps, func(i, j int) bool {
		if filteredApps[i].Maintainer == filteredApps[j].Maintainer {
			return filteredApps[i].AppName < filteredApps[j].AppName
		}
		return filteredApps[i].Maintainer < filteredApps[j].Maintainer
	})

	return &OidcClientsPageContent{Apps: filteredApps}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildMaintenancePage() (*MaintenancePage, error) {
	appsForAdmin, err := b.AppService.ListAppsForAdmin()
	if err != nil {
		return nil, err
	}

	sort.Slice(appsForAdmin, func(i, j int) bool {
		if appsForAdmin[i].Maintainer == appsForAdmin[j].Maintainer {
			return appsForAdmin[i].AppName < appsForAdmin[j].AppName
		}
		return appsForAdmin[i].Maintainer < appsForAdmin[j].Maintainer
	})

	return &MaintenancePage{
		Apps: appsForAdmin,
	}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildUserEditPageData(userId string) (*UserEditPage, error) {
	id, err := strconv.Atoi(userId)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}

	user, err := b.UserRepo.GetUserById(id)
	if err != nil {
		return nil, err
	}

	return &UserEditPage{
		UserId: userId,
		User:   user,
	}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildAccountPageData(user *tools.User) *AccountPageData {
	page := &AccountPageData{
		Username: user.Username,
		Email:    user.Email,
	}
	if user.IsAdmin {
		page.Role = "admin"
	} else {
		page.Role = "user"
	}
	return page
}
