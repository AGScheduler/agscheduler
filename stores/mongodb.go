package stores

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
	MONGODB_COLLECTION = "jobs"
)

// Stores jobs in a MongoDB database.
type MongoDBStore struct {
	Client     *mongo.Client
	Database   string
	Collection string
	coll       *mongo.Collection
}

func (s *MongoDBStore) Init() error {
	if s.Database == "" {
		s.Database = MONGODB_DATABASE
	}
	if s.Collection == "" {
		s.Collection = MONGODB_COLLECTION
	}

	s.coll = s.Client.Database(s.Database).Collection(s.Collection)

	indexModel := mongo.IndexModel{
		Keys: bson.M{
			"next_run_time": 1,
		},
	}
	_, err := s.coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create index: %s", err)
	}

	return nil
}

func (s *MongoDBStore) AddJob(j agscheduler.Job) error {
	bJ, err := agscheduler.JobMarshal(j)
	if err != nil {
		return err
	}

	_, err = s.coll.InsertOne(ctx,
		bson.M{
			"_id":           j.Id,
			"next_run_time": j.NextRunTime.UTC().Unix(),
			"data":          bJ,
		},
	)

	return err
}

func (s *MongoDBStore) GetJob(id string) (agscheduler.Job, error) {
	var result bson.M
	err := s.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return agscheduler.Job{}, agscheduler.JobNotFoundError(id)
	}
	if err != nil {
		return agscheduler.Job{}, err
	}

	bJ := result["data"].(primitive.Binary).Data
	return agscheduler.JobUnmarshal(bJ)
}

func (s *MongoDBStore) GetAllJobs() ([]agscheduler.Job, error) {
	cursor, err := s.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var jobList []agscheduler.Job
	for cursor.Next(ctx) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
		bJ := result["data"].(primitive.Binary).Data
		aj, err := agscheduler.JobUnmarshal(bJ)
		if err != nil {
			return nil, err
		}
		jobList = append(jobList, aj)
	}

	return jobList, nil
}

func (s *MongoDBStore) UpdateJob(j agscheduler.Job) error {
	bJ, err := agscheduler.JobMarshal(j)
	if err != nil {
		return err
	}

	var result bson.M
	err = s.coll.FindOneAndReplace(ctx,
		bson.M{"_id": j.Id},
		bson.M{
			"next_run_time": j.NextRunTime.UTC().Unix(),
			"data":          bJ,
		},
	).Decode(&result)

	return err
}

func (s *MongoDBStore) DeleteJob(id string) error {
	_, err := s.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (s *MongoDBStore) DeleteAllJobs() error {
	_, err := s.coll.DeleteMany(ctx, bson.M{})
	return err
}

func (s *MongoDBStore) GetNextRunTime() (time.Time, error) {
	var result bson.M
	opts := options.FindOne().SetSort(bson.M{"next_run_time": 1})
	err := s.coll.FindOne(ctx, bson.M{}, opts).Decode(&result)
	if err != nil {
		return time.Time{}, err
	}
	if err == mongo.ErrNoDocuments {
		return time.Time{}, nil
	}

	nextRunTimeMin := time.Unix(result["next_run_time"].(int64), 0).UTC()
	return nextRunTimeMin, nil
}

func (s *MongoDBStore) Clear() error {
	return s.Client.Database(s.Database).Collection(s.Collection).Drop(ctx)
}
