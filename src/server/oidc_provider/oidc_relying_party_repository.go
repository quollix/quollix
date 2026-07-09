package oidc_provider

import (
	"server/tools"

	u "github.com/quollix/common/utils"
)

const oidcRelyingPartySelect = `
	SELECT
		oidc_client_id,
		name,
		domain,
		client_id,
		client_secret
	FROM oidc_clients
`

const OidcRelyingPartyNotFoundError = "OIDC relying party not found"

type OidcRelyingPartyDto struct {
	Id           int    `json:"id"`
	Name         string `json:"name" validate:"loose"`
	Domain       string `json:"domain" validate:"domain"`
	ClientId     string `json:"client_id" validate:"loose"`
	ClientSecret string `json:"client_secret" validate:"loose"`
}

type OidcRelyingPartyRequest struct {
	Id     string `json:"id" validate:"number"`
	Name   string `json:"name" validate:"loose"`
	Domain string `json:"domain" validate:"domain"`
}

type OidcRelyingPartyRepository interface {
	CreateClient(client *OidcRelyingPartyDto) (int, error)
	UpdateClient(client *OidcRelyingPartyDto) error
	GetClientById(id int) (*OidcRelyingPartyDto, error)
	GetClientByName(name string) (*OidcRelyingPartyDto, bool, error)
	GetClientByClientId(clientId string) (*OidcRelyingPartyDto, bool, error)
	ListClients() ([]OidcRelyingPartyDto, error)
	DeleteClient(id int) error
}

type OidcRelyingPartyRepositoryImpl struct {
	DbConnector tools.DatabaseConnector
}

func (r *OidcRelyingPartyRepositoryImpl) CreateClient(client *OidcRelyingPartyDto) (int, error) {
	var id int
	err := r.DbConnector.GetDB().QueryRow(
		`INSERT INTO oidc_clients (name, domain, client_id, client_secret)
		 VALUES ($1, $2, $3, $4)
		 RETURNING oidc_client_id`,
		client.Name,
		client.Domain,
		client.ClientId,
		client.ClientSecret,
	).Scan(&id)
	if err != nil {
		return 0, u.Logger.NewError(err.Error())
	}
	return id, nil
}

func (r *OidcRelyingPartyRepositoryImpl) UpdateClient(client *OidcRelyingPartyDto) error {
	result, err := r.DbConnector.GetDB().Exec(
		`UPDATE oidc_clients
		 SET name = $2, domain = $3, client_id = $4, client_secret = $5
		 WHERE oidc_client_id = $1`,
		client.Id,
		client.Name,
		client.Domain,
		client.ClientId,
		client.ClientSecret,
	)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	if rowsAffected == 0 {
		return u.Logger.NewError(OidcRelyingPartyNotFoundError)
	}
	return nil
}

func (r *OidcRelyingPartyRepositoryImpl) GetClientById(id int) (*OidcRelyingPartyDto, error) {
	clients, err := r.queryClients("WHERE oidc_client_id = $1", id)
	if err != nil {
		return nil, err
	}
	if len(clients) == 0 {
		return nil, u.Logger.NewError(OidcRelyingPartyNotFoundError)
	}
	return &clients[0], nil
}

func (r *OidcRelyingPartyRepositoryImpl) GetClientByName(name string) (*OidcRelyingPartyDto, bool, error) {
	clients, err := r.queryClients("WHERE name = $1", name)
	if err != nil {
		return nil, false, err
	}
	if len(clients) == 0 {
		return nil, false, nil
	}
	return &clients[0], true, nil
}

func (r *OidcRelyingPartyRepositoryImpl) GetClientByClientId(clientId string) (*OidcRelyingPartyDto, bool, error) {
	clients, err := r.queryClients("WHERE client_id = $1", clientId)
	if err != nil {
		return nil, false, err
	}
	if len(clients) == 0 {
		return nil, false, nil
	}
	return &clients[0], true, nil
}

func (r *OidcRelyingPartyRepositoryImpl) ListClients() ([]OidcRelyingPartyDto, error) {
	return r.queryClients("ORDER BY name")
}

func (r *OidcRelyingPartyRepositoryImpl) queryClients(where string, args ...any) ([]OidcRelyingPartyDto, error) {
	rows, err := r.DbConnector.GetDB().Query(oidcRelyingPartySelect+" "+where, args...)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	defer u.Close(rows)

	var clients []OidcRelyingPartyDto
	for rows.Next() {
		var client OidcRelyingPartyDto
		if err := rows.Scan(&client.Id, &client.Name, &client.Domain, &client.ClientId, &client.ClientSecret); err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		clients = append(clients, client)
	}
	if rows.Err() != nil {
		return nil, u.Logger.NewError(rows.Err().Error())
	}
	return clients, nil
}

func (r *OidcRelyingPartyRepositoryImpl) DeleteClient(id int) error {
	_, err := r.DbConnector.GetDB().Exec(`DELETE FROM oidc_clients WHERE oidc_client_id = $1`, id)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *OidcRelyingPartyRepositoryImpl) Wipe() {
	_, err := r.DbConnector.GetDB().Exec(`DELETE FROM oidc_clients`)
	if err != nil {
		u.Logger.Error(err)
	}
}
