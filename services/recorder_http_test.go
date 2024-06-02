package services

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
)

func testRecorderHTTP(t *testing.T, baseUrl string) {
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodDelete, baseUrl+"/recorder/records", nil)
	assert.NoError(t, err)
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	_, err = http.Post(baseUrl+"/scheduler/start", CONTENT_TYPE, nil)
	assert.NoError(t, err)

	mJ := map[string]any{
		"name":      "Job",
		"type":      agscheduler.JOB_TYPE_DATETIME,
		"start_at":  "2023-09-22 07:30:08",
		"func_name": "github.com/agscheduler/agscheduler/services.dryRunHTTP",
	}
	bJ, err := json.Marshal(mJ)
	assert.NoError(t, err)
	resp, err = http.Post(baseUrl+"/scheduler/job", CONTENT_TYPE, bytes.NewReader(bJ))
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ := &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)

	time.Sleep(1500 * time.Millisecond)

	jobId := rJ.Data.(map[string]any)["id"].(string)
	resp, err = http.Get(baseUrl + "/recorder/records" + "/" + jobId)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ = &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	total := int(rJ.Data.(map[string]any)["total"].(float64))
	assert.Equal(t, 1, total)

	resp, err = http.Post(baseUrl+"/scheduler/job", CONTENT_TYPE, bytes.NewReader(bJ))
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	time.Sleep(1500 * time.Millisecond)

	resp, err = http.Get(baseUrl + "/recorder/records")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ = &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	totalAll := int(rJ.Data.(map[string]any)["total"].(float64))
	assert.Equal(t, 2, totalAll)

	assert.Less(t, total, totalAll)

	req, err = http.NewRequest(http.MethodDelete, baseUrl+"/recorder/records"+"/"+jobId, nil)
	assert.NoError(t, err)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	resp, err = http.Get(baseUrl + "/recorder/records")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ = &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	totalAll = int(rJ.Data.(map[string]any)["total"].(float64))
	assert.Equal(t, 1, totalAll)

	req, err = http.NewRequest(http.MethodDelete, baseUrl+"/recorder/records", nil)
	assert.NoError(t, err)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	resp, err = http.Get(baseUrl + "/recorder/records")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ = &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	totalAll = int(rJ.Data.(map[string]any)["total"].(float64))
	assert.Equal(t, 0, totalAll)

	_, err = http.Post(baseUrl+"/scheduler/stop", CONTENT_TYPE, nil)
	assert.NoError(t, err)
}
