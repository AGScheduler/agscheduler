package services

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/stores"
)

func testHTTPAuth(t *testing.T, baseUrl string) {
	client := &http.Client{}

	resp, err := http.Get(baseUrl + "/info")
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	req, err := http.NewRequest(http.MethodGet, baseUrl+"/info", bytes.NewReader([]byte{}))
	assert.NoError(t, err)
	req.Header.Add("Auth-Password-SHA2", "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918")
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestAuth(t *testing.T) {
	scheduler := &agscheduler.Scheduler{}

	store := &stores.MemoryStore{}
	err := scheduler.SetStore(store)
	assert.NoError(t, err)

	hservice := HTTPService{
		Scheduler:    scheduler,
		PasswordSha2: "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918",
	}
	err = hservice.Start()
	assert.NoError(t, err)

	time.Sleep(time.Second)

	baseUrl := "http://" + hservice.Address
	testHTTPAuth(t, baseUrl)

	err = hservice.Stop()
	assert.NoError(t, err)

	err = store.Clear()
	assert.NoError(t, err)
}
