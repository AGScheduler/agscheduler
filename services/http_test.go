package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/backends"
	"github.com/agscheduler/agscheduler/stores"
)

type result struct {
	Data  any    `json:"data"`
	Error string `json:"error"`
}

func dryRunHTTP(ctx context.Context, j agscheduler.Job) (result string) { return }

func testHTTP(t *testing.T, baseUrl string) {
	resp, err := http.Get(baseUrl + "/info")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ := &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	assert.Len(t, rJ.Data.(map[string]any), 5)
	assert.Equal(t, agscheduler.Version, rJ.Data.(map[string]any)["version"])

	resp, err = http.Get(baseUrl + "/funcs")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ = &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	funcLen := len(agscheduler.FuncMap)
	assert.Len(t, rJ.Data, funcLen)
}

func TestHTTPService(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	agscheduler.RegisterFuncs(
		agscheduler.FuncPkg{Func: dryRunHTTP},
	)

	scheduler := &agscheduler.Scheduler{}

	store := &stores.MemoryStore{}
	err := scheduler.SetStore(store)
	assert.NoError(t, err)

	mb := &backends.MemoryBackend{}
	recorder := &agscheduler.Recorder{Backend: mb}
	err = scheduler.SetRecorder(recorder)
	assert.NoError(t, err)

	hservice := HTTPService{Scheduler: scheduler}
	err = hservice.Start()
	assert.NoError(t, err)

	time.Sleep(time.Second)

	baseUrl := "http://" + hservice.Address
	testHTTP(t, baseUrl)
	testSchedulerHTTP(t, baseUrl)
	testRecorderHTTP(t, baseUrl)

	err = hservice.Stop()
	assert.NoError(t, err)

	err = store.Clear()
	assert.NoError(t, err)
}
