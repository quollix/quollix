package app_migrations

import (
	"strconv"
	"strings"

	u "github.com/quollix/common/utils"
	"gopkg.in/yaml.v3"
)

const (
	MultiplePostgresServicesError    = "multiple postgres services found"
	MissingPostgresDataVolumeError   = "missing postgres data volume"
	MultiplePostgresDataVolumesError = "multiple postgres data volumes found"
	InvalidPostgresDataVolumeError   = "invalid postgres data volume"
	InvalidPostgresEnvironmentError  = "invalid postgres environment"
	MultipleRabbitMQServicesError    = "multiple rabbitmq services found"
)

const defaultPostgresUser = "postgres"

type PostgresService struct {
	ServiceName   string
	ContainerName string
	ImageMajor    int
	User          string
	DataVolume    string
}

type RabbitMQService struct {
	ServiceName   string
	ContainerName string
	ImageVersion  string
}

type ComposeDatabaseExtractor interface {
	ExtractPostgresService(composeContent []byte) (*PostgresService, bool, error)
	ExtractRabbitMQService(composeContent []byte) (*RabbitMQService, bool, error)
}

type ComposeDatabaseExtractorImpl struct{}

type composeFile struct {
	Services map[string]composeService `yaml:"services"`
}

type composeService struct {
	Image         string   `yaml:"image"`
	ContainerName string   `yaml:"container_name"`
	Environment   any      `yaml:"environment"`
	Volumes       []string `yaml:"volumes"`
}

type composeServiceWithName struct {
	ServiceName string
	composeService
}

func (e *ComposeDatabaseExtractorImpl) ExtractPostgresService(composeContent []byte) (*PostgresService, bool, error) {
	service, exists, err := extractSingleService(composeContent, isPostgresImage, MultiplePostgresServicesError)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}

	imageMajor, _ := parsePostgresImageMajor(service.Image)
	dataVolume, err := extractPostgresDataVolume(service.Volumes)
	if err != nil {
		return nil, false, err
	}

	env, err := extractEnvironment(service.Environment)
	if err != nil {
		return nil, false, err
	}

	user := env["POSTGRES_USER"]
	if user == "" {
		u.Logger.Warn("postgres service has no explicit POSTGRES_USER, defaulting to postgres", "service_name", service.ServiceName)
		user = defaultPostgresUser
	}

	return &PostgresService{
		ServiceName:   service.ServiceName,
		ContainerName: service.ContainerName,
		ImageMajor:    imageMajor,
		User:          user,
		DataVolume:    dataVolume,
	}, true, nil
}

func (e *ComposeDatabaseExtractorImpl) ExtractRabbitMQService(composeContent []byte) (*RabbitMQService, bool, error) {
	service, exists, err := extractSingleService(composeContent, isRabbitMQImage, MultipleRabbitMQServicesError)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}

	return &RabbitMQService{
		ServiceName:   service.ServiceName,
		ContainerName: service.ContainerName,
		ImageVersion:  extractImageTag(service.Image),
	}, true, nil
}

func extractSingleService(composeContent []byte, matchesImage func(image string) bool, multipleServicesError string) (*composeServiceWithName, bool, error) {
	var compose composeFile
	if err := yaml.Unmarshal(composeContent, &compose); err != nil {
		return nil, false, err
	}

	var matchingService *composeServiceWithName
	for serviceName, service := range compose.Services {
		if !matchesImage(service.Image) {
			continue
		}
		if matchingService != nil {
			return nil, false, u.Logger.NewError(multipleServicesError)
		}

		matchingService = &composeServiceWithName{
			ServiceName:    serviceName,
			composeService: service,
		}
	}

	if matchingService == nil {
		return nil, false, nil
	}
	return matchingService, true, nil
}

func parsePostgresImageMajor(image string) (int, bool) {
	name, tag, ok := strings.Cut(image, ":")
	if !ok {
		return 0, false
	}
	if !isPostgresImageName(name) {
		return 0, false
	}

	majorString := tag
	if index := strings.IndexFunc(tag, func(r rune) bool { return r < '0' || r > '9' }); index >= 0 {
		majorString = tag[:index]
	}
	if majorString == "" {
		return 0, false
	}

	major, err := strconv.Atoi(majorString)
	if err != nil {
		return 0, false
	}
	return major, true
}

func isPostgresImage(image string) bool {
	name, _, ok := strings.Cut(image, ":")
	return ok && isPostgresImageName(name)
}

func isPostgresImageName(name string) bool {
	return name == "postgres" || name == "zulip/zulip-postgresql"
}

func isRabbitMQImage(image string) bool {
	name, _, _ := strings.Cut(image, ":")
	return name == "rabbitmq"
}

func extractImageTag(image string) string {
	_, tag, _ := strings.Cut(image, ":")
	return tag
}

func extractPostgresDataVolume(volumes []string) (string, error) {
	if len(volumes) == 0 {
		return "", u.Logger.NewError(MissingPostgresDataVolumeError)
	}
	if len(volumes) > 1 {
		return "", u.Logger.NewError(MultiplePostgresDataVolumesError)
	}

	source, target, ok := strings.Cut(volumes[0], ":")
	if !ok || source == "" || target == "" {
		return "", u.Logger.NewError(InvalidPostgresDataVolumeError)
	}
	return source, nil
}

func extractEnvironment(environment any) (map[string]string, error) {
	env := map[string]string{}
	if environment == nil {
		return env, nil
	}

	switch typedEnvironment := environment.(type) {
	case map[string]any:
		for key, value := range typedEnvironment {
			valueString, isString := value.(string)
			if !isString {
				continue
			}
			env[key] = valueString
		}
	case []any:
		for _, entry := range typedEnvironment {
			entryString, isString := entry.(string)
			if !isString {
				return nil, u.Logger.NewError(InvalidPostgresEnvironmentError)
			}
			key, value, ok := strings.Cut(entryString, "=")
			if !ok {
				continue
			}
			env[key] = value
		}
	default:
		return nil, u.Logger.NewError(InvalidPostgresEnvironmentError)
	}

	return env, nil
}
