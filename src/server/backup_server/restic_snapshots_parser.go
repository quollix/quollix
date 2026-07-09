package backup_server

import (
	"encoding/json"
	"server/tools"
	"strings"
	"time"

	u "github.com/quollix/common/utils"
)

const (
	MaintainerResticTag         = "maintainer"
	AppResticTag                = "app"
	VersionResticTag            = "version"
	DescriptionResticTag        = "description"
	ApplicationVersionResticTag = "application_version"
)

type ResticSnapshotsParser interface {
	Parse(jsonStr string) ([]tools.BackupInfo, error)
}

type ResticSnapshotsParserImpl struct{}

func (p *ResticSnapshotsParserImpl) Parse(jsonStr string) ([]tools.BackupInfo, error) {
	var snapshots []Snapshot
	if err := json.Unmarshal([]byte(jsonStr), &snapshots); err != nil {
		return nil, err
	}

	var backups []tools.BackupInfo
	for _, snap := range snapshots {
		parsedTime, err := time.Parse(time.RFC3339, snap.Time)
		if err != nil {
			return nil, u.Logger.NewError(err.Error())
		}

		tagMap := make(map[string]string)
		for _, tag := range snap.Tags {
			parts := strings.SplitN(tag, "=", 2)
			if len(parts) == 2 {
				tagMap[parts[0]] = parts[1]
			}
		}

		backups = append(backups, tools.BackupInfo{
			BackupId:                snap.Id,
			Maintainer:              tagMap[MaintainerResticTag],
			AppName:                 tagMap[AppResticTag],
			VersionName:             tagMap[VersionResticTag],
			Description:             tagMap[DescriptionResticTag],
			ApplicationVersion:      tagMap[ApplicationVersionResticTag],
			BackupCreationTimestamp: parsedTime.UTC(),
		})
	}

	return backups, nil
}

type Snapshot struct {
	Time string   `json:"time"`
	Tags []string `json:"tags"`
	Id   string   `json:"id"`
}
