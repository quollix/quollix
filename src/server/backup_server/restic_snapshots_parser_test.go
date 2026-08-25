package backup_server

import (
	"testing"
	"time"

	"server/tools"

	"github.com/quollix/common/assert"
)

var resticSnapshotsParser = &ResticSnapshotsParserImpl{}

// generated from real restic command
const sampleResticJsonOutput = `
[{
  "time":"2025-12-27T11:20:54.598383072+01:00",
  "tree":"7593ddb95519c2e86119ceeadb60fe7ed7322d5b38121335d2dab67588d57c23",
  "paths":[
    "/tmp/restic-data"
   ],
  "hostname":"tux",
  "username":"tux",
  "uid":1000,
  "gid":1000,
  "tags":[
    "maintainer=sampleMaintainer",
    "app=sampleApp",
    "version=2.0",
    "version_creation_timestamp=2021-01-01T01:00:00Z",
    "description=manual"
  ],
  "program_version":"restic 0.16.4",
  "id":"50d0aa9b730a304c47be60157e6eaab168fd89642375cc931fe98aba09bcf415",
  "short_id":"50d0aa9b"
}]`

func TestResticSnapshotsParser_ParseSingleSnapshot(t *testing.T) {
	backups, err := resticSnapshotsParser.Parse(sampleResticJsonOutput)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(backups))

	backup := backups[0]
	assert.Equal(t, "50d0aa9b730a304c47be60157e6eaab168fd89642375cc931fe98aba09bcf415", backup.BackupId)
	assert.Equal(t, "sampleMaintainer", backup.Maintainer)
	assert.Equal(t, "sampleApp", backup.AppName)
	assert.Equal(t, "2.0", backup.VersionName)
	assert.Equal(t, time.Date(2021, 1, 1, 1, 0, 0, 0, time.UTC), backup.VersionCreationTimestamp)
	assert.Equal(t, "manual", backup.Description)
	assert.Equal(t, time.Date(2025, 12, 27, 10, 20, 54, 598383072, time.UTC), backup.BackupCreationTimestamp)
}

func TestResticSnapshotsParser_ParseLegacySnapshotUsesDefaultVersionCreationTimestamp(t *testing.T) {
	backups, err := resticSnapshotsParser.Parse(`
[{
  "time":"2025-12-27T11:20:54.598383072+01:00",
  "tags":[
    "maintainer=sampleMaintainer",
    "app=sampleApp",
    "version=2.0",
    "description=manual"
  ],
  "id":"50d0aa9b730a304c47be60157e6eaab168fd89642375cc931fe98aba09bcf415"
}]`)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(backups))
	assert.Equal(t, tools.DefaultTime, backups[0].VersionCreationTimestamp)
}
