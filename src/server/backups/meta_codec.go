package backups

import (
	"os"
	"server/apps_basic"
	"server/tools"
	"time"

	u "github.com/quollix/common/utils"
	"gopkg.in/yaml.v3"
)

func NewMetaData(
	clientId, clientSecret, accessPolicy, port string,
	versionCreationTimestamp time.Time,
	automaticUpdatesEnabled, automaticBackupsEnabled bool,
) *MetaData {
	return &MetaData{
		AccessPolicy:             accessPolicy,
		VersionCreationTimestamp: versionCreationTimestamp,
		ClientId:                 clientId,
		ClientSecret:             clientSecret,
		Port:                     port,
		AutomaticUpdatesEnabled:  automaticUpdatesEnabled,
		AutomaticBackupsEnabled:  automaticBackupsEnabled,
	}
}

type MetaCodecImpl struct {
	ClientCredentialsCreator apps_basic.ClientCredentialsGenerator
}

func (c *MetaCodecImpl) Load(path string) (*MetaData, error) {
	meta := NewMetaData("", "", tools.Policies.AdminOnlyAccessPolicy, "80", tools.DefaultTime, true, true) // default values for unexpected fallback
	data, err := os.ReadFile(path)                                                                         // #nosec G304 G703: path is the trusted backup metadata file selected by application workflow
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	if err = yaml.Unmarshal(data, meta); err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	if meta.ClientId == "" || meta.ClientSecret == "" {
		meta.ClientId, meta.ClientSecret, err = c.ClientCredentialsCreator.Generate()
		if err != nil {
			return nil, err
		}
	}
	return meta, nil
}

func (c *MetaCodecImpl) Save(path string, meta *MetaData) error {
	data, err := yaml.Marshal(meta) // #nosec G117: backup metadata intentionally persists client credentials for snapshot restore
	if err != nil {
		return err
	}
	if err = os.WriteFile(path, data, 0o600); err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}
