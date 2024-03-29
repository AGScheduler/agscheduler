// go run examples/stores/base.go examples/stores/mongodb.go

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/stores"
)

func main() {
	uri := "mongodb://127.0.0.1:27017/"
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to database: %s", err))
		os.Exit(1)
	}
	defer client.Disconnect(context.Background())
	store := &stores.MongoDBStore{Client: client, Collection: "example_jobs"}

	scheduler := &agscheduler.Scheduler{}
	err = scheduler.SetStore(store)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to set store: %s", err))
		os.Exit(1)
	}

	runExample(scheduler)
}
