package src

import (
	"github.com/quollix/common/quollix/test_environment"
)

func ConfigureOidcTestEnvironment() {
	Tr.Log.TaskDescription("Configuring OIDC two-instance environment")

	clients := test_environment.NewOidcTwoInstanceClients()

	if err := clients.Reset(); err != nil {
		Tr.Log.Error("Failed to reset OIDC two-instance environment: %v", err)
		Tr.ExitWithError()
	}
	if err := clients.Configure(); err != nil {
		Tr.Log.Error("Failed to configure OIDC two-instance environment: %v", err)
		Tr.ExitWithError()
	}
}
