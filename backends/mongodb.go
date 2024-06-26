package backends

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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

func (b *MongoDBBackend) Name() string {
	return "MongoDB"
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

func (b *MongoDBBackend) RecordResult(id uint64, status string, result string) error {
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

func (b *MongoDBBackend) _getRecords(page, pageSize int, filter any) ([]agscheduler.Record, int64, error) {
	total := int64(0)

	optsFind := options.Find().SetSort(bson.M{"start_at": -1}).
		SetLimit(int64(pageSize)).SetSkip(int64((page - 1) * pageSize))
	cursor, err := b.coll.Find(ctx, filter, optsFind)
	if err != nil {
		return nil, total, err
	}

	optsCount := options.Count().SetHint("_id_")
	total, err = b.coll.CountDocuments(ctx, filter, optsCount)
	if err != nil {
		return nil, total, err
	}

	recordList := []agscheduler.Record{}
	for cursor.Next(ctx) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			return nil, total, err
		}
		recordList = append(recordList, agscheduler.Record{
			Id:      uint64(result["_id"].(int64)),
			JobId:   result["job_id"].(string),
			JobName: result["job_name"].(string),
			Status:  result["status"].(string),
			Result:  result["result"].(string),
			StartAt: time.Unix(result["start_at"].(int64), 0),
			EndAt:   time.Unix(result["end_at"].(int64), 0),
		})
	}

	return recordList, total, nil
}

func (b *MongoDBBackend) GetRecords(jId string, page, pageSize int) ([]agscheduler.Record, int64, error) {
	return b._getRecords(page, pageSize, bson.M{"job_id": jId})
}

func (b *MongoDBBackend) GetAllRecords(page, pageSize int) ([]agscheduler.Record, int64, error) {
	return b._getRecords(page, pageSize, bson.M{})
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
