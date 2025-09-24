// go run examples/backends/base.go examples/backends/mongodb.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/backends"
)

func main() {
	uri := "mongodb://127.0.0.1:27017/"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to database: %s", err))
		os.Exit(1)
	}
	defer client.Disconnect(ctx)

	backend := &backends.MongoDBBackend{
		Client:     client,
		Database:   backends.MONGODB_DATABASE,
		Collection: "example_records",
	}
	recorder := &agscheduler.Recorder{Backend: backend}

	runExample(recorder)
}
