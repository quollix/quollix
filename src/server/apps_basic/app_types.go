package apps_basic

import "time"

func NewRepoApp(
	maintainer, appName, versionName, accessPolicy, port, clientId, clientSecret string,
	versionCreationTimestamp time.Time,
	versionContent []byte,
	isRunning bool,
	automaticUpdatesEnabled bool,
	automaticBackupsEnabled bool,
) *RepoApp {
	app := &RepoApp{
		AppId:                    -1,
		Maintainer:               maintainer,
		AppName:                  appName,
		VersionName:              versionName,
		Port:                     port,
		VersionCreationTimestamp: versionCreationTimestamp,
		VersionContent:           versionContent,
		ShouldBeRunning:          isRunning,
		AccessPolicy:             accessPolicy,
		ClientId:                 clientId,
		ClientSecret:             clientSecret,
		AutomaticUpdatesEnabled:  automaticUpdatesEnabled,
		AutomaticBackupsEnabled:  automaticBackupsEnabled,
	}
	return app
}

type AppDto struct {
	AppId, Maintainer, AppName, VersionName, AccessPolicy,
	Port, ClientId, ClientSecret, DocsUrl, VersionCreationTimestampFormatted, VersionCreationTimestampTooltip string
	IsRunning, IsOfficialDatabaseApp, AutomaticBackupsEnabled, AutomaticUpdatesEnabled, IsOfficial bool
	VersionCreationTimestamp                                                                       time.Time
	VersionContent                                                                                 []byte
}

type RepoApp struct {
	AppId                                                                        int
	Maintainer, AppName, VersionName, AccessPolicy, ClientId, ClientSecret, Port string
	VersionCreationTimestamp                                                     time.Time
	VersionContent                                                               []byte
	ShouldBeRunning, AutomaticBackupsEnabled, AutomaticUpdatesEnabled            bool
	Metadata                                                                     map[string]string
}

type AppRequestData struct {
	Maintainer, AppName, AccessPolicy, Port string
}
