package services

import (
	"bytes"
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

func dryRunHTTP(j agscheduler.Job) {}

func testAGSchedulerHTTP(t *testing.T, baseUrl string) {
	client := &http.Client{}

	http.Post(baseUrl+"/scheduler/start", CONTENT_TYPE, nil)

	mJ := map[string]any{
		"name":      "Job",
		"type":      agscheduler.TYPE_INTERVAL,
		"interval":  "1s",
		"func_name": "github.com/kwkwc/agscheduler/services.dryRunHTTP",
		"args":      map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	bJ, _ := json.Marshal(mJ)
	resp, _ := http.Post(baseUrl+"/scheduler/job", CONTENT_TYPE, bytes.NewReader(bJ))
	body, _ := io.ReadAll(resp.Body)
	rJ := &result{}
	json.Unmarshal(body, &rJ)
	assert.Equal(t, agscheduler.STATUS_RUNNING, rJ.Data.(map[string]any)["status"].(string))

	id := rJ.Data.(map[string]any)["id"].(string)
	mJ["id"] = id
	mJ["type"] = "cron"
	mJ["cron_expr"] = "*/1 * * * *"
	bJ, _ = json.Marshal(mJ)
	req, _ := http.NewRequest(http.MethodPut, baseUrl+"/scheduler/job", bytes.NewReader(bJ))
	req.Header.Add("content-type", CONTENT_TYPE)
	resp, _ = client.Do(req)
	body, _ = io.ReadAll(resp.Body)
	rJ = &result{}
	json.Unmarshal(body, &rJ)
	assert.Equal(t, agscheduler.TYPE_CRON, rJ.Data.(map[string]any)["type"].(string))

	timezone, _ := time.LoadLocation(rJ.Data.(map[string]any)["timezone"].(string))
	nextRunTimeMax, _ := time.ParseInLocation(time.DateTime, "9999-09-09 09:09:09", timezone)

	resp, _ = http.Post(baseUrl+"/scheduler/job/"+id+"/pause", CONTENT_TYPE, nil)
	body, _ = io.ReadAll(resp.Body)
	rJ = &result{}
	json.Unmarshal(body, &rJ)
	assert.Equal(t, agscheduler.STATUS_PAUSED, rJ.Data.(map[string]any)["status"].(string))
	nextRunTime, _ := time.ParseInLocation(time.RFC3339, rJ.Data.(map[string]any)["next_run_time"].(string), timezone)
	assert.Equal(t, nextRunTimeMax.Unix(), nextRunTime.Unix())

	resp, _ = http.Post(baseUrl+"/scheduler/job/"+id+"/resume", CONTENT_TYPE, nil)
	body, _ = io.ReadAll(resp.Body)
	rJ = &result{}
	json.Unmarshal(body, &rJ)
	nextRunTime, _ = time.ParseInLocation(time.RFC3339, rJ.Data.(map[string]any)["next_run_time"].(string), timezone)
	assert.NotEqual(t, nextRunTimeMax.Unix(), nextRunTime.Unix())

	req, _ = http.NewRequest(http.MethodDelete, baseUrl+"/scheduler/job"+"/"+id, nil)
	client.Do(req)
	resp, _ = http.Get(baseUrl + "/scheduler/job" + "/" + id)
	body, _ = io.ReadAll(resp.Body)
	rJ = &result{}
	json.Unmarshal(body, &rJ)
	assert.Equal(t, rJ.Error, agscheduler.JobNotFoundError(id).Error())

	req, _ = http.NewRequest(http.MethodDelete, baseUrl+"/scheduler/jobs", nil)
	client.Do(req)
	resp, _ = http.Get(baseUrl + "/scheduler/jobs")
	body, _ = io.ReadAll(resp.Body)
	rJs := &result{}
	json.Unmarshal(body, &rJs)
	assert.Empty(t, rJs.Data)

	http.Post(baseUrl+"/scheduler/stop", CONTENT_TYPE, nil)
}

func TestHTTPService(t *testing.T) {
	agscheduler.RegisterFuncs(dryRunHTTP)

	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	hservice := SchedulerHTTPService{Scheduler: scheduler}
	hservice.Start("127.0.0.1:63636")

	time.Sleep(time.Second)

	baseUrl := "http://127.0.0.1:63636"

	testAGSchedulerHTTP(t, baseUrl)

	store.Clean()
}
