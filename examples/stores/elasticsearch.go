// go run examples/stores/base.go examples/stores/elasticsearch.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	es8 "github.com/elastic/go-elasticsearch/v8"

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
	_, err = tClient.Ping().Do(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to database: %s", err))
		os.Exit(1)
	}

	store := &stores.ElasticsearchStore{
		TClient: tClient,
		Index:   "agscheduler_example_jobs",
	}

	runExample(store)
}
