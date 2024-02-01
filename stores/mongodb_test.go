package stores

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/kwkwc/agscheduler"
)

func TestMongoDBStore(t *testing.T) {
	uri := "mongodb://127.0.0.1:27017/"
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	assert.NoError(t, err)
	defer func() {
		err := client.Disconnect(context.Background())
		assert.NoError(t, err)
	}()
	store := &MongoDBStore{Client: client, Collection: "test_jobs"}

	scheduler := &agscheduler.Scheduler{}
	err = scheduler.SetStore(store)
	assert.NoError(t, err)

	testAGScheduler(t, scheduler)

	err = store.Clear()
	assert.NoError(t, err)
}
