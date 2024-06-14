package services

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testBrokerHTTP(t *testing.T, baseUrl string) {
	resp, err := http.Get(baseUrl + "/broker/queues")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	res := &result{}
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err)
	assert.Len(t, res.Data.([]any), 1)
}
