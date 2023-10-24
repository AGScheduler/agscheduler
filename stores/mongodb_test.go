package stores

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/kwkwc/agscheduler"
)

func TestMongoDBStore(t *testing.T) {
	uri := "mongodb://127.0.0.1:27017/"
	client, _ := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	defer client.Disconnect(context.Background())
	store := &MongoDBStore{Client: client, Collection: "test_jobs"}

	scheduler := &agscheduler.Scheduler{}
	scheduler.SetStore(store)

	testAGScheduler(t, scheduler)

	store.Clear()
}
