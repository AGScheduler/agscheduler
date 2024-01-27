package services

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/stores"
)

func testHTTP(t *testing.T, baseUrl string) {
	resp, err := http.Get(baseUrl + "/info")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	rJ := &result{}
	err = json.Unmarshal(body, &rJ)
	assert.NoError(t, err)
	assert.Len(t, rJ.Data.(map[string]any), 4)

	assert.Equal(t, agscheduler.Version, rJ.Data.(map[string]any)["version"])
}

func TestHTTPService(t *testing.T) {
	agscheduler.RegisterFuncs(dryRunHTTP)

	store := &stores.MemoryStore{}

	scheduler := &agscheduler.Scheduler{}
	err := scheduler.SetStore(store)
	assert.NoError(t, err)

	hservice := HTTPService{Scheduler: scheduler}
	err = hservice.Start()
	assert.NoError(t, err)

	time.Sleep(time.Second)

	baseUrl := "http://" + hservice.Address
	testHTTP(t, baseUrl)
	testSchedulerHTTP(t, baseUrl)

	err = hservice.Stop()
	assert.NoError(t, err)

	err = store.Clear()
	assert.NoError(t, err)
}
