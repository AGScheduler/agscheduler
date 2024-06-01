package backends

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/agscheduler/agscheduler"
)

const (
	MONGODB_DATABASE   = "agscheduler"
	MONGODB_COLLECTION = "records"
)

// Store job records in a MongoDB database.
type MongoDBBackend struct {
	Client     *mongo.Client
	Database   string
	Collection string
	coll       *mongo.Collection
}

func (b *MongoDBBackend) Init() error {
	if b.Database == "" {
		b.Database = MONGODB_DATABASE
	}
	if b.Collection == "" {
		b.Collection = MONGODB_COLLECTION
	}

	b.coll = b.Client.Database(b.Database).Collection(b.Collection)

	iMJobId := mongo.IndexModel{
		Keys: bson.M{
			"job_id": 1,
		},
	}
	iMStartAt := mongo.IndexModel{
		Keys: bson.M{
			"start_at": -1,
		},
	}
	_, err := b.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{iMJobId, iMStartAt})
	if err != nil {
		return fmt.Errorf("failed to create index: %s", err)
	}

	return nil
}

func (b *MongoDBBackend) RecordMetadata(r agscheduler.Record) error {
	_, err := b.coll.InsertOne(ctx,
		bson.M{
			"_id":      r.Id,
			"job_id":   r.JobId,
			"job_name": r.JobName,
			"status":   r.Status,
			"result":   r.Result,
			"start_at": r.StartAt.Unix(),
			"end_at":   r.StartAt.Unix(),
		},
	)

	return err
}

func (b *MongoDBBackend) RecordResult(id uint64, status string, result []byte) error {
	var resultB bson.M
	err := b.coll.FindOneAndUpdate(ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"status": status,
				"result": result,
				"end_at": time.Now().UTC().Unix(),
			},
		},
	).Decode(&resultB)

	return err
}

func (b *MongoDBBackend) _getRecords(filter any) ([]agscheduler.Record, error) {
	opts := options.Find().SetSort(bson.M{"start_at": -1})
	cursor, err := b.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	var recordList []agscheduler.Record
	for cursor.Next(ctx) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		recordList = append(recordList, agscheduler.Record{
			Id:      uint64(result["_id"].(int64)),
			JobId:   result["job_id"].(string),
			JobName: result["job_name"].(string),
			Status:  result["status"].(string),
			Result:  result["result"].(primitive.Binary).Data,
			StartAt: time.Unix(result["start_at"].(int64), 0),
			EndAt:   time.Unix(result["end_at"].(int64), 0),
		})
	}

	return recordList, nil
}

func (b *MongoDBBackend) GetRecords(jId string) ([]agscheduler.Record, error) {
	return b._getRecords(bson.M{"job_id": jId})
}

func (b *MongoDBBackend) GetAllRecords() ([]agscheduler.Record, error) {
	return b._getRecords(bson.M{})
}

func (b *MongoDBBackend) DeleteRecords(jId string) error {
	_, err := b.coll.DeleteMany(ctx, bson.M{"job_id": jId})
	return err
}

func (b *MongoDBBackend) DeleteAllRecords() error {
	_, err := b.coll.DeleteMany(ctx, bson.M{})
	return err
}

func (b *MongoDBBackend) Clear() error {
	return b.Client.Database(b.Database).Collection(b.Collection).Drop(ctx)
}
