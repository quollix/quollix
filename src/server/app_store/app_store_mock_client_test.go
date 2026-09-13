package app_store

import (
	"testing"
	"time"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/store"
)

func TestAppStoreClientMockDownloadNextVersionForUpdate_ReturnsLatestVersionAfterCurrentTimestamp(t *testing.T) {
	currentTimestamp := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	olderTimestamp := currentTimestamp.Add(time.Hour)
	latestTimestamp := currentTimestamp.Add(2 * time.Hour)
	client := &AppStoreClientMock{
		Versions: []store.Version{
			{
				VersionId:                1,
				Maintainer:               "maintainer",
				AppName:                  "app",
				VersionName:              "1.1",
				VersionCreationTimestamp: olderTimestamp,
			},
			{
				VersionId:                2,
				Maintainer:               "maintainer",
				AppName:                  "app",
				VersionName:              "2.0",
				VersionCreationTimestamp: latestTimestamp,
			},
		},
	}

	response, err := client.DownloadNextVersionForUpdate("maintainer", "app", currentTimestamp)

	assert.Nil(t, err)
	assert.True(t, response.UpdateAvailable)
	assert.NotNil(t, response.Version)
	assert.Equal(t, 2, response.Version.VersionId)
	assert.Equal(t, "2.0", response.Version.VersionName)
}
