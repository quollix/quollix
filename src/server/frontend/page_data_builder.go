package frontend

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"server/app_store"
	"server/apps_basic"
	"server/backup_server"
	"server/configs"
	"server/groups"
	"server/maintenance/retention"
	"server/oidc_client"
	"server/oidc_provider"
	"server/tools"
	"server/users"

	api "github.com/quollix/common/quollix/api"
	u "github.com/quollix/common/utils"
	"gopkg.in/yaml.v3"
)

type FrontendPageDataBuilder interface {
	BuildSignInPage() (*SignInPageContent, error)
	BuildSettingsPage() (*SettingsPageContent, error)
	BuildInstalledAppsPage(usersId int, role tools.UserAccessLevel) (*AppsPageContent, error)
	BuildAppsWithSecretsPage() (*AppsWithSecretsPageContent, error)
	BuildAppSecretPage(appId int) (*AppSecretPageContent, error)
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
	BuildBackupsPage(request api.MaintainerAndApp) (*BackupsPageContent, error)
	BuildSetPasswordPage(token string) (*SetPasswordPageContent, error)
	BuildAppSsoPage() (*AppSsoPageContent, error)
	BuildProvidersPage() (*ProvidersPageContent, error)
	BuildOidcClientsPage() (*OidcClientsPageContent, error)
	BuildMaintenancePage() (*MaintenancePage, error)
	BuildUserEditPageData(userId string) (*UserEditPage, error)
	BuildAccountPageData(user *api.User) *AccountPageData
}

type FrontendPageDataBuilderImpl struct {
	AppService                 apps_basic.AppService
	AppRepo                    apps_basic.AppRepository
	ComposeSecretExtractor     apps_basic.ComposeSecretExtractor
	ConfigsRepo                configs.ConfigsRepository
	ConfigsService             configs.ConfigsService
	OidcEmailService           configs.OidcEmailExposureService
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
	OidcAuthProviderRepo       oidc_client.OidcAuthProviderRepository
	OidcRelyingPartyRepo       oidc_provider.OidcRelyingPartyRepository
	GlobalConfig               *tools.GlobalConfig
}

func (b *FrontendPageDataBuilderImpl) BuildSignInPage() (*SignInPageContent, error) {
	providers, err := b.OidcAuthProviderRepo.ListProviders()
	if err != nil {
		return nil, err
	}

	authProviders := make([]SignInOidcProviderDto, 0, len(providers))
	for _, provider := range providers {
		authProviders = append(authProviders, SignInOidcProviderDto{
			Id:   provider.Id,
			Name: provider.Name,
		})
	}

	return &SignInPageContent{OidcAuthProviders: authProviders}, nil
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
	apps, err := b.buildInstalledAppPageDtos(usersId, role)
	if err != nil {
		return nil, err
	}

	isBackupEnabled, err := b.SshRepository.IsRemoteBackupEnabled()
	if err != nil {
		return nil, err
	}

	sort.Slice(apps, func(i int, j int) bool {
		return isMaintainerAndAppNameBefore(apps[i].Maintainer, apps[i].AppName, apps[j].Maintainer, apps[j].AppName)
	})

	return &AppsPageContent{
		Apps:            apps,
		IsBackupEnabled: isBackupEnabled,
	}, nil
}

func (b *FrontendPageDataBuilderImpl) buildInstalledAppPageDtos(usersId int, role tools.UserAccessLevel) ([]InstalledAppPageDto, error) {
	if role == tools.AdminLevel {
		apps, err := b.AppService.ListAppsForAdmin()
		if err != nil {
			return nil, err
		}
		now := b.OsWrapper.Now()
		return installedAppPageDtosForAdmin(apps, now), nil
	}

	apps, err := b.AppService.ListAppsForNonAdmin(usersId, role)
	if err != nil {
		return nil, err
	}
	return installedAppPageDtosForNonAdmin(apps), nil
}

func installedAppPageDtosForAdmin(apps []api.AdminAppDto, now time.Time) []InstalledAppPageDto {
	appDtos := make([]InstalledAppPageDto, 0, len(apps))
	for _, app := range apps {
		appDtos = append(appDtos, InstalledAppPageDto{
			AppId:                             app.AppId,
			Maintainer:                        app.Maintainer,
			AppName:                           app.AppName,
			VersionName:                       app.VersionName,
			AccessPolicy:                      app.AccessPolicy,
			DocsUrl:                           app.DocsUrl,
			VersionCreationTimestampFormatted: u.FormatRelativeDuration(now, app.VersionCreationTimestamp),
			VersionCreationTimestampTooltip:   app.VersionCreationTimestamp.UTC().Format(tools.PrettyFrontendTimeLayout),
			IsRunning:                         app.IsRunning,
			IsOfficialDatabaseApp:             app.IsOfficialDatabaseApp,
			IsOfficial:                        app.IsOfficial,
			IsPublic:                          app.AccessPolicy == api.Policies.PublicAccessPolicy,
		})
	}
	return appDtos
}

func installedAppPageDtosForNonAdmin(apps []api.NonAdminAppDto) []InstalledAppPageDto {
	appDtos := make([]InstalledAppPageDto, 0, len(apps))
	for _, app := range apps {
		appDtos = append(appDtos, InstalledAppPageDto{
			Maintainer: app.Maintainer,
			AppName:    app.AppName,
			IsRunning:  true,
			IsPublic:   app.IsPublic,
		})
	}
	return appDtos
}

func (b *FrontendPageDataBuilderImpl) BuildAppsWithSecretsPage() (*AppsWithSecretsPageContent, error) {
	appsForAdmin, err := b.AppService.ListAppsForAdmin()
	if err != nil {
		return nil, err
	}

	appsWithSecrets := make([]api.AdminAppDto, 0, len(appsForAdmin))
	for _, app := range appsForAdmin {
		if len(app.Secrets) > 0 {
			appsWithSecrets = append(appsWithSecrets, app)
		}
	}
	sort.Slice(appsWithSecrets, func(i int, j int) bool {
		return isAppBefore(appsWithSecrets[i], appsWithSecrets[j])
	})

	return &AppsWithSecretsPageContent{Apps: appsWithSecrets}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildAppSecretPage(appId int) (*AppSecretPageContent, error) {
	appsForAdmin, err := b.AppService.ListAppsForAdmin()
	if err != nil {
		return nil, err
	}

	appIdString := strconv.Itoa(appId)
	for _, app := range appsForAdmin {
		if app.AppId != appIdString {
			continue
		}

		requiredSecrets, err := b.ComposeSecretExtractor.ExtractSecretSet(app.VersionContent)
		if err != nil {
			return nil, err
		}

		secretRows := make([]AppSecretRow, 0, len(app.Secrets))
		for name, value := range app.Secrets {
			secretRows = append(secretRows, AppSecretRow{Name: name, Value: value, Used: requiredSecrets[name]})
		}
		sort.Slice(secretRows, func(i int, j int) bool {
			return secretRows[i].Name < secretRows[j].Name
		})
		return &AppSecretPageContent{App: app, Secrets: secretRows}, nil
	}

	return nil, u.Logger.NewError("app not found", tools.AppIdField, appId)
}

func (b *FrontendPageDataBuilderImpl) BuildUsersPage() (*UsersPageContent, error) {
	databaseUsers, err := b.UserRepo.ListUsers()
	if err != nil {
		return nil, err
	}

	sort.Slice(databaseUsers, func(i, j int) bool {
		return databaseUsers[i].Username < databaseUsers[j].Username
	})

	host, err := b.ConfigsService.GetBaseDomain()
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
			IsEnabled:       user.IsEnabled,
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

	exposeRealEmail, err := b.OidcEmailService.ReadExposeRealEmailInOidcToken()
	if err != nil {
		return nil, err
	}

	return &EmailPageContent{
		EmailConfig:                emailConfig,
		ExposeRealEmailInOidcToken: exposeRealEmail,
		InvitationEmailTemplate:    invitationTemplate,
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

func (b *FrontendPageDataBuilderImpl) listRunningAppsForAdminSorted() ([]api.AdminAppDto, error) {
	appsForAdmin, err := b.AppService.ListAppsForAdmin()
	if err != nil {
		return nil, err
	}

	runningApps := make([]api.AdminAppDto, 0, len(appsForAdmin)+1)
	for _, app := range appsForAdmin {
		if app.IsRunning {
			runningApps = append(runningApps, app)
		}
	}

	runningApps = append(runningApps, api.AdminAppDto{
		Maintainer: u.OfficialMaintainer,
		AppName:    u.OfficialBrandAppName,
	})

	sort.Slice(runningApps, func(i, j int) bool {
		return isAppBefore(runningApps[i], runningApps[j])
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

		installedAppsByName, err := b.installedAppsByName()
		if err != nil {
			return nil, err
		}

		appsToDisplay = make([]StoreAppDto, 0, len(foundApps))
		for _, app := range foundApps {
			installedApp, isInstalled := installedAppsByName[app.AppName]
			canInstall := canInstallStoreVersion(
				&installedApp,
				isInstalled,
				app.Maintainer,
				app.LatestVersionCreationTimestamp,
			)
			appsToDisplay = append(appsToDisplay, StoreAppDto{
				Maintainer:                     app.Maintainer,
				AppName:                        app.AppName,
				LatestVersionId:                app.LatestVersionId,
				LatestVersionName:              app.LatestVersionName,
				LatestVersionCreationTimestamp: app.LatestVersionCreationTimestamp.UTC().Format(tools.PrettyFrontendTimeLayout),
				CanInstall:                     canInstall,
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

func canInstallStoreVersion(installedApp *apps_basic.RepoApp, isInstalled bool, maintainer string, versionCreationTimestamp time.Time) bool {
	if !isInstalled {
		return true
	}

	if installedApp.Maintainer != maintainer {
		return false
	}

	return versionCreationTimestamp.After(installedApp.VersionCreationTimestamp)
}

func (b *FrontendPageDataBuilderImpl) installedAppsByName() (map[string]apps_basic.RepoApp, error) {
	installedApps, err := b.AppRepo.ListApps()
	if err != nil {
		return nil, err
	}

	installedAppsByName := make(map[string]apps_basic.RepoApp, len(installedApps))
	for _, app := range installedApps {
		installedAppsByName[app.AppName] = app
	}
	return installedAppsByName, nil
}

func (b *FrontendPageDataBuilderImpl) BuildVersionsPage(maintainer string, app string) (*VersionsPageContent, error) {
	versions, err := b.AppStoreClient.ListVersions(maintainer, app)
	if err != nil {
		return nil, err
	}
	installedApp, isInstalled, err := b.installedApp(app)
	if err != nil {
		return nil, err
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].CreationTimestamp.After(versions[j].CreationTimestamp)
	})

	versionDtos := make([]VersionDto, 0, len(versions))
	for _, version := range versions {
		canInstall := canInstallStoreVersion(installedApp, isInstalled, maintainer, version.CreationTimestamp)
		versionDtos = append(versionDtos, VersionDto{
			VersionId:                  version.VersionId,
			Name:                       version.Name,
			CreationTimestampFormatted: version.CreationTimestamp.UTC().Format(tools.PrettyFrontendTimeLayout),
			CanInstall:                 canInstall,
		})
	}

	return &VersionsPageContent{
		Maintainer:  maintainer,
		App:         app,
		IsInstalled: isInstalled,
		Versions:    versionDtos,
	}, nil
}

func (b *FrontendPageDataBuilderImpl) installedApp(appName string) (*apps_basic.RepoApp, bool, error) {
	doesAppExist, err := b.AppRepo.DoesAppExist(appName)
	if err != nil {
		return nil, false, err
	}
	if !doesAppExist {
		return nil, false, nil
	}
	app, err := b.AppRepo.GetAppByName(appName)
	if err != nil {
		return nil, false, err
	}
	return app, true, nil
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

func (b *FrontendPageDataBuilderImpl) BuildBackupsPage(request api.MaintainerAndApp) (*BackupsPageContent, error) {
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

func (b *FrontendPageDataBuilderImpl) BuildAppSsoPage() (*AppSsoPageContent, error) {
	appsForAdmin, err := b.AppService.ListAppsForAdmin()
	if err != nil {
		return nil, err
	}

	var filteredApps []api.AdminAppDto
	for _, app := range appsForAdmin {
		if app.AppName != u.OfficialDatabaseAppName {
			filteredApps = append(filteredApps, app)
		}
	}

	sort.Slice(filteredApps, func(i, j int) bool {
		return isAppBefore(filteredApps[i], filteredApps[j])
	})

	return &AppSsoPageContent{Apps: filteredApps}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildProvidersPage() (*ProvidersPageContent, error) {
	authProviders, err := b.OidcAuthProviderRepo.ListProviders()
	if err != nil {
		return nil, err
	}
	return &ProvidersPageContent{AuthProviders: authProviders}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildOidcClientsPage() (*OidcClientsPageContent, error) {
	clients, err := b.OidcRelyingPartyRepo.ListClients()
	if err != nil {
		return nil, err
	}
	return &OidcClientsPageContent{Clients: clients}, nil
}

func (b *FrontendPageDataBuilderImpl) BuildMaintenancePage() (*MaintenancePage, error) {
	appsForAdmin, err := b.AppService.ListAppsForAdmin()
	if err != nil {
		return nil, err
	}

	sort.Slice(appsForAdmin, func(i, j int) bool {
		return isAppBefore(appsForAdmin[i], appsForAdmin[j])
	})

	return &MaintenancePage{
		Apps: appsForAdmin,
	}, nil
}

func isAppBefore(left, right api.AdminAppDto) bool {
	return isMaintainerAndAppNameBefore(left.Maintainer, left.AppName, right.Maintainer, right.AppName)
}

func isMaintainerAndAppNameBefore(leftMaintainer, leftAppName, rightMaintainer, rightAppName string) bool {
	if leftMaintainer == rightMaintainer {
		return leftAppName < rightAppName
	}
	return leftMaintainer < rightMaintainer
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

func (b *FrontendPageDataBuilderImpl) BuildAccountPageData(user *api.User) *AccountPageData {
	page := &AccountPageData{
		Username:      user.Username,
		Email:         user.Email,
		IsPasswordSet: user.IsPasswordSet(),
	}
	if user.IsAdmin {
		page.Role = "admin"
	} else {
		page.Role = "user"
	}
	return page
}
