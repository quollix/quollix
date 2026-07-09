package ingress

import (
	"testing"

	"server/configs"

	"github.com/quollix/common/assert"
)

func TestAuthServiceImpl_IsRequestAddressedToAnApp_HappyPath(t *testing.T) {
	configsService := configs.NewConfigsServiceMock(t)
	appRequestPolicy := NewAppRequestPolicyMock(t)
	certificateTools := NewCertificateToolsMock(t)

	authService := &AuthServiceImpl{
		ConfigsService:   configsService,
		AppRequestPolicy: appRequestPolicy,
		CertificateTools: certificateTools,
	}

	requestHost := "app.example.com"
	originHeaderValue := "https://quollix.example.com"
	databaseHost := "example.com"
	requestOriginHost := "quollix.example.com"
	expectedIsApp := true

	configsService.EXPECT().GetBaseDomain().Return(databaseHost, nil)
	certificateTools.EXPECT().GetHostFromOriginHeaderValue(originHeaderValue).Return(requestOriginHost, nil)
	appRequestPolicy.EXPECT().ValidateRequestOrigin(requestHost, requestOriginHost, databaseHost).Return(nil)
	appRequestPolicy.EXPECT().IsRequestAddressedToAnApp(requestHost, databaseHost).Return(expectedIsApp)

	isApp, err := authService.IsRequestAddressedToAnApp(requestHost, originHeaderValue)

	assert.Nil(t, err)
	assert.Equal(t, expectedIsApp, isApp)
}
