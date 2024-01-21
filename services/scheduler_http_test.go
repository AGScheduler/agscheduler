package services

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/stores"
)

const CONTENT_TYPE = "application/json"

type result struct {
	Data  any    `json:"data"`
	Error string `json:"error"`
}

func dryRunHTTP(ctx context.Context, j agscheduler.Job) {}

func testAGSchedulerHTTP(t *testing.T, baseUrl string) {
	client := &http.Client{}

	_, err := http.Post(baseUrl+"/scheduler/start", CONTENT_TYPE, nil)
	assert.NoError(t, err)

	mJ := map[string]any{
		"name":      "Job",
		"type":      agscheduler.TYPE_INTERVAL,
		"interval":  "1s",
		"func_name": "github.com/kwkwc/agscheduler/services.dryRunHTTP",
		"args":      map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	bJ, err := json.Marshal(mJ)
	assert.NoError(t, err)
	resp, err := http.Post(baseUrl+"/scheduler/job", CONTENT_TYPE, bytes.NewReader(bJ))
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ := &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.STATUS_RUNNING, rJ.Data.(map[string]any)["status"].(string))

	id := rJ.Data.(map[string]any)["id"].(string)
	mJ["id"] = id
	mJ["timeout"] = rJ.Data.(map[string]any)["timeout"].(string)
	mJ["type"] = "cron"
	mJ["cron_expr"] = "*/1 * * * *"
	bJ, err = json.Marshal(mJ)
	assert.NoError(t, err)
	req, err := http.NewRequest(http.MethodPut, baseUrl+"/scheduler/job", bytes.NewReader(bJ))
	assert.NoError(t, err)
	req.Header.Add("content-type", CONTENT_TYPE)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ = &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.TYPE_CRON, rJ.Data.(map[string]any)["type"].(string))

	timezone, err := time.LoadLocation(rJ.Data.(map[string]any)["timezone"].(string))
	assert.NoError(t, err)
	nextRunTimeMax, err := time.ParseInLocation(time.DateTime, "9999-09-09 09:09:09", timezone)
	assert.NoError(t, err)

	resp, err = http.Post(baseUrl+"/scheduler/job/"+id+"/pause", CONTENT_TYPE, nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ = &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.STATUS_PAUSED, rJ.Data.(map[string]any)["status"].(string))
	nextRunTime, err := time.ParseInLocation(time.RFC3339, rJ.Data.(map[string]any)["next_run_time"].(string), timezone)
	assert.NoError(t, err)
	assert.Equal(t, nextRunTimeMax.Unix(), nextRunTime.Unix())

	resp, err = http.Post(baseUrl+"/scheduler/job/"+id+"/resume", CONTENT_TYPE, nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ = &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	nextRunTime, err = time.ParseInLocation(time.RFC3339, rJ.Data.(map[string]any)["next_run_time"].(string), timezone)
	assert.NoError(t, err)
	assert.NotEqual(t, nextRunTimeMax.Unix(), nextRunTime.Unix())

	bJ, err = json.Marshal(rJ.Data.(map[string]any))
	assert.NoError(t, err)
	resp, err = http.Post(baseUrl+"/scheduler/job/run", CONTENT_TYPE, bytes.NewReader(bJ))
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ = &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	assert.Empty(t, rJ.Error)

	resp, err = http.Post(baseUrl+"/scheduler/job/schedule", CONTENT_TYPE, bytes.NewReader(bJ))
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ = &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	assert.Empty(t, rJ.Error)

	req, err = http.NewRequest(http.MethodDelete, baseUrl+"/scheduler/job"+"/"+id, nil)
	assert.NoError(t, err)
	client.Do(req)
	resp, err = http.Get(baseUrl + "/scheduler/job" + "/" + id)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ = &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.JobNotFoundError(id).Error(), rJ.Error)

	req, err = http.NewRequest(http.MethodDelete, baseUrl+"/scheduler/jobs", nil)
	assert.NoError(t, err)
	client.Do(req)
	resp, err = http.Get(baseUrl + "/scheduler/jobs")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJs := &result{}
	err = json.Unmarshal(body, &rJs)
	assert.NoError(t, err)
	assert.Empty(t, rJs.Data)

	_, err = http.Post(baseUrl+"/scheduler/stop", CONTENT_TYPE, nil)
	assert.NoError(t, err)
}

func TestHTTPService(t *testing.T) {
	agscheduler.RegisterFuncs(dryRunHTTP)

	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	assert.NoError(t, err)

	shservice := SchedulerHTTPService{
		Scheduler: scheduler,
		// Address:   "127.0.0.1:36370",
	}
	err = shservice.Start()
	assert.NoError(t, err)

	time.Sleep(time.Second)

	baseUrl := "http://" + shservice.Address

	testAGSchedulerHTTP(t, baseUrl)

	err = shservice.Stop()
	assert.NoError(t, err)

	err = store.Clear()
	assert.NoError(t, err)
}
