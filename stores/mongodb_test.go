package stores

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoDBStore(t *testing.T) {
	uri := "mongodb://127.0.0.1:27017/"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	assert.NoError(t, err)
	defer client.Disconnect(ctx)

	store := &MongoDBStore{
		Client:     client,
		Database:   MONGODB_DATABASE,
		Collection: "test_jobs",
	}

	runTest(t, store)
}
