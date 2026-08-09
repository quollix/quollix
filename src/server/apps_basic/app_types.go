package apps_basic

import "time"

func NewRepoApp(
	maintainer, appName, versionName, accessPolicy, port, clientId, clientSecret, appSecret string,
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
		AppSecret:                appSecret,
		AutomaticUpdatesEnabled:  automaticUpdatesEnabled,
		AutomaticBackupsEnabled:  automaticBackupsEnabled,
		Secrets:                  map[string]string{},
	}
	return app
}

type RepoApp struct {
	AppId                                      int
	Maintainer, AppName, VersionName           string
	AccessPolicy, ClientId, ClientSecret, Port string
	// AppSecret is deprecated legacy APP_SECRET compatibility. New app definitions should use purpose-specific SECRET_* placeholders.
	AppSecret                                        string
	VersionCreationTimestamp                         time.Time
	VersionContent                                   []byte
	ShouldBeRunning                                  bool
	AutomaticBackupsEnabled, AutomaticUpdatesEnabled bool
	Metadata                                         map[string]string
	Secrets                                          map[string]string
}

type AppRequestData struct {
	Maintainer, AppName, AccessPolicy, Port string
}
