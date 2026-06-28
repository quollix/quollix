package main

import (
	"os"

	"server/di"

	u "github.com/quollix/common/utils"
)

func main() {
	runMain(func() error {
		deps, err := di.WireDependencies()
		if err != nil {
			return err
		}
		u.Logger = deps.Logger
		if err := prepareServer(deps); err != nil {
			return err
		}
		if err := deps.HandlerRegisterer.RegisterApplicationHandlers(); err != nil {
			return err
		}
		return runServer(deps)
	})
}

func runMain(run func() error) {
	if err := run(); err != nil {
		u.Logger.Error(err, "error", "application failed to start")
		os.Exit(1)
	}
}

func prepareServer(deps *di.Dependencies) error {
	if err := deps.CliToolInstallationVerifier.Verify(); err != nil {
		return err
	}
	if err := deps.DirectoryProvider.InitializeDirectories(); err != nil {
		return err
	}
	if err := deps.DatabaseConnector.StartDatabaseAndConnect(); err != nil {
		return err
	}
	if err := deps.AppStoreClientLean.InitializeOnStartup(); err != nil {
		return err
	}
	if err := deps.SystemConfigMigrator.Run(); err != nil {
		return err
	}
	if deps.Config.CreateDatabaseSnapshotOnStartup {
		if err := deps.DatabaseSnapshotRepository.CreateDatabaseSnapshot(); err != nil {
			return err
		}
	}
	if err := deps.AcmeClient.LoadAcmeAccountClient(); err != nil {
		return err
	}
	if err := deps.ResticDockerImageService.UpdateResticDockerImage(); err != nil {
		return err
	}
	return nil
}

func runServer(deps *di.Dependencies) error {
	if err := deps.TemplateService.ReloadTemplateFromFileSystem(); err != nil {
		return err
	}
	deps.AppManager.StartAppsThatShouldBeRunning()
	if err := deps.OidcService.InitializeOidcService(); err != nil {
		return err
	}
	deps.MaintenanceAgent.StartMaintenanceAgentDaemon()
	if err := deps.TimezoneProvider.InitializeIanaTimezones(); err != nil {
		return err
	}

	servers, err := deps.ServerListener.OpenPortsAndListen()
	if err != nil {
		return err
	}

	deps.ShutdownHandler.WaitAndShutdown(servers)
	return nil
}
