package component

import (
	"server/tools"

	"github.com/quollix/common/assert"
)

type QuollixTestClient struct {
	quollix *QuollixClient
}

func (c *QuollixTestClient) ResetTestState() {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendResetTestState, nil)
	assert.Nil(c.quollix.T, err)
}
