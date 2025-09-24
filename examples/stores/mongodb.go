// go run examples/stores/base.go examples/stores/mongodb.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/agscheduler/agscheduler/stores"
)

func main() {
	uri := "mongodb://127.0.0.1:27017/"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to database: %s", err))
		os.Exit(1)
	}
	defer client.Disconnect(ctx)

	store := &stores.MongoDBStore{
		Client:     client,
		Database:   stores.MONGODB_DATABASE,
		Collection: "example_jobs",
	}

	runExample(store)
}
