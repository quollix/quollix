package tools

import (
	"testing"

	"github.com/quollix/common/assert"
)

func TestUserAccessLevel_String(t *testing.T) {
	assert.Equal(t, "anonymous", AnonymousLevel.String())
	assert.Equal(t, "user", UserLevel.String())
	assert.Equal(t, "admin", AdminLevel.String())
}
