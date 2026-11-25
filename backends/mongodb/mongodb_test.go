package backends

import (
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler/internal/adaptertest"

	"github.com/agscheduler/agscheduler"
)

func TestMongoDBBackend(t *testing.T) {
	uri := "mongodb://127.0.0.1:27017/"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	assert.NoError(t, err)
	defer func() {
		err = client.Disconnect(ctx)
		assert.NoError(t, err)
	}()

	backend := &MongoDBBackend{
		Client:     client,
		Database:   MONGODB_DATABASE,
		Collection: "test_records",
	}
	recorder := &agscheduler.Recorder{Backend: backend}

	adaptertest.RunBackend(t, recorder)
}
