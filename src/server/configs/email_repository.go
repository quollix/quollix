package configs

import (
	"fmt"
	"server/tools"

	u "github.com/quollix/common/utils"
)

type EmailRepository interface {
	SaveEmailConfig(config *u.EmailConfig) error
	ReadEmailConfig() (*u.EmailConfig, error)
	IsEmailConfigPresent() (bool, error)
}

type EmailRepoImpl struct {
	DatabaseConnector tools.DatabaseConnector
}

func (e *EmailRepoImpl) SaveEmailConfig(config *u.EmailConfig) error {
	_, err := e.DatabaseConnector.GetDB().Exec(`
INSERT INTO configs (key, value)
VALUES
  ($1,  $2),
  ($3,  $4),
  ($5,  $6),
  ($7,  $8),
  ($9, $10),
  ($11, $12)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value;
`,
		ConfigKeys.SmtpHost, config.SMTPHost,
		ConfigKeys.SmtpPort, config.SMTPPort,
		ConfigKeys.SmtpFromEmailAddress, config.FromEmailAddress,
		ConfigKeys.SmtpUsername, config.EmailAccountUsername,
		ConfigKeys.SmtpPassword, config.EmailAccountPassword,
		ConfigKeys.SmtpEnabled, fmt.Sprintf("%t", config.IsEnabled),
	)
	return err
}

func (e *EmailRepoImpl) ReadEmailConfig() (*u.EmailConfig, error) {
	var cfg u.EmailConfig
	var isEnabledString string

	err := e.DatabaseConnector.GetDB().QueryRow(`
SELECT
	(SELECT value FROM configs WHERE key = $1),
	(SELECT value FROM configs WHERE key = $2),
	(SELECT value FROM configs WHERE key = $3),
	(SELECT value FROM configs WHERE key = $4),
	(SELECT value FROM configs WHERE key = $5),
	(SELECT value FROM configs WHERE key = $6);
`,
		ConfigKeys.SmtpHost,
		ConfigKeys.SmtpPort,
		ConfigKeys.SmtpFromEmailAddress,
		ConfigKeys.SmtpUsername,
		ConfigKeys.SmtpPassword,
		ConfigKeys.SmtpEnabled,
	).Scan(
		&cfg.SMTPHost,
		&cfg.SMTPPort,
		&cfg.FromEmailAddress,
		&cfg.EmailAccountUsername,
		&cfg.EmailAccountPassword,
		&isEnabledString,
	)

	if err != nil {
		return nil, err
	}

	cfg.IsEnabled = isEnabledString == "true"
	return &cfg, nil
}

func (e *EmailRepoImpl) IsEmailConfigPresent() (bool, error) {
	var count int
	err := e.DatabaseConnector.GetDB().QueryRow(`
SELECT COUNT(*) FROM configs
WHERE key IN ($1, $2, $3, $4, $5, $6);
`,
		ConfigKeys.SmtpHost,
		ConfigKeys.SmtpPort,
		ConfigKeys.SmtpFromEmailAddress,
		ConfigKeys.SmtpUsername,
		ConfigKeys.SmtpPassword,
		ConfigKeys.SmtpEnabled,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 6, nil
}

func GetEmptyEmailConfig() *u.EmailConfig {
	return &u.EmailConfig{
		SMTPHost:             "",
		SMTPPort:             "",
		FromEmailAddress:     "",
		EmailAccountUsername: "",
		EmailAccountPassword: "",
		IsEnabled:            false,
	}
}

func GetSampleEmailConfig() *u.EmailConfig {
	return &u.EmailConfig{
		SMTPHost:             tools.SampleSMTPHost,
		SMTPPort:             tools.SampleSMTPPort,
		FromEmailAddress:     tools.SampleFromEmailAddress,
		EmailAccountUsername: tools.SampleEmailUsername,
		EmailAccountPassword: tools.SampleEmailPassword,
		IsEnabled:            true,
	}
}
