package stores

import (
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kwkwc/agscheduler"
)

const (
	JOBS_KEY      = "agscheduler.jobs"
	RUN_TIMES_KEY = "agscheduler.run_times"
)

// Stores jobs in a Redis database.
type RedisStore struct {
	RDB         *redis.Client
	JobsKey     string
	RunTimesKey string
}

func (s *RedisStore) Init() error {
	if s.JobsKey == "" {
		s.JobsKey = JOBS_KEY
	}
	if s.RunTimesKey == "" {
		s.RunTimesKey = RUN_TIMES_KEY
	}

	return nil
}

func (s *RedisStore) AddJob(j agscheduler.Job) error {
	state, err := agscheduler.StateDump(j)
	if err != nil {
		return err
	}

	_, err = s.RDB.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, s.JobsKey, j.Id, state)
		pipe.ZAdd(ctx, s.RunTimesKey, redis.Z{Score: float64(j.NextRunTime.UTC().Unix()), Member: j.Id})
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *RedisStore) GetJob(id string) (agscheduler.Job, error) {
	state, err := s.RDB.HGet(ctx, s.JobsKey, id).Bytes()
	if err == redis.Nil {
		return agscheduler.Job{}, agscheduler.JobNotFoundError(id)
	}
	if err != nil {
		return agscheduler.Job{}, err
	}

	return agscheduler.StateLoad(state)
}

func (s *RedisStore) GetAllJobs() ([]agscheduler.Job, error) {
	mapStates, err := s.RDB.HGetAll(ctx, s.JobsKey).Result()
	if err != nil {
		return nil, err
	}

	var jobList []agscheduler.Job
	for _, v := range mapStates {
		j, err := agscheduler.StateLoad([]byte(v))
		if err != nil {
			return nil, err
		}
		jobList = append(jobList, j)
	}

	return jobList, nil
}

func (s *RedisStore) UpdateJob(j agscheduler.Job) error {
	state, err := agscheduler.StateDump(j)
	if err != nil {
		return err
	}

	_, err = s.RDB.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, s.JobsKey, j.Id, state)
		pipe.ZAdd(ctx, s.RunTimesKey, redis.Z{Score: float64(j.NextRunTime.UTC().Unix()), Member: j.Id})
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *RedisStore) DeleteJob(id string) error {
	_, err := s.RDB.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HDel(ctx, s.JobsKey, id)
		pipe.ZRem(ctx, s.RunTimesKey, id)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *RedisStore) DeleteAllJobs() error {
	_, err := s.RDB.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, s.JobsKey)
		pipe.Del(ctx, s.RunTimesKey)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *RedisStore) GetNextRunTime() (time.Time, error) {
	sliceRunTimes, err := s.RDB.ZRangeWithScores(ctx, s.RunTimesKey, 0, 0).Result()
	if err != nil || len(sliceRunTimes) == 0 {
		return time.Time{}, nil
	}

	nextRunTimeMin := time.Unix(int64(sliceRunTimes[0].Score), 0).UTC()
	return nextRunTimeMin, nil
}

func (s *RedisStore) Clear() error {
	return s.DeleteAllJobs()
}
