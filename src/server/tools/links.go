package tools

const (
	websiteBaseUrl          = "https://quollix.org"
	UsageDocsBaseUrl        = websiteBaseUrl + "/docs/usage"
	InstalledAppDocsBaseUrl = UsageDocsBaseUrl + "/installed-apps"
)

var OfficialAppNames = []string{
	"forgejo",
	"hedgedoc",
	"jitsi",
	"nextcloud",
	"vaultwarden",
	"wikijs",
	"wordpress",
	"zulip",
}

type UsageDocsLinksType struct {
	Settings      string
	InstalledApps string
	Users         string
	AppStore      string
	StoreVersions string
	Backups       string
	OidcClients   string
	Maintenance   string
	Email         string
	Terminal      string
	Groups        string
}

type LinksType struct {
	Website            string
	GitHubRepositories string
	FeedbackDocs       string
	UsageDocs          UsageDocsLinksType
}

func InstalledAppDocsUrl(appName string) string {
	return InstalledAppDocsBaseUrl + "/" + appName
}

var Links = LinksType{
	Website:            websiteBaseUrl,
	GitHubRepositories: "https://github.com/orgs/quollix/repositories",
	FeedbackDocs:       websiteBaseUrl + "/docs/feedback/",
	UsageDocs: UsageDocsLinksType{
		Settings:      UsageDocsBaseUrl + "/settings",
		InstalledApps: UsageDocsBaseUrl + "/installed-apps",
		Users:         UsageDocsBaseUrl + "/users",
		AppStore:      UsageDocsBaseUrl + "/app-store",
		StoreVersions: UsageDocsBaseUrl + "/app-store",
		Backups:       UsageDocsBaseUrl + "/backups",
		OidcClients:   UsageDocsBaseUrl + "/oidc-clients",
		Maintenance:   UsageDocsBaseUrl + "/maintenance",
		Email:         UsageDocsBaseUrl + "/email",
		Terminal:      UsageDocsBaseUrl + "/terminal",
		Groups:        UsageDocsBaseUrl + "/groups",
	},
}
