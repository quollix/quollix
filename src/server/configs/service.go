package configs

type ConfigsService interface {
	GetServerHost() (string, error)
	SetServerHost(host string) error
	ResetServerHostCacheToLocalhost()
}

type ConfigsServiceImpl struct {
	ConfigsRepo ConfigsRepository
	serverHost  string `wire:"-"`
}

func (s *ConfigsServiceImpl) GetServerHost() (string, error) {
	if s.serverHost != "" {
		return s.serverHost, nil
	}

	serverHost, err := s.ConfigsRepo.GetConfig(ConfigKeys.ServerHost)
	if err != nil {
		return "", err
	}
	s.serverHost = serverHost
	return serverHost, nil
}

func (s *ConfigsServiceImpl) SetServerHost(host string) error {
	if err := s.ConfigsRepo.SetConfig(ConfigKeys.ServerHost, host); err != nil {
		return err
	}

	s.serverHost = host
	return nil
}

func (s *ConfigsServiceImpl) ResetServerHostCacheToLocalhost() {
	s.serverHost = "localhost"
}
