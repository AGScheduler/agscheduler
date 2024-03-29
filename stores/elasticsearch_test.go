package stores

import (
	"testing"

	es8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
)

func TestElasticsearchStore(t *testing.T) {
	tClient, err := es8.NewTypedClient(es8.Config{
		Addresses: []string{"http://127.0.0.1:9200"},
	})
	assert.NoError(t, err)
	_, err = tClient.Ping().Do(ctx)
	assert.NoError(t, err)
	store := &ElasticsearchStore{
		TClient: tClient,
		Index:   "agscheduler_test_jobs",
	}

	scheduler := &agscheduler.Scheduler{}
	err = scheduler.SetStore(store)
	assert.NoError(t, err)

	testAGScheduler(t, scheduler)

	err = store.Clear()
	assert.NoError(t, err)
}
