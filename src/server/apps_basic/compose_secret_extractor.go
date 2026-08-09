package apps_basic

import (
	"regexp"
	"sort"
	"strings"

	u "github.com/quollix/common/utils"
	"gopkg.in/yaml.v3"
)

const InvalidSecretPlaceholderError = "invalid secret placeholder"

type ComposeSecretExtractor interface {
	Extract(composeContent []byte) ([]string, error)
}

type ComposeSecretExtractorImpl struct{}

var (
	composePlaceholderPattern = regexp.MustCompile(`\$\{([^}]+)\}`)
	secretNamePattern         = regexp.MustCompile(`^SECRET_[A-Z0-9_]+$`)
)

func (e *ComposeSecretExtractorImpl) Extract(composeContent []byte) ([]string, error) {
	var composeMap map[string]any
	if err := yaml.Unmarshal(composeContent, &composeMap); err != nil {
		return nil, err
	}

	secrets := map[string]struct{}{}
	matches := composePlaceholderPattern.FindAllSubmatch(composeContent, -1)
	for _, match := range matches {
		placeholderName := string(match[1])
		if !strings.HasPrefix(placeholderName, "SECRET") {
			continue
		}
		if !secretNamePattern.MatchString(placeholderName) {
			return nil, u.Logger.NewError(InvalidSecretPlaceholderError, "placeholder", placeholderName)
		}
		secrets[placeholderName] = struct{}{}
	}

	secretNames := make([]string, 0, len(secrets))
	for secretName := range secrets {
		secretNames = append(secretNames, secretName)
	}
	sort.Strings(secretNames)
	return secretNames, nil
}
