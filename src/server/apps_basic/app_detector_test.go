package apps_basic

import (
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
)

var detector = &AppDetectorImpl{}

func TestAppDetectorImpl_IsOfficialApp(t *testing.T) {
	assert.True(t, detector.IsOfficialApp(u.OfficialMaintainer))
	assert.False(t, detector.IsOfficialApp("samplemaintainer"))
}

func TestAppDetectorImpl_IsOfficialDatabaseApp(t *testing.T) {
	assert.True(t, detector.IsOfficialDatabaseApp(u.OfficialDatabaseAppName))
	assert.False(t, detector.IsOfficialDatabaseApp("sampleapp"))
}

func TestAppDetectorImpl_IsSystemApp(t *testing.T) {
	assert.True(t, detector.IsSystemApp(u.OfficialDatabaseAppName))
	assert.True(t, detector.IsSystemApp(u.OfficialBrandAppName))
	assert.True(t, detector.IsSystemApp(tools.QuollogAppName))
	assert.False(t, detector.IsSystemApp("sampleapp"))
}
