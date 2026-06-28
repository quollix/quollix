package src

import (
	"server/tests/oidc"
)

func ConfigureOidcTestEnvironment() {
	Tr.Log.TaskDescription("Configuring OIDC two-instance environment")

	oidcProviderAdmin := oidc.NewProviderClient()
	oidcClientAdmin := oidc.NewClientClient()

	if err := oidc.ResetTwoInstanceEnvironment(oidcProviderAdmin, oidcClientAdmin); err != nil {
		Tr.Log.Error("Failed to reset OIDC two-instance environment: %v", err)
		Tr.ExitWithError()
	}
	if err := oidc.ConfigureTwoInstanceEnvironment(oidcProviderAdmin, oidcClientAdmin); err != nil {
		Tr.Log.Error("Failed to configure OIDC two-instance environment: %v", err)
		Tr.ExitWithError()
	}
}
