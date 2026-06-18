package setup

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
	"server/maintenance"
	"server/maintenance/retention"
	"server/tools"
	"server/users"
	"time"

	u "github.com/quollix/common/utils"
	"github.com/quollix/common/validation"
)

var (
	PostgresVersion            = "17.5"
	SamplePostgresCreationDate = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
)

const (
	InitialAdminPasswordEnvVar = "INITIAL_ADMIN_PASSWORD"
)

type DatabaseSeeder interface {
	CreateInitialDatabaseEntriesIfNotPresent() error
}

type DatabaseSeederImpl struct {
	UserRepo                   users.UserRepository
	DbProvider                 tools.DatabaseConnector
	AppService                 apps_basic.AppService
	AppRepo                    apps_basic.AppRepository
	DirectoryProvider          tools.DirectoryProvider
	AuthHelper                 u.AuthHelper
	ClientCredentialsGenerator apps_basic.ClientCredentialsGeneratorImpl
	ConfigsRepo                configs.ConfigsRepository
	CertificatePersister       certificates.CertificatePersister
	EmailRepo                  configs.EmailRepository
	SshRepository              backup_server.SshRepository
	RetentionPolicyRepository  retention.RetentionPolicyRepository
	MaintenanceServiceHelper   maintenance.AgentHelper
	MaintenanceRepo            configs.MaintenanceRepository
	OsWrapper                  u.OsWrapper
	CertificateService         certificates.CertificateService
	DatabaseSnapshotRepository u.DatabaseSnapshotRepository
	DatabaseUtils              u.DatabaseUtils
	GlobalConfig               *tools.GlobalConfig
}

func (d *DatabaseSeederImpl) CreateInitialDatabaseEntriesIfNotPresent() error {
	if err := d.createAdminUserIfNotExist(); err != nil {
		return err
	}
	if err := d.setHostIfNotExist(); err != nil {
		return err
	}
	if err := d.createOidcPrivateKeyIfNotExist(); err != nil {
		return err
	}
	if err := d.createServerCertificateIfNotExist(); err != nil {
		return err
	}
	if err := d.loadEmailSettingsIfNotExist(); err != nil {
		return err
	}
	if err := d.createLetsEncryptAccountKeyIfNotExist(); err != nil {
		return err
	}
	if err := d.createAcmeAccountCreatedFlagIfNotExist(); err != nil {
		return err
	}
	if err := d.createBackupServerConfigIfItNotExist(); err != nil {
		return err
	}
	if err := d.createEmailTemplatesIfNotExist(); err != nil {
		return err
	}
	if err := d.createRetentionPolicyIfNotExist(); err != nil {
		return err
	}
	if err := d.createMaintenanceConfigIfNotExist(); err != nil {
		return err
	}
	if err := d.createVersionConfigIfNotExist(); err != nil {
		return err
	}
	if err := d.createOfficialDatabaseAppEntryIfNotExist(); err != nil {
		return err
	}

	if d.GlobalConfig.CreateDatabaseSnapshotOnStartup {
		return d.DatabaseSnapshotRepository.CreateDatabaseSnapshot()
	}
	return nil
}

func (d *DatabaseSeederImpl) createBackupServerConfigIfItNotExist() error {
	doesBackupServerConfigExist, err := d.SshRepository.IsRemoteBackupConfigPresent()
	if err != nil {
		return err
	}
	if !doesBackupServerConfigExist {
		u.Logger.Info("no remote backup server configuration found in database, creating empty remote backup server configuration")
		if err := d.SshRepository.SaveRemoteBackupRepository(backup_server.GetEmptyRemoteRepoConfigs()); err != nil {
			return err
		}
	}
	return nil
}

func (d *DatabaseSeederImpl) createLetsEncryptAccountKeyIfNotExist() error {
	isSet, err := d.ConfigsRepo.IsConfigSet(configs.ConfigKeys.AcmeAccountPrivateKey)
	if err != nil {
		return err
	}
	if isSet {
		return nil
	}

	privateKeyPath := filepath.Join(d.DirectoryProvider.GetCacheDir(), "acme_account_private_key.pem")
	_, err = os.Stat(privateKeyPath)
	doesAcmePrivateKeyFileExist := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return u.Logger.NewError(err.Error())
	}

	privateKeyBytes, err := d.getPrivateKeyBytes(doesAcmePrivateKeyFileExist, privateKeyPath)
	if err != nil {
		return err
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return d.ConfigsRepo.SetConfig(configs.ConfigKeys.AcmeAccountPrivateKey, string(pemBytes))
}

func (d *DatabaseSeederImpl) getPrivateKeyBytes(doesAcmePrivateKeyFileExist bool, privateKeyPath string) ([]byte, error) {
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

func (d *DatabaseSeederImpl) loadEmailSettingsIfNotExist() error {
	isEmailConfigPresent, err := d.EmailRepo.IsEmailConfigPresent()
	if err != nil {
		return err
	}
	if !isEmailConfigPresent {
		u.Logger.Info("no email configuration found in database, creating empty email configuration")
		if err := d.EmailRepo.SaveEmailConfig(configs.GetEmptyEmailConfig()); err != nil {
			return err
		}
	}
	return nil
}

func (d *DatabaseSeederImpl) createServerCertificateIfNotExist() error {
	if err := d.CertificatePersister.LoadCertificateFromHostSystemToDatabaseIfExist(); err != nil {
		return err
	}

	isCertificatePresent, err := d.ConfigsRepo.IsConfigSet(configs.ConfigKeys.CertificatePemBundle)
	if err != nil {
		return err
	}

	if isCertificatePresent {
		return nil
	}

	u.Logger.Info("no TLS certificate found in database, creating self-signed certificate")
	certBundle, err := d.CertificateService.GenerateUniversalSelfSignedCert()
	if err != nil {
		return err
	}
	return d.CertificateService.ReplaceCertificate(certBundle)
}

func (d *DatabaseSeederImpl) createOidcPrivateKeyIfNotExist() error {
	ok, err := d.ConfigsRepo.IsConfigSet(configs.ConfigKeys.OidcPrivateKey)
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

		if err := d.ConfigsRepo.SetConfig(configs.ConfigKeys.OidcPrivateKey, string(pemBytes)); err != nil {
			return err
		}
	}
	return nil
}

func (d *DatabaseSeederImpl) createOfficialDatabaseAppEntryIfNotExist() error {
	doesAppExist, err := d.AppRepo.DoesAppExist(u.OfficialDatabaseAppName)
	if err != nil {
		return err
	}
	if doesAppExist {
		u.Logger.Info("postgres app already exists, skipping creation", tools.AppField, u.OfficialDatabaseAppName)
		return nil
	}

	u.Logger.Info("creating postgres app")
	composeFilePath, err := url.JoinPath(d.DirectoryProvider.GetOfficialDatabaseDockerDir(), "docker-compose.yml")
	if err != nil {
		return u.Logger.NewError(err.Error(), "compose_file_path", composeFilePath)
	}
	appBytes, err := os.ReadFile(composeFilePath) // #nosec G304 (CWE-22): Potential file inclusion via variable
	if err != nil {
		return u.Logger.NewError(err.Error(), "compose_file_path", composeFilePath)
	}
	// client credentials not needed for postgres, since OIDC is not possible here, but keeping it for consistency
	clientId, clientSecret, err := d.ClientCredentialsGenerator.Generate()
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
	return d.AppService.UpsertAppInDatabase(app)
}

func (d *DatabaseSeederImpl) setHostIfNotExist() error {
	ok, err := d.ConfigsRepo.IsConfigSet(configs.ConfigKeys.ServerHost)
	if err != nil {
		return err
	}
	if !ok {
		err = d.ConfigsRepo.SetConfig(configs.ConfigKeys.ServerHost, "localhost")
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DatabaseSeederImpl) createAdminUserIfNotExist() error {
	doesAdminExist, err := d.UserRepo.DoesAnyAdminUserExist()
	if err != nil {
		return err
	}

	if doesAdminExist {
		u.Logger.Info("At least one Admin user has been found in the database, so the creation of an admin account is being skipped.")
		return nil
	} else {
		u.Logger.Info("No Admin user was found in the database, but the application needs at least one. Admin user is being created.")
		return d.createAdminsUser()
	}
}

func (d *DatabaseSeederImpl) createAdminsUser() error {
	adminPassword := os.Getenv(InitialAdminPasswordEnvVar)

	if adminPassword == "" {
		u.Logger.Info("INITIAL_ADMIN_PASSWORD environment variable is not set, creating default admin user with password 'password'")
		adminPassword = tools.DefaultAdminPassword
	}

	err := validation.Validate("adminPassword", validation.FieldPassword, adminPassword)
	if err != nil {
		return u.Logger.NewError("env variable is not valid", "env_variable", InitialAdminPasswordEnvVar)
	}

	hashedPassword, err := d.AuthHelper.SaltAndHash(adminPassword)
	if err != nil {
		return err
	}
	user := users.NewUser(
		tools.DefaultAdminName,
		tools.DefaultAdminEmail,
		hashedPassword,
		"",
		tools.DefaultTime,
		true,
		"",
		tools.DefaultTime,
		d.OsWrapper.Now(),
	)

	_, err = d.UserRepo.CreateUser(user)
	if err != nil {
		return err
	}
	u.Logger.Info("Initial admin user created")
	return nil
}

func (d *DatabaseSeederImpl) createEmailTemplatesIfNotExist() error {
	doesInvitationEmailTemplateExist, err := d.ConfigsRepo.IsConfigSet(configs.ConfigKeys.InvitationEmailTemplate)
	if err != nil {
		return err
	}
	if !doesInvitationEmailTemplateExist {
		u.Logger.Info("no invitation email template found in database, creating default invitation email template")
		if err := d.ConfigsRepo.SetConfig(configs.ConfigKeys.InvitationEmailTemplate, users.DefaultInvitationEmailTemplate); err != nil {
			return err
		}
	}

	return nil
}

func (d *DatabaseSeederImpl) createRetentionPolicyIfNotExist() error {
	isSet, err := d.RetentionPolicyRepository.IsRetentionPolicySet()
	if err != nil {
		return err
	}
	if !isSet {
		u.Logger.Info("no retention policy found in database, creating default retention policy")
		defaultPolicy := retention.GetDefaultRetentionPolicy()
		if err := d.RetentionPolicyRepository.SetRetentionPolicy(defaultPolicy); err != nil {
			return err
		}
	}
	return nil
}

func (d *DatabaseSeederImpl) createMaintenanceConfigIfNotExist() error {
	isSet, err := d.MaintenanceRepo.IsMaintenanceConfigSet()
	if err != nil {
		return err
	}
	if isSet {
		return nil
	}

	u.Logger.Info("no maintenance config found in database, creating default maintenance config")

	maintenanceWindowStartHour := 2
	defaultIanaTimeZone := "Europe/London"
	nextMaintenanceAtUtc, err := d.MaintenanceServiceHelper.CalculateNextMaintenanceAtUtc(d.OsWrapper.Now(), defaultIanaTimeZone, maintenanceWindowStartHour)
	if err != nil {
		return err
	}

	config := &configs.MaintenanceConfig{
		MaintenanceWindowStartHour: maintenanceWindowStartHour,
		NextMaintenanceAt:          *nextMaintenanceAtUtc,
		IanaTimezone:               defaultIanaTimeZone,
	}
	return d.MaintenanceRepo.SetMaintenanceConfig(config)
}

func (d *DatabaseSeederImpl) createVersionConfigIfNotExist() error {
	isSet, err := d.ConfigsRepo.IsConfigSet(configs.ConfigKeys.ApplicationVersion)
	if err != nil {
		return err
	}
	if !isSet {
		u.Logger.Info("no application version found in database, creating application version entry")
		if err := d.ConfigsRepo.SetConfig(configs.ConfigKeys.ApplicationVersion, tools.ApplicationVersion); err != nil {
			return err
		}
	}
	return nil
}

func (d *DatabaseSeederImpl) createAcmeAccountCreatedFlagIfNotExist() error {
	isSet, err := d.ConfigsRepo.IsConfigSet(configs.ConfigKeys.AcmeAccountRegistered)
	if err != nil {
		return err
	}
	if !isSet {
		u.Logger.Info("no ACME account created flag found in database, creating ACME account created flag and setting it to false")
		if err := d.ConfigsRepo.SetConfig(configs.ConfigKeys.AcmeAccountRegistered, "false"); err != nil {
			return err
		}
	}
	return nil
}
