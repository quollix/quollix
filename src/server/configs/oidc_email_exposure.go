package configs

import "strconv"

type OidcEmailExposureService interface {
	ReadExposeRealEmailInOidcToken() (bool, error)
	SaveExposeRealEmailInOidcToken(exposeRealEmail bool) error
}

type OidcEmailExposureServiceImpl struct {
	ConfigsRepo ConfigsRepository
}

func (s *OidcEmailExposureServiceImpl) ReadExposeRealEmailInOidcToken() (bool, error) {
	value, err := s.ConfigsRepo.GetConfig(ConfigKeys.ExposeRealEmailInOidcToken)
	if err != nil {
		return false, err
	}
	return value == "true", nil
}

func (s *OidcEmailExposureServiceImpl) SaveExposeRealEmailInOidcToken(exposeRealEmail bool) error {
	return s.ConfigsRepo.SetConfig(ConfigKeys.ExposeRealEmailInOidcToken, strconv.FormatBool(exposeRealEmail))
}
