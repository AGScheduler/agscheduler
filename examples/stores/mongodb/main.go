// go run examples/stores/mongodb/main.go

package main

import (
	"fmt"
	"log/slog"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	es "github.com/agscheduler/agscheduler/examples/stores"
	"github.com/agscheduler/agscheduler/stores"
)

func main() {
	uri := "mongodb://127.0.0.1:27017/"
	client, err := mongo.Connect(es.Ctx, options.Client().ApplyURI(uri))
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to database: %s", err))
		os.Exit(1)
	}
	defer func() {
		_ = client.Disconnect(es.Ctx)
	}()

	store := &stores.MongoDBStore{
		Client:     client,
		Database:   stores.MONGODB_DATABASE,
		Collection: "example_jobs",
	}

	es.RunExample(store)
}
