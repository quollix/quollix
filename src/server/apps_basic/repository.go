package apps_basic

import (
	"database/sql"
	"encoding/json"

	"server/tools"

	u "github.com/quollix/common/utils"
)

var (
	appSelect = `
		SELECT
			apps.app_id,
			apps.maintainer,
			apps.app_name,
			apps.version_name,
			apps.version_creation_date,
			apps.version_content,
			apps.should_be_running,
			apps.access_policy,
			apps.client_id,
			apps.client_secret,
			apps.app_secret,
			apps.port,
			apps.automatic_backups_enabled,
			apps.automatic_updates_enabled,
			COALESCE(app_secrets.values, '{}'::jsonb)
		FROM apps
		LEFT JOIN LATERAL (
			SELECT jsonb_object_agg(name, value) AS values
			FROM app_secrets
			WHERE app_secrets.app_id = apps.app_id
		) app_secrets ON true
		WHERE ($1::int IS NULL OR apps.app_id = $1)
		  AND ($2::text IS NULL OR apps.app_name = $2)
		  AND ($3::text IS NULL OR apps.client_id = $3)
	`

	appRequestSelect = `
		SELECT
			apps.maintainer,
			apps.app_name,
			apps.access_policy,
			apps.port
		FROM apps
		WHERE apps.app_name = $1
	`

	appUpdate = `
		UPDATE apps SET
			maintainer = $1,
			app_name = $2,
			version_name = $3,
			version_creation_date = $4,
			version_content = $5,
			should_be_running = $6,
			access_policy = $7,
			client_id = $8,
			client_secret = $9,
			app_secret = $10,
			port = $11,
			automatic_backups_enabled = $12,
			automatic_updates_enabled = $13
		WHERE app_id = $14
	`

	appInsert = `
		INSERT INTO apps (
			maintainer,
			app_name,
			version_name,
			version_creation_date,
			version_content,
			should_be_running,
			access_policy,
			client_id,
			client_secret,
			app_secret,
			port,
			automatic_backups_enabled,
			automatic_updates_enabled
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING app_id
	`
)

type AppRepository interface {
	CreateApp(app *RepoApp) (int, error)
	GetAppById(appId int) (*RepoApp, error)
	GetAppByName(app string) (*RepoApp, error)
	GetAppByClientId(clientId string) (*RepoApp, bool, error)
	GetAppRequestData(appName string) (*AppRequestData, error)
	ListApps() ([]RepoApp, error)
	UpdateApp(app *RepoApp) error
	DeleteApp(appId int) error
	DoesAppExist(appName string) (bool, error)
	DoesAppWithMaintainerExist(maintainer, appName string) (bool, error)
}

type AppRepositoryImpl struct {
	DbProvider tools.DatabaseConnector
}

func (n *AppRepositoryImpl) UpdateApp(app *RepoApp) error {
	tx, err := n.DbProvider.GetDB().Begin()
	if err != nil {
		return u.Logger.NewError(err.Error(), tools.MaintainerField, app.Maintainer, tools.AppField, app.AppName)
	}
	defer rollbackAppRepositoryTx(tx)

	_, err = tx.Exec(appUpdate, appUpdateArgs(app)...)
	if err != nil {
		return u.Logger.NewError(err.Error(), tools.MaintainerField, app.Maintainer, tools.AppField, app.AppName)
	}

	if err = replaceAppSecretsTx(tx, app.AppId, app.Secrets); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return u.Logger.NewError(err.Error(), tools.MaintainerField, app.Maintainer, tools.AppField, app.AppName)
	}
	return nil
}

func (n *AppRepositoryImpl) CreateApp(app *RepoApp) (int, error) {
	tx, err := n.DbProvider.GetDB().Begin()
	if err != nil {
		return -1, u.Logger.NewError(err.Error())
	}
	defer rollbackAppRepositoryTx(tx)

	var appId int
	if err = tx.QueryRow(appInsert, appArgsNoId(app)...).Scan(&appId); err != nil {
		return -1, u.Logger.NewError(err.Error())
	}

	if err = saveAppSecretsTx(tx, appId, app.Secrets); err != nil {
		return -1, err
	}

	if err = tx.Commit(); err != nil {
		return -1, u.Logger.NewError(err.Error())
	}
	return appId, nil
}

func (n *AppRepositoryImpl) GetAppByName(app string) (*RepoApp, error) {
	row := n.DbProvider.GetDB().QueryRow(appSelect, nil, app, nil)
	return scanApp(row)
}

func (n *AppRepositoryImpl) GetAppById(appId int) (*RepoApp, error) {
	row := n.DbProvider.GetDB().QueryRow(appSelect, appId, nil, nil)
	return scanApp(row)
}

func (n *AppRepositoryImpl) GetAppByClientId(clientId string) (*RepoApp, bool, error) {
	row := n.DbProvider.GetDB().QueryRow(appSelect, nil, nil, clientId)
	app, exists, err := scanAppIfExists(row)
	if err != nil || !exists {
		return app, exists, err
	}
	return app, true, nil
}

func (n *AppRepositoryImpl) GetAppRequestData(appName string) (*AppRequestData, error) {
	row := n.DbProvider.GetDB().QueryRow(appRequestSelect, appName)
	return scanAppRequestData(row)
}

func (n *AppRepositoryImpl) DeleteApp(appId int) error {
	if _, err := n.DbProvider.GetDB().Exec("DELETE FROM apps WHERE app_id = $1", appId); err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (n *AppRepositoryImpl) DoesAppExist(appName string) (bool, error) {
	var exists bool
	if err := n.DbProvider.GetDB().QueryRow(
		"SELECT EXISTS(SELECT 1 FROM apps WHERE app_name = $1)",
		appName,
	).Scan(&exists); err != nil {
		return false, u.Logger.NewError(err.Error(), tools.AppField, appName)
	}
	return exists, nil
}

func (n *AppRepositoryImpl) DoesAppWithMaintainerExist(maintainer, appName string) (bool, error) {
	var exists bool
	if err := n.DbProvider.GetDB().QueryRow(
		"SELECT EXISTS(SELECT 1 FROM apps WHERE maintainer = $1 AND app_name = $2)",
		maintainer,
		appName,
	).Scan(&exists); err != nil {
		return false, u.Logger.NewError(err.Error(),
			tools.MaintainerField, maintainer,
			tools.AppField, appName,
		)
	}
	return exists, nil
}

// only used during testing
func (n *AppRepositoryImpl) Wipe() {
	_, err := n.DbProvider.GetDB().Exec("DELETE FROM apps")
	if err != nil {
		u.Logger.Error(err)
	}
}

func appArgsNoId(app *RepoApp) []any {
	return []any{
		app.Maintainer,
		app.AppName,
		app.VersionName,
		app.VersionCreationTimestamp.UTC(),
		app.VersionContent,
		app.ShouldBeRunning,
		app.AccessPolicy,
		app.ClientId,
		app.ClientSecret,
		app.AppSecret,
		app.Port,
		app.AutomaticBackupsEnabled,
		app.AutomaticUpdatesEnabled,
	}
}

func appUpdateArgs(app *RepoApp) []any {
	args := appArgsNoId(app)
	return append(args, app.AppId)
}

type appRowScanner interface {
	Scan(dest ...any) error
}

func (n *AppRepositoryImpl) ListApps() ([]RepoApp, error) {
	var apps []RepoApp

	rows, err := n.DbProvider.GetDB().Query(appSelect, nil, nil, nil)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	defer u.Close(rows)

	for rows.Next() {
		app, err := scanApp(rows)
		if err != nil {
			return nil, err
		}
		apps = append(apps, *app)
	}
	if rows.Err() != nil {
		return nil, u.Logger.NewError(rows.Err().Error())
	}

	return apps, nil
}

func saveAppSecretsTx(tx *sql.Tx, appId int, secrets map[string]string) error {
	for name, value := range secrets {
		if _, err := tx.Exec("INSERT INTO app_secrets (app_id, name, value) VALUES ($1, $2, $3)", appId, name, value); err != nil {
			return u.Logger.NewError(err.Error(), tools.AppIdField, appId, "secret_name", name)
		}
	}
	return nil
}

func replaceAppSecretsTx(tx *sql.Tx, appId int, secrets map[string]string) error {
	if _, err := tx.Exec("DELETE FROM app_secrets WHERE app_id = $1", appId); err != nil {
		return u.Logger.NewError(err.Error(), tools.AppIdField, appId)
	}
	return saveAppSecretsTx(tx, appId, secrets)
}

func rollbackAppRepositoryTx(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		u.Logger.Error(err)
	}
}

func scanApp(scanner appRowScanner) (*RepoApp, error) {
	app, exists, err := scanAppIfExists(scanner)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, u.Logger.NewError(sql.ErrNoRows.Error())
	}
	return app, nil
}

func scanAppIfExists(scanner appRowScanner) (*RepoApp, bool, error) {
	var app RepoApp
	var secrets []byte

	if err := scanner.Scan(
		&app.AppId,
		&app.Maintainer,
		&app.AppName,
		&app.VersionName,
		&app.VersionCreationTimestamp,
		&app.VersionContent,
		&app.ShouldBeRunning,
		&app.AccessPolicy,
		&app.ClientId,
		&app.ClientSecret,
		&app.AppSecret,
		&app.Port,
		&app.AutomaticBackupsEnabled,
		&app.AutomaticUpdatesEnabled,
		&secrets,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, u.Logger.NewError(err.Error())
	}

	app.VersionCreationTimestamp = app.VersionCreationTimestamp.UTC()
	if err := json.Unmarshal(secrets, &app.Secrets); err != nil {
		return nil, false, u.Logger.NewError(err.Error(), tools.AppIdField, app.AppId)
	}
	if app.Secrets == nil {
		app.Secrets = map[string]string{}
	}
	return &app, true, nil
}

func scanAppRequestData(scanner appRowScanner) (*AppRequestData, error) {
	var app AppRequestData

	if err := scanner.Scan(
		&app.Maintainer,
		&app.AppName,
		&app.AccessPolicy,
		&app.Port,
	); err != nil {
		return nil, u.Logger.NewError(err.Error())
	}

	return &app, nil
}
