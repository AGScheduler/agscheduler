package services

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testRecorderHTTP(t *testing.T, baseUrl string) {
	resp, err := http.Get(baseUrl + "/recorder/records/test")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ := &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	total := rJ.Data.(map[string]any)["total"].(float64)

	resp, err = http.Get(baseUrl + "/recorder/records")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ = &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	totalAll := rJ.Data.(map[string]any)["total"].(float64)

	assert.Less(t, total, totalAll)
}
