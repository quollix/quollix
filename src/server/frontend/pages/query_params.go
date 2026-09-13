package pages

type QueryParamsType struct {
	Store StoreQueryParams
}

type StoreQueryParams struct {
	MaintainerName string
	AppName        string
	ShowUnofficial string
	IsSearch       string
}

var QueryParams = QueryParamsType{
	Store: StoreQueryParams{
		MaintainerName: "maintainer_name",
		AppName:        "app_name",
		ShowUnofficial: "show_unofficial",
		IsSearch:       "is_search",
	},
}
