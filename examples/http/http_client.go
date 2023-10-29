// 1. go run examples/http/http_server.go
// 2. go run examples/http/http_client.go

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
}

func main() {
	baseUrl := "http://127.0.0.1:63636"

	runExampleHTTP(baseUrl)
}
