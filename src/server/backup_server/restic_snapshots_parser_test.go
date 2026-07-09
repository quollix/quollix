package backup_server

import (
	"testing"
	"time"

	"github.com/quollix/common/assert"
)

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
    "description=manual"
  ],
  "program_version":"restic 0.16.4",
  "id":"50d0aa9b730a304c47be60157e6eaab168fd89642375cc931fe98aba09bcf415",
  "short_id":"50d0aa9b"
}]`

func TestResticSnapshotsParser_ParseSingleSnapshot(t *testing.T) {
	parser := &ResticSnapshotsParserImpl{}

	backups, err := parser.Parse(sampleResticJsonOutput)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(backups))

	backup := backups[0]
	assert.Equal(t, "50d0aa9b730a304c47be60157e6eaab168fd89642375cc931fe98aba09bcf415", backup.BackupId)
	assert.Equal(t, "sampleMaintainer", backup.Maintainer)
	assert.Equal(t, "sampleApp", backup.AppName)
	assert.Equal(t, "2.0", backup.VersionName)
	assert.Equal(t, "manual", backup.Description)
	assert.Equal(t, time.Date(2025, 12, 27, 10, 20, 54, 598383072, time.UTC), backup.BackupCreationTimestamp)
}
