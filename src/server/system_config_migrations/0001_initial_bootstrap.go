package system_config_migrations

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/url"
	"os"
	"path/filepath"

	"server/apps_basic"
	"server/backup_server"
	"server/certificates"
	"server/configs"
	"server/maintenance/retention"
	"server/tools"
	"server/users"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

func (s *SystemConfigMigrationsProviderImpl) runInitialBootstrapMigration() error {
	steps := []func() error{
		s.createAdminUserIfNotExist,
		s.setHostIfNotExist,
		s.createOidcPrivateKeyIfNotExist,
		s.createOidcEmailExposureConfigIfNotExist,
		s.createServerCertificateIfNotExist,
		s.loadEmailSettingsIfNotExist,
		s.createLetsEncryptAccountKeyIfNotExist,
		s.createAcmeAccountCreatedFlagIfNotExist,
		s.createBackupServerConfigIfItNotExist,
		s.createEmailTemplatesIfNotExist,
		s.createRetentionPolicyIfNotExist,
		s.createMaintenanceConfigIfNotExist,
		s.createVersionConfigIfNotExist,
		s.createOfficialDatabaseAppEntryIfNotExist,
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func (s *SystemConfigMigrationsProviderImpl) createBackupServerConfigIfItNotExist() error {
	doesBackupServerConfigExist, err := s.SshRepository.IsRemoteBackupConfigPresent()
	if err != nil {
		return err
	}
	if !doesBackupServerConfigExist {
		u.Logger.Info("no remote backup server configuration found in database, creating empty remote backup server configuration")
		if err := s.SshRepository.SaveRemoteBackupRepository(backup_server.GetEmptyRemoteRepoConfigs()); err != nil {
			return err
		}
	}
	return nil
}

func (s *SystemConfigMigrationsProviderImpl) createLetsEncryptAccountKeyIfNotExist() error {
	isSet, err := s.ConfigsRepo.IsConfigSet(configs.ConfigKeys.AcmeAccountPrivateKey)
	if err != nil {
		return err
	}
	if isSet {
		return nil
	}

	privateKeyPath := filepath.Join(s.DirectoryProvider.GetCacheDir(), "acme_account_private_key.pem")
	_, err = os.Stat(privateKeyPath)
	doesAcmePrivateKeyFileExist := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return u.Logger.NewError(err.Error())
	}

	privateKeyBytes, err := s.getPrivateKeyBytes(doesAcmePrivateKeyFileExist, privateKeyPath)
	if err != nil {
		return err
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return s.ConfigsRepo.SetConfig(configs.ConfigKeys.AcmeAccountPrivateKey, string(pemBytes))
}

func (s *SystemConfigMigrationsProviderImpl) getPrivateKeyBytes(doesAcmePrivateKeyFileExist bool, privateKeyPath string) ([]byte, error) {
	if doesAcmePrivateKeyFileExist {
		privateKeyBytes, err := os.ReadFile(privateKeyPath) // #nosec G304: privateKeyPath is derived from trusted setup configuration for local bootstrap
		if err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		return privateKeyBytes, nil
	} else {
		privateKey, err := certificates.GenerateRsaKey()
		if err != nil {
			return nil, err
		}
		privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		return privateKeyBytes, nil
	}
}

func (s *SystemConfigMigrationsProviderImpl) loadEmailSettingsIfNotExist() error {
	isEmailConfigPresent, err := s.EmailRepo.IsEmailConfigPresent()
	if err != nil {
		return err
	}
	if !isEmailConfigPresent {
		u.Logger.Info("no email configuration found in database, creating empty email configuration")
		if err := s.EmailRepo.SaveEmailConfig(configs.GetEmptyEmailConfig()); err != nil {
			return err
		}
	}
	return nil
}

func (s *SystemConfigMigrationsProviderImpl) createServerCertificateIfNotExist() error {
	if err := s.CertificatePersister.LoadCertificateFromHostSystemToDatabaseIfExist(); err != nil {
		return err
	}

	isCertificatePresent, err := s.ConfigsRepo.IsConfigSet(configs.ConfigKeys.CertificatePemBundle)
	if err != nil {
		return err
	}

	if isCertificatePresent {
		return nil
	}

	u.Logger.Info("no TLS certificate found in database, creating self-signed certificate")
	certBundle, err := s.CertificateService.GenerateUniversalSelfSignedCert()
	if err != nil {
		return err
	}
	return s.CertificateService.ReplaceCertificate(certBundle)
}

func (s *SystemConfigMigrationsProviderImpl) createOidcPrivateKeyIfNotExist() error {
	ok, err := s.ConfigsRepo.IsConfigSet(configs.ConfigKeys.OidcPrivateKey)
	if err != nil {
		return err
	}
	if !ok {
		key, err := certificates.GenerateRsaKey()
		if err != nil {
			return u.Logger.NewError(err.Error())
		}

		der := x509.MarshalPKCS1PrivateKey(key)
		pemBytes := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: der,
		})

		if err := s.ConfigsRepo.SetConfig(configs.ConfigKeys.OidcPrivateKey, string(pemBytes)); err != nil {
			return err
		}
	}
	return nil
}

func (s *SystemConfigMigrationsProviderImpl) createOidcEmailExposureConfigIfNotExist() error {
	isSet, err := s.ConfigsRepo.IsConfigSet(configs.ConfigKeys.ExposeRealEmailInOidcToken)
	if err != nil {
		return err
	}
	if isSet {
		return nil
	}

	u.Logger.Info("no OIDC real email exposure config found in database, disabling real email exposure")
	return s.OidcEmailService.SaveExposeRealEmailInOidcToken(false)
}

func (s *SystemConfigMigrationsProviderImpl) createOfficialDatabaseAppEntryIfNotExist() error {
	doesAppExist, err := s.AppRepo.DoesAppExist(u.OfficialDatabaseAppName)
	if err != nil {
		return err
	}
	if doesAppExist {
		u.Logger.Info("postgres app already exists, skipping creation", tools.AppField, u.OfficialDatabaseAppName)
		return nil
	}

	u.Logger.Info("creating postgres app")
	composeFilePath, err := url.JoinPath(s.DirectoryProvider.GetOfficialDatabaseDockerDir(), "docker-compose.yml")
	if err != nil {
		return u.Logger.NewError(err.Error(), "compose_file_path", composeFilePath)
	}
	appBytes, err := os.ReadFile(composeFilePath) // #nosec G304 (CWE-22): Potential file inclusion via variable
	if err != nil {
		return u.Logger.NewError(err.Error(), "compose_file_path", composeFilePath)
	}
	// client credentials not needed for postgres, since OIDC is not possible here, but keeping it for consistency
	clientId, clientSecret, err := s.ClientCredentialsGenerator.Generate()
	if err != nil {
		return err
	}
	app := apps_basic.NewRepoApp(
		u.OfficialMaintainer,
		u.OfficialDatabaseAppName,
		PostgresVersion,
		tools.Policies.AdminOnlyAccessPolicy,
		"1",
		clientId,
		clientSecret,
		SamplePostgresCreationDate,
		appBytes,
		true,
		false, // disabling auto updates for postgres, since it is managed by quollix itself
		true,
	)
	return s.AppService.UpsertAppInDatabase(app)
}

func (s *SystemConfigMigrationsProviderImpl) setHostIfNotExist() error {
	ok, err := s.ConfigsRepo.IsConfigSet(legacyServerHostConfigKey)
	if err != nil {
		return err
	}
	if !ok {
		return s.ConfigsRepo.SetConfig(legacyServerHostConfigKey, "localhost")
	}
	return nil
}

func (s *SystemConfigMigrationsProviderImpl) createAdminUserIfNotExist() error {
	doesAdminExist, err := s.UserRepo.DoesAnyAdminUserExist()
	if err != nil {
		return err
	}

	if doesAdminExist {
		u.Logger.Info("At least one Admin user has been found in the database, so the creation of an admin account is being skipped.")
		return nil
	} else {
		u.Logger.Info("No Admin user was found in the database, but the application needs at least one. Admin user is being created.")
		return s.createAdminsUser()
	}
}

func (s *SystemConfigMigrationsProviderImpl) createAdminsUser() error {
	adminName, err := s.getInitialAdminName()
	if err != nil {
		return err
	}

	adminPassword, err := s.getInitialAdminPassword(adminName)
	if err != nil {
		return err
	}

	hashedPassword, err := s.AuthHelper.SaltAndHash(adminPassword)
	if err != nil {
		return err
	}
	user := users.NewUser(
		adminName,
		initialAdminEmail(adminName),
		hashedPassword,
		"",
		tools.DefaultTime,
		true,
		"",
		tools.DefaultTime,
		s.OsWrapper.Now(),
	)

	_, err = s.UserRepo.CreateUser(user)
	if err != nil {
		return err
	}
	u.Logger.Info("Initial admin user created")
	return nil
}

func (s *SystemConfigMigrationsProviderImpl) getInitialAdminName() (string, error) {
	adminName := os.Getenv(InitialAdminNameEnvVar)
	if adminName == "" {
		adminName = DefaultInitialAdminName
	}

	err := validation.Validate("adminName", validation.FieldUsername, adminName)
	if err != nil {
		return "", u.Logger.NewError("env variable is not valid", "env_variable", InitialAdminNameEnvVar)
	}
	return adminName, nil
}

func (s *SystemConfigMigrationsProviderImpl) getInitialAdminPassword(adminName string) (string, error) {
	adminPassword := os.Getenv(InitialAdminPasswordEnvVar)
	isGeneratedPassword := adminPassword == ""
	if isGeneratedPassword {
		var err error
		adminPassword, err = s.AuthHelper.GenerateSecret()
		if err != nil {
			return "", err
		}
		if len(adminPassword) > GeneratedInitialAdminPasswordSize {
			adminPassword = adminPassword[:GeneratedInitialAdminPasswordSize]
		}
	}

	err := validation.Validate("adminPassword", validation.FieldPassword, adminPassword)
	if err != nil {
		return "", u.Logger.NewError("env variable is not valid", "env_variable", InitialAdminPasswordEnvVar)
	}

	if isGeneratedPassword {
		u.Logger.Info("INITIAL_ADMIN_PASSWORD environment variable is not set, generated random initial admin password", "username", adminName, "password", adminPassword)
	} else {
		u.Logger.Info("Creating initial admin user from environment configuration", "username", adminName)
	}
	return adminPassword, nil
}

func initialAdminEmail(adminName string) string {
	return adminName + "@example.invalid"
}

func (s *SystemConfigMigrationsProviderImpl) createEmailTemplatesIfNotExist() error {
	doesInvitationEmailTemplateExist, err := s.ConfigsRepo.IsConfigSet(configs.ConfigKeys.InvitationEmailTemplate)
	if err != nil {
		return err
	}
	if !doesInvitationEmailTemplateExist {
		u.Logger.Info("no invitation email template found in database, creating default invitation email template")
		if err := s.ConfigsRepo.SetConfig(configs.ConfigKeys.InvitationEmailTemplate, users.DefaultInvitationEmailTemplate); err != nil {
			return err
		}
	}

	return nil
}

func (s *SystemConfigMigrationsProviderImpl) createRetentionPolicyIfNotExist() error {
	isSet, err := s.RetentionPolicyRepository.IsRetentionPolicySet()
	if err != nil {
		return err
	}
	if !isSet {
		u.Logger.Info("no retention policy found in database, creating default retention policy")
		defaultPolicy := retention.GetDefaultRetentionPolicy()
		if err := s.RetentionPolicyRepository.SetRetentionPolicy(defaultPolicy); err != nil {
			return err
		}
	}
	return nil
}

func (s *SystemConfigMigrationsProviderImpl) createMaintenanceConfigIfNotExist() error {
	isSet, err := s.MaintenanceRepo.IsMaintenanceConfigSet()
	if err != nil {
		return err
	}
	if isSet {
		return nil
	}

	u.Logger.Info("no maintenance config found in database, creating default maintenance config")

	maintenanceWindowStartHour := 2
	defaultIanaTimeZone := "Europe/London"
	nextMaintenanceAtUtc, err := s.MaintenanceServiceHelper.CalculateNextMaintenanceAtUtc(s.OsWrapper.Now(), defaultIanaTimeZone, maintenanceWindowStartHour)
	if err != nil {
		return err
	}

	config := &configs.MaintenanceConfig{
		MaintenanceWindowStartHour: maintenanceWindowStartHour,
		NextMaintenanceAt:          *nextMaintenanceAtUtc,
		IanaTimezone:               defaultIanaTimeZone,
	}
	return s.MaintenanceRepo.SetMaintenanceConfig(config)
}

func (s *SystemConfigMigrationsProviderImpl) createVersionConfigIfNotExist() error {
	isSet, err := s.ConfigsRepo.IsConfigSet(configs.ConfigKeys.ApplicationVersion)
	if err != nil {
		return err
	}
	if !isSet {
		u.Logger.Info("no application version found in database, creating application version entry")
		if err := s.ConfigsRepo.SetConfig(configs.ConfigKeys.ApplicationVersion, tools.ApplicationVersion); err != nil {
			return err
		}
	}
	return nil
}

func (s *SystemConfigMigrationsProviderImpl) createAcmeAccountCreatedFlagIfNotExist() error {
	isSet, err := s.ConfigsRepo.IsConfigSet(configs.ConfigKeys.AcmeAccountRegistered)
	if err != nil {
		return err
	}
	if !isSet {
		u.Logger.Info("no ACME account created flag found in database, creating ACME account created flag and setting it to false")
		if err := s.ConfigsRepo.SetConfig(configs.ConfigKeys.AcmeAccountRegistered, "false"); err != nil {
			return err
		}
	}
	return nil
}
