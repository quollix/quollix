package configs

type ConfigsService interface {
	GetBaseDomain() (string, error)
	SetBaseDomain(baseDomain string) error
	ResetBaseDomainCacheToLocalhost()
}

type ConfigsServiceImpl struct {
	ConfigsRepo ConfigsRepository
	baseDomain  string `wire:"-"`
}

func (s *ConfigsServiceImpl) GetBaseDomain() (string, error) {
	if s.baseDomain != "" {
		return s.baseDomain, nil
	}

	baseDomain, err := s.ConfigsRepo.GetConfig(ConfigKeys.BaseDomain)
	if err != nil {
		return "", err
	}
	s.baseDomain = baseDomain
	return baseDomain, nil
}

func (s *ConfigsServiceImpl) SetBaseDomain(baseDomain string) error {
	if err := s.ConfigsRepo.SetConfig(ConfigKeys.BaseDomain, baseDomain); err != nil {
		return err
	}

	s.baseDomain = baseDomain
	return nil
}

func (s *ConfigsServiceImpl) ResetBaseDomainCacheToLocalhost() {
	s.baseDomain = "localhost"
}
