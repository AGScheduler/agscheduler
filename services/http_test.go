package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kwkwc/agscheduler"
	"github.com/kwkwc/agscheduler/stores"
)

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
	testSchedulerHTTP(t, baseUrl)

	err = hservice.Stop()
	assert.NoError(t, err)

	err = store.Clear()
	assert.NoError(t, err)
}
