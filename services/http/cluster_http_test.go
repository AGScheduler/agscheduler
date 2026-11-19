package services

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testClusterHTTP(t *testing.T, baseUrl string) {
	resp, err := http.Get(baseUrl + "/cluster/nodes")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ := &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	assert.Len(t, rJ.Data.(map[string]any), 2)
}
