package api_client

import (
	"server/tools"
)

type QuollixTestClient struct {
	quollix *QuollixClient
}

func (c *QuollixTestClient) ResetTestState() error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendResetTestState, nil)
	return err
}
