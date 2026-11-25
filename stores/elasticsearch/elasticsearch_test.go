package stores

import (
	"fmt"
	"testing"
	"time"

	es8 "github.com/elastic/go-elasticsearch/v8"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/internal/adaptertest"
)

func newTestClient(t *testing.T) *es8.TypedClient {
	t.Helper()

	tClient, err := es8.NewTypedClient(es8.Config{
		Addresses: []string{"http://127.0.0.1:9200"},
	})
	assert.NoError(t, err)
	_, err = tClient.Ping().Do(ctx)
	assert.NoError(t, err)
	return tClient
}

func TestElasticsearchStore(t *testing.T) {
	tClient := newTestClient(t)

	store := &ElasticsearchStore{
		TClient: tClient,
		Index:   "agscheduler_test_jobs",
	}

	adaptertest.RunStore(t, store)
}

func TestElasticsearchStoreReturnsAllJobs(t *testing.T) {
	store := &ElasticsearchStore{
		TClient: newTestClient(t),
		Index:   fmt.Sprintf("agscheduler_test_all_jobs_%d", time.Now().UnixNano()),
	}
	assert.NoError(t, store.Init())
	t.Cleanup(func() {
		assert.NoError(t, store.Clear())
	})

	nextRunTime, err := store.GetNextRunTime()
	assert.NoError(t, err)
	assert.True(t, nextRunTime.IsZero())

	start := time.Now().UTC().Truncate(time.Second).Add(time.Minute)
	for i := 0; i < 12; i++ {
		assert.NoError(t, store.AddJob(agscheduler.Job{
			Id:          fmt.Sprintf("job-%02d", i),
			Name:        fmt.Sprintf("Job %d", i),
			Type:        agscheduler.JOB_TYPE_INTERVAL,
			Interval:    "1h",
			NextRunTime: start.Add(time.Duration(i) * time.Minute),
		}))
	}

	jobs, err := store.GetAllJobs()
	assert.NoError(t, err)
	assert.Len(t, jobs, 12)
	assert.Equal(t, "job-00", jobs[0].Id)
	assert.Equal(t, "job-11", jobs[11].Id)

	nextRunTime, err = store.GetNextRunTime()
	assert.NoError(t, err)
	assert.Equal(t, start, nextRunTime)
}
