package oidc_client

import u "github.com/quollix/common/utils"

const (
	OidcAuthProviderNameAlreadyExistsError             = "OIDC provider name already exists"
	OidcAuthProviderIssuerDomainPathAlreadyExistsError = "OIDC provider issuer domain path already exists"
	OidcAuthProviderRequiredFieldMissingError          = "OIDC provider fields must not be empty"
)

type OidcAuthProviderService interface {
	CreateProvider(provider *OidcAuthProviderDto) error
	UpdateProvider(provider *OidcAuthProviderDto) error
	ListProviders() ([]OidcAuthProviderDto, error)
	DeleteProvider(providerId int) error
	TestDiscovery(issuerDomainPath string) error
}

type OidcAuthProviderServiceImpl struct {
	ProviderRepo   OidcAuthProviderRepository
	ProviderClient OidcProviderClient
}

func (s *OidcAuthProviderServiceImpl) CreateProvider(provider *OidcAuthProviderDto) error {
	if err := normalizeAndValidateProvider(provider); err != nil {
		return err
	}
	if err := s.validateProviderUniqueness(provider); err != nil {
		return err
	}
	_, err := s.ProviderRepo.CreateProvider(provider)
	return err
}

func (s *OidcAuthProviderServiceImpl) UpdateProvider(provider *OidcAuthProviderDto) error {
	if err := normalizeAndValidateProvider(provider); err != nil {
		return err
	}
	if err := s.validateProviderUniqueness(provider); err != nil {
		return err
	}
	return s.ProviderRepo.UpdateProvider(provider)
}

func (s *OidcAuthProviderServiceImpl) ListProviders() ([]OidcAuthProviderDto, error) {
	return s.ProviderRepo.ListProviders()
}

func (s *OidcAuthProviderServiceImpl) DeleteProvider(providerId int) error {
	return s.ProviderRepo.DeleteProvider(providerId)
}

func (s *OidcAuthProviderServiceImpl) TestDiscovery(issuerDomainPath string) error {
	if issuerDomainPath == "" {
		return u.Logger.NewError(OidcAuthProviderRequiredFieldMissingError)
	}
	return s.ProviderClient.TestDiscovery(issuerDomainPath)
}

func (s *OidcAuthProviderServiceImpl) validateProviderUniqueness(provider *OidcAuthProviderDto) error {
	existingProvider, exists, err := s.ProviderRepo.GetProviderByName(provider.Name)
	if err != nil {
		return err
	}
	if exists && existingProvider.Id != provider.Id {
		return u.Logger.NewError(OidcAuthProviderNameAlreadyExistsError)
	}

	existingProvider, exists, err = s.ProviderRepo.GetProviderByIssuerDomainPath(provider.IssuerDomainPath)
	if err != nil {
		return err
	}
	if exists && existingProvider.Id != provider.Id {
		return u.Logger.NewError(OidcAuthProviderIssuerDomainPathAlreadyExistsError)
	}
	return nil
}

func normalizeAndValidateProvider(provider *OidcAuthProviderDto) error {
	if provider.Name == "" || provider.IssuerDomainPath == "" || provider.ClientId == "" || provider.ClientSecret == "" {
		return u.Logger.NewError(OidcAuthProviderRequiredFieldMissingError)
	}
	return nil
}
