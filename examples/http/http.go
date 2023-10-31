// go run examples/http/http.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/examples"
	"github.com/kwkwc/agscheduler/services"
	"github.com/kwkwc/agscheduler/stores"
)

const CONTENT_TYPE = "application/json"

type result struct {
	Data  any    `json:"data"`
	Error string `json:"error"`
}

func runExampleHTTP(baseUrl string) {
	client := &http.Client{}

	mJob1 := map[string]any{
		"name":      "Job1",
		"type":      "interval",
		"interval":  "2s",
		"timezone":  "UTC",
		"func_name": "github.com/kwkwc/agscheduler/examples.PrintMsg",
		"args":      map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	bJob1, _ := json.Marshal(mJob1)
	resp, _ := http.Post(baseUrl+"/scheduler/job", CONTENT_TYPE, bytes.NewReader(bJob1))
	body, _ := io.ReadAll(resp.Body)
	rJob1 := &result{}
	json.Unmarshal(body, &rJob1)
	slog.Info(fmt.Sprintf("%s.\n\n", rJob1.Data))

	mJob2 := map[string]any{
		"name":      "Job2",
		"type":      "cron",
		"cron_expr": "*/1 * * * *",
		"timezone":  "Asia/Shanghai",
		"func_name": "github.com/kwkwc/agscheduler/examples.PrintMsg",
		"args":      map[string]any{"arg4": "4", "arg5": "5", "arg6": "6", "arg7": "7"},
	}
	bJob2, _ := json.Marshal(mJob2)
	resp, _ = http.Post(baseUrl+"/scheduler/job", CONTENT_TYPE, bytes.NewReader(bJob2))
	body, _ = io.ReadAll(resp.Body)
	rJob2 := &result{}
	json.Unmarshal(body, &rJob2)
	slog.Info(fmt.Sprintf("%s.\n\n", rJob2.Data))

	http.Post(baseUrl+"/scheduler/start", CONTENT_TYPE, nil)

	mJob3 := map[string]any{
		"name":      "Job3",
		"type":      "datetime",
		"start_at":  "2023-09-22 07:30:08",
		"timezone":  "America/New_York",
		"func_name": "github.com/kwkwc/agscheduler/examples.PrintMsg",
		"args":      map[string]any{"arg8": "8", "arg9": "9"},
	}
	bJob3, _ := json.Marshal(mJob3)
	resp, _ = http.Post(baseUrl+"/scheduler/job", CONTENT_TYPE, bytes.NewReader(bJob3))
	body, _ = io.ReadAll(resp.Body)
	rJob3 := &result{}
	json.Unmarshal(body, &rJob3)
	slog.Info(fmt.Sprintf("%s.\n\n", rJob3.Data))

	resp, _ = http.Get(baseUrl + "/scheduler/jobs")
	body, _ = io.ReadAll(resp.Body)
	res := &result{}
	json.Unmarshal(body, &res)
	slog.Info(fmt.Sprintf("Scheduler get all jobs %s.\n\n", res.Data))

	slog.Info("Sleep 5s......\n\n")
	time.Sleep(5 * time.Second)

	resp, _ = http.Get(baseUrl + "/scheduler/job" + "/" + rJob1.Data.(map[string]any)["id"].(string))
	body, _ = io.ReadAll(resp.Body)
	rJob1 = &result{}
	json.Unmarshal(body, &rJob1)
	slog.Info(fmt.Sprintf("Scheduler get job `%s:%s` %s.\n\n", rJob1.Data.(map[string]any)["id"].(string), rJob1.Data.(map[string]any)["name"].(string), rJob1.Data))

	mJob2["id"] = rJob2.Data.(map[string]any)["id"].(string)
	mJob2["timeout"] = rJob2.Data.(map[string]any)["timeout"].(string)
	mJob2["type"] = "interval"
	mJob2["interval"] = "3s"
	bJob2, _ = json.Marshal(mJob2)
	req, _ := http.NewRequest(http.MethodPut, baseUrl+"/scheduler/job", bytes.NewReader(bJob2))
	req.Header.Add("content-type", CONTENT_TYPE)
	resp, _ = client.Do(req)
	body, _ = io.ReadAll(resp.Body)
	rJob2 = &result{}
	json.Unmarshal(body, &rJob2)
	slog.Info(fmt.Sprintf("Scheduler update job `%s:%s` %s.\n\n", rJob2.Data.(map[string]any)["id"].(string), rJob2.Data.(map[string]any)["name"].(string), rJob2.Data))

	slog.Info("Sleep 4s......")
	time.Sleep(4 * time.Second)

	http.Post(baseUrl+"/scheduler/job/"+rJob1.Data.(map[string]any)["id"].(string)+"/pause", CONTENT_TYPE, nil)

	slog.Info("Sleep 3s......\n\n")
	time.Sleep(3 * time.Second)

	http.Post(baseUrl+"/scheduler/job/"+rJob1.Data.(map[string]any)["id"].(string)+"/resume", CONTENT_TYPE, nil)

	req, _ = http.NewRequest(http.MethodDelete, baseUrl+"/scheduler/job"+"/"+rJob2.Data.(map[string]any)["id"].(string), nil)
	client.Do(req)

	slog.Info("Sleep 4s......\n\n")
	time.Sleep(4 * time.Second)

	http.Post(baseUrl+"/scheduler/stop", CONTENT_TYPE, nil)

	bJob1, _ = json.Marshal(rJob1.Data.(map[string]any))
	http.Post(baseUrl+"/scheduler/job/run", CONTENT_TYPE, bytes.NewReader(bJob1))

	slog.Info("Sleep 3s......\n\n")
	time.Sleep(3 * time.Second)

	http.Post(baseUrl+"/scheduler/start", CONTENT_TYPE, nil)

	slog.Info("Sleep 3s......\n\n")
	time.Sleep(3 * time.Second)

	req, _ = http.NewRequest(http.MethodDelete, baseUrl+"/scheduler/jobs", nil)
	client.Do(req)
}

func main() {
	agscheduler.RegisterFuncs(examples.PrintMsg)

	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}

	shservice := services.SchedulerHTTPService{
		Scheduler: scheduler,
		Address:   "127.0.0.1:63636",
	}
	err = shservice.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to start service: %s", err))
		os.Exit(1)
	}

	time.Sleep(time.Second)

	baseUrl := "http://127.0.0.1:63636"

	runExampleHTTP(baseUrl)
}
