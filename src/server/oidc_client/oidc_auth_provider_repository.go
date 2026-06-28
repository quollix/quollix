package oidc_client

import (
	"server/tools"

	u "github.com/quollix/common/utils"
)

const oidcAuthProviderSelect = `
	SELECT
		oidc_auth_provider_id,
		name,
		issuer_domain_path,
		client_id,
		client_secret
	FROM oidc_auth_providers
`

const OidcAuthProviderNotFoundError = "OIDC provider not found"

type OidcAuthProviderDto struct {
	Id               int    `json:"id"`
	Name             string `json:"name" validate:"loose"`
	IssuerDomainPath string `json:"issuer_domain_path" validate:"domain_path"`
	ClientId         string `json:"client_id" validate:"credential"`
	ClientSecret     string `json:"client_secret" validate:"credential"`
}

type OidcAuthProviderDiscoveryRequest struct {
	IssuerDomainPath string `json:"issuer_domain_path" validate:"domain_path"`
}

type OidcAuthProviderRepository interface {
	CreateProvider(provider *OidcAuthProviderDto) (int, error)
	UpdateProvider(provider *OidcAuthProviderDto) error
	GetProviderById(providerId int) (*OidcAuthProviderDto, error)
	GetProviderByName(name string) (*OidcAuthProviderDto, bool, error)
	GetProviderByIssuerDomainPath(issuerDomainPath string) (*OidcAuthProviderDto, bool, error)
	ListProviders() ([]OidcAuthProviderDto, error)
	DeleteProvider(providerId int) error
}

type OidcAuthProviderRepositoryImpl struct {
	DbConnector tools.DatabaseConnector
}

func (r *OidcAuthProviderRepositoryImpl) CreateProvider(provider *OidcAuthProviderDto) (int, error) {
	var id int
	err := r.DbConnector.GetDB().QueryRow(
		`INSERT INTO oidc_auth_providers (name, issuer_domain_path, client_id, client_secret)
         VALUES ($1, $2, $3, $4)
         RETURNING oidc_auth_provider_id`,
		provider.Name,
		provider.IssuerDomainPath,
		provider.ClientId,
		provider.ClientSecret,
	).Scan(&id)
	if err != nil {
		return 0, u.Logger.NewError(err.Error())
	}
	return id, nil
}

func (r *OidcAuthProviderRepositoryImpl) UpdateProvider(provider *OidcAuthProviderDto) error {
	result, err := r.DbConnector.GetDB().Exec(
		`UPDATE oidc_auth_providers
         SET name = $2, issuer_domain_path = $3, client_id = $4, client_secret = $5
         WHERE oidc_auth_provider_id = $1`,
		provider.Id,
		provider.Name,
		provider.IssuerDomainPath,
		provider.ClientId,
		provider.ClientSecret,
	)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	if rowsAffected == 0 {
		return u.Logger.NewError(OidcAuthProviderNotFoundError)
	}
	return nil
}

func (r *OidcAuthProviderRepositoryImpl) GetProviderById(providerId int) (*OidcAuthProviderDto, error) {
	providers, err := r.queryProviders("WHERE oidc_auth_provider_id = $1", providerId)
	if err != nil {
		return nil, err
	}
	if len(providers) == 0 {
		return nil, u.Logger.NewError(OidcAuthProviderNotFoundError)
	}
	return &providers[0], nil
}

func (r *OidcAuthProviderRepositoryImpl) GetProviderByName(name string) (*OidcAuthProviderDto, bool, error) {
	providers, err := r.queryProviders("WHERE name = $1", name)
	if err != nil {
		return nil, false, err
	}
	if len(providers) == 0 {
		return nil, false, nil
	}
	return &providers[0], true, nil
}

func (r *OidcAuthProviderRepositoryImpl) GetProviderByIssuerDomainPath(issuerDomainPath string) (*OidcAuthProviderDto, bool, error) {
	providers, err := r.queryProviders("WHERE issuer_domain_path = $1", issuerDomainPath)
	if err != nil {
		return nil, false, err
	}
	if len(providers) == 0 {
		return nil, false, nil
	}
	return &providers[0], true, nil
}

func (r *OidcAuthProviderRepositoryImpl) ListProviders() ([]OidcAuthProviderDto, error) {
	return r.queryProviders("ORDER BY name")
}

func (r *OidcAuthProviderRepositoryImpl) queryProviders(where string, args ...any) ([]OidcAuthProviderDto, error) {
	rows, err := r.DbConnector.GetDB().Query(oidcAuthProviderSelect+" "+where, args...)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	defer u.Close(rows)

	var providers []OidcAuthProviderDto
	for rows.Next() {
		var provider OidcAuthProviderDto
		if err := rows.Scan(&provider.Id, &provider.Name, &provider.IssuerDomainPath, &provider.ClientId, &provider.ClientSecret); err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		providers = append(providers, provider)
	}
	if rows.Err() != nil {
		return nil, u.Logger.NewError(rows.Err().Error())
	}
	return providers, nil
}

func (r *OidcAuthProviderRepositoryImpl) DeleteProvider(providerId int) error {
	_, err := r.DbConnector.GetDB().Exec(`DELETE FROM oidc_auth_providers WHERE oidc_auth_provider_id = $1`, providerId)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *OidcAuthProviderRepositoryImpl) Wipe() {
	_, err := r.DbConnector.GetDB().Exec(`DELETE FROM oidc_auth_providers`)
	if err != nil {
		u.Logger.Error(err)
	}
}
