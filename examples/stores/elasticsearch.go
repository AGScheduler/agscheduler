// go run examples/stores/base.go examples/stores/elasticsearch.go

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	es8 "github.com/elastic/go-elasticsearch/v8"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/stores"
)

func main() {
	tClient, err := es8.NewTypedClient(es8.Config{
		Addresses: []string{"http://127.0.0.1:9200"},
	})
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create client: %s", err))
		os.Exit(1)
	}
	_, err = tClient.Ping().Do(context.Background())
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to database: %s", err))
		os.Exit(1)
	}
	store := &stores.ElasticsearchStore{
		TClient: tClient,
		Index:   "agscheduler_example_jobs",
	}

	scheduler := &agscheduler.Scheduler{}
	err = scheduler.SetStore(store)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}

	runExample(scheduler)
}
