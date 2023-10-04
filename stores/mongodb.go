package stores

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/kwkwc/agscheduler"
)

const (
	database   = "agscheduler"
	collection = "jobs"
)

type MongoDBStore struct {
	Client *mongo.Client
	coll   *mongo.Collection
}

func (s *MongoDBStore) Init() {
	s.coll = s.Client.Database(database).Collection(collection)

	indexModel := mongo.IndexModel{
		Keys: bson.M{
			"next_run_time": 1,
		},
	}
	_, err := s.coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		panic(fmt.Sprintf("Failed to create index: %s\n", err))
	}
}

func (s *MongoDBStore) AddJob(j agscheduler.Job) error {
	state, err := agscheduler.StateDumps(j)
	if err != nil {
		return err
	}

	_, err = s.coll.InsertOne(ctx,
		bson.M{
			"_id":           j.Id,
			"next_run_time": j.NextRunTime.UTC().Unix(),
			"state":         state,
		},
	)

	return err
}

func (s *MongoDBStore) GetJob(id string) (agscheduler.Job, error) {
	var result bson.M
	err := s.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		return agscheduler.Job{}, err
	}
	if err == mongo.ErrNoDocuments {
		return agscheduler.Job{}, agscheduler.JobNotFound(id)
	}

	state := result["state"].(primitive.Binary).Data
	return agscheduler.StateLoads(state)
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
		state := result["state"].(primitive.Binary).Data
		aj, err := agscheduler.StateLoads(state)
		if err != nil {
			return nil, err
		}
		jobList = append(jobList, aj)
	}

	return jobList, nil
}

func (s *MongoDBStore) UpdateJob(j agscheduler.Job) error {
	j.NextRunTime = agscheduler.CalcNextRunTime(j)

	state, err := agscheduler.StateDumps(j)
	if err != nil {
		return err
	}

	var result bson.M
	err = s.coll.FindOneAndReplace(ctx,
		bson.M{"_id": j.Id},
		bson.M{
			"next_run_time": j.NextRunTime.UTC().Unix(),
			"state":         state,
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

	minNextRunTime := time.Unix(result["next_run_time"].(int64), 0)
	return minNextRunTime, nil
}
