package backups

import (
	"testing"

	"github.com/quollix/common/assert"
	"github.com/quollix/common/quollix/api"
)

func TestFindUniqueMaintainerAndAppNamePairs(t *testing.T) {
	service := &BackupQueryServiceImpl{}

	input := []api.BackupInfo{
		{Maintainer: "a", AppName: "x"},
		{Maintainer: "a", AppName: "y"},
		{Maintainer: "b", AppName: "y"},
		{Maintainer: "a", AppName: "x"},
		{Maintainer: "a", AppName: "z"},
	}

	expected := []api.MaintainerAndApp{
		{Maintainer: "a", AppName: "x"},
		{Maintainer: "a", AppName: "y"},
		{Maintainer: "b", AppName: "y"},
		{Maintainer: "a", AppName: "z"},
	}

	actual := service.UniqueMaintainerAndAppPairs(input)
	assert.Equal(t, expected, actual)
}
