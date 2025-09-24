// go run examples/backends/mongodb/main.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/backends"
	eb "github.com/agscheduler/agscheduler/examples/backends"
)

func main() {
	uri := "mongodb://127.0.0.1:27017/"
	client, err := mongo.Connect(eb.Ctx, options.Client().ApplyURI(uri))
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to database: %s", err))
		os.Exit(1)
	}
	defer func() {
		_ = client.Disconnect(eb.Ctx)
	}()

	backend := &backends.MongoDBBackend{
		Client:     client,
		Database:   backends.MONGODB_DATABASE,
		Collection: "example_records",
	}
	recorder := &agscheduler.Recorder{Backend: backend}

	eb.RunExample(recorder)
}
