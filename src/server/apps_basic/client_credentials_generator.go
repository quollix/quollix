package apps_basic

import u "github.com/quollix/common/utils"

type ClientCredentialsGenerator interface {
	Generate() (string, string, error)
}

type ClientCredentialsGeneratorImpl struct {
	AuthHelper u.AuthHelper
}

func (c ClientCredentialsGeneratorImpl) Generate() (string, string, error) {
	longId, err := c.AuthHelper.GenerateSecret()
	if err != nil {
		return "", "", err
	}
	clientId := longId[:16]
	clientSecret, err := c.AuthHelper.GenerateSecret()
	if err != nil {
		return "", "", err
	}
	return clientId, clientSecret, nil
}
