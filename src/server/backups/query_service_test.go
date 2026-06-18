package backups

import (
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
)

func TestFindUniqueMaintainerAndAppNamePairs(t *testing.T) {
	service := &BackupQueryServiceImpl{}

	input := []tools.BackupInfo{
		{Maintainer: "a", AppName: "x"},
		{Maintainer: "a", AppName: "y"},
		{Maintainer: "b", AppName: "y"},
		{Maintainer: "a", AppName: "x"},
		{Maintainer: "a", AppName: "z"},
	}

	expected := []tools.MaintainerAndApp{
		{Maintainer: "a", AppName: "x"},
		{Maintainer: "a", AppName: "y"},
		{Maintainer: "b", AppName: "y"},
		{Maintainer: "a", AppName: "z"},
	}

	actual := service.UniqueMaintainerAndAppPairs(input)
	assert.Equal(t, expected, actual)
}
