package stores

import (
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler/internal/adaptertest"
)

func TestMongoDBStore(t *testing.T) {
	uri := "mongodb://127.0.0.1:27017/"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	assert.NoError(t, err)
	defer func() {
		err = client.Disconnect(ctx)
		assert.NoError(t, err)
	}()

	store := &MongoDBStore{
		Client:     client,
		Database:   MONGODB_DATABASE,
		Collection: "test_jobs",
	}

	adaptertest.RunStore(t, store)
}

func TestMongoDBStoreEmptyNextRunTime(t *testing.T) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:27017/"))
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, client.Disconnect(ctx))
	}()

	store := &MongoDBStore{
		Client:     client,
		Database:   MONGODB_DATABASE,
		Collection: fmt.Sprintf("test_empty_jobs_%d", time.Now().UnixNano()),
	}
	assert.NoError(t, store.Init())
	defer func() {
		assert.NoError(t, store.Clear())
	}()

	nextRunTime, err := store.GetNextRunTime()
	assert.NoError(t, err)
	assert.True(t, nextRunTime.IsZero())
}
