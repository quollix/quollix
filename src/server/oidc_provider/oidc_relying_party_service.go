package oidc_provider

import (
	"strconv"

	"server/apps_basic"

	u "github.com/quollix/common/utils"
)

const (
	OidcRelyingPartyNameAlreadyExistsError     = "OIDC client name already exists"
	OidcRelyingPartyClientIdAlreadyExistsError = "OIDC client ID already exists"
	OidcRelyingPartyRequiredFieldMissingError  = "OIDC client fields must not be empty"
)

type OidcRelyingPartyService interface {
	CreateClient(client *OidcRelyingPartyRequest) error
	UpdateClient(client *OidcRelyingPartyRequest) error
	RegenerateClientCredentials(clientId int) error
}

type OidcRelyingPartyServiceImpl struct {
	ClientRepo                 OidcRelyingPartyRepository
	ClientCredentialsGenerator apps_basic.ClientCredentialsGenerator
}

func (s *OidcRelyingPartyServiceImpl) CreateClient(client *OidcRelyingPartyRequest) error {
	if err := validateRelyingParty(client.Name, client.Domain); err != nil {
		return err
	}
	if err := s.validateNewClientNameIsUnique(client.Name); err != nil {
		return err
	}
	clientId, clientSecret, err := s.ClientCredentialsGenerator.Generate()
	if err != nil {
		return err
	}
	fullClient := &OidcRelyingPartyDto{
		Name:         client.Name,
		Domain:       client.Domain,
		ClientId:     clientId,
		ClientSecret: clientSecret,
	}
	if err := s.validateClientIdUniqueness(fullClient); err != nil {
		return err
	}
	_, err = s.ClientRepo.CreateClient(fullClient)
	return err
}

func (s *OidcRelyingPartyServiceImpl) UpdateClient(client *OidcRelyingPartyRequest) error {
	clientId, err := strconv.Atoi(client.Id)
	if err != nil {
		return err
	}
	if err := validateRelyingParty(client.Name, client.Domain); err != nil {
		return err
	}
	if err := s.validateUpdatedClientNameIsUnique(client.Name, clientId); err != nil {
		return err
	}
	existingClient, err := s.ClientRepo.GetClientById(clientId)
	if err != nil {
		return err
	}
	existingClient.Name = client.Name
	existingClient.Domain = client.Domain
	return s.ClientRepo.UpdateClient(existingClient)
}

func (s *OidcRelyingPartyServiceImpl) RegenerateClientCredentials(clientId int) error {
	client, err := s.ClientRepo.GetClientById(clientId)
	if err != nil {
		return err
	}
	client.ClientId, client.ClientSecret, err = s.ClientCredentialsGenerator.Generate()
	if err != nil {
		return err
	}
	if err := s.validateClientIdUniqueness(client); err != nil {
		return err
	}
	return s.ClientRepo.UpdateClient(client)
}

func (s *OidcRelyingPartyServiceImpl) validateNewClientNameIsUnique(clientName string) error {
	_, exists, err := s.ClientRepo.GetClientByName(clientName)
	if err != nil {
		return err
	}
	if exists {
		return u.Logger.NewError(OidcRelyingPartyNameAlreadyExistsError)
	}
	return nil
}

func (s *OidcRelyingPartyServiceImpl) validateUpdatedClientNameIsUnique(clientName string, clientId int) error {
	existingClient, exists, err := s.ClientRepo.GetClientByName(clientName)
	if err != nil {
		return err
	}
	if exists && existingClient.Id != clientId {
		return u.Logger.NewError(OidcRelyingPartyNameAlreadyExistsError)
	}
	return nil
}

func (s *OidcRelyingPartyServiceImpl) validateClientIdUniqueness(client *OidcRelyingPartyDto) error {
	existingClient, exists, err := s.ClientRepo.GetClientByClientId(client.ClientId)
	if err != nil {
		return err
	}
	if exists && existingClient.Id != client.Id {
		return u.Logger.NewError(OidcRelyingPartyClientIdAlreadyExistsError)
	}
	return nil
}

func validateRelyingParty(name string, domain string) error {
	if name == "" || domain == "" {
		return u.Logger.NewError(OidcRelyingPartyRequiredFieldMissingError)
	}
	return nil
}
