package services

import (
	"encoding/gob"
	"net/rpc"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
)

func testClusterRPC(t *testing.T, c *rpc.Client) {
	gob.Register(map[string]any{})

	filters := make(map[string]any)
	var info map[string]any

	err := c.Call("CRPCService.GetInfo", filters, &info)
	assert.NoError(t, err)
	assert.Len(t, info, 6)

	assert.Equal(t, info["version"], agscheduler.Version)
}
