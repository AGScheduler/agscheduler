// 1. go run examples/http/http_server/main.go
// 2. go run examples/http/http_client/main.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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
		"func_name": "github.com/agscheduler/agscheduler/examples.PrintMsg",
		"args":      map[string]any{"arg1": "1", "arg2": "2", "arg3": "3"},
	}
	bJob1, _ := json.Marshal(mJob1)
	req, _ := http.NewRequest(http.MethodPost, baseUrl+"/scheduler/job", bytes.NewReader(bJob1))
	// req.Header.Add("Auth-Password-SHA2", "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918")
	resp, _ := client.Do(req)
	body, _ := io.ReadAll(resp.Body)
	rJob1 := &result{}
	_ = json.Unmarshal(body, &rJob1)
	slog.Info(fmt.Sprintf("%s.\n\n", rJob1.Data))

	_, _ = http.Post(baseUrl+"/scheduler/start", CONTENT_TYPE, nil)
}

func main() {
	baseUrl := "http://127.0.0.1:36370"

	runExampleHTTP(baseUrl)
}
