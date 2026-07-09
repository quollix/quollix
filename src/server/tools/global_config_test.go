package tools

import (
	"testing"

	"github.com/quollix/common/assert"
)

func TestNewGlobalConfigFromEnv_RedirectHttpToHttpsDefaultsToTrue(t *testing.T) {
	t.Setenv(RedirectHttpToHttpsEnvVar, "")
	config := NewGlobalConfigFromEnv()
	assert.True(t, config.RedirectHttpToHttps)
}

func TestNewGlobalConfigFromEnv_RedirectHttpToHttpsCanBeDisabled(t *testing.T) {
	t.Setenv(RedirectHttpToHttpsEnvVar, "false")
	config := NewGlobalConfigFromEnv()
	assert.False(t, config.RedirectHttpToHttps)
}

func TestNewGlobalConfigFromEnv_RedirectHttpToHttpsOnlyDisablesOnExplicitFalse(t *testing.T) {
	t.Setenv(RedirectHttpToHttpsEnvVar, "FALSE")
	config := NewGlobalConfigFromEnv()
	assert.True(t, config.RedirectHttpToHttps)
}

func TestNewGlobalConfigFromEnv_AppForwardedProtoDefaultsToHttps(t *testing.T) {
	t.Setenv(AppForwardedProtoEnvVar, "")
	config := NewGlobalConfigFromEnv()
	assert.Equal(t, AppForwardedProtoHttps, config.AppForwardedProto)
}

func TestNewGlobalConfigFromEnv_AppForwardedProtoCanBeSetToHttp(t *testing.T) {
	t.Setenv(AppForwardedProtoEnvVar, AppForwardedProtoHttp)
	config := NewGlobalConfigFromEnv()
	assert.Equal(t, AppForwardedProtoHttp, config.AppForwardedProto)
}

func TestNewGlobalConfigFromEnv_AppForwardedProtoDefaultsToHttpsForUnknownValues(t *testing.T) {
	t.Setenv(AppForwardedProtoEnvVar, "ftp")
	config := NewGlobalConfigFromEnv()
	assert.Equal(t, AppForwardedProtoHttps, config.AppForwardedProto)
}
