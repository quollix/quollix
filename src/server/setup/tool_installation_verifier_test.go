package setup

import (
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
	u "github.com/quollix/common/utils"
	"github.com/quollix/deepstack"
)

func TestCrashIfToolIsNotInstalled_InstalledToolReturnsNil(t *testing.T) {
	err := crashIfToolIsNotInstalled("go", []string{"version"})

	assert.Nil(t, err)
}

func TestCrashIfToolIsNotInstalled_MissingToolReturnsError(t *testing.T) {
	toolName := "quollix-command-that-should-not-exist"
	err := crashIfToolIsNotInstalled(toolName, []string{"version"})

	assert.Equal(t, "tried command but CLI tool seems not to be installed properly", u.ExtractError(err))
	deepStackError := err.(*deepstack.DeepStackError)
	assert.Equal(t, toolName, deepStackError.Context[tools.CommandField])
	assert.Equal(t, "version", deepStackError.Context[tools.CommandArgsField])
}
