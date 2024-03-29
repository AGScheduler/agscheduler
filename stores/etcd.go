package stores

import (
	"path"
	"strconv"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/agscheduler/agscheduler"
)

const (
	JOBS_PATH      = "/agscheduler/jobs"
	RUN_TIMES_PATH = "/agscheduler/run_times"
)

// Stores jobs in a etcd.
type EtcdStore struct {
	Cli          *clientv3.Client
	JobsPath     string
	RunTimesPath string
}

func (s *EtcdStore) Init() error {
	if s.JobsPath == "" {
		s.JobsPath = JOBS_PATH
	}
	if s.RunTimesPath == "" {
		s.RunTimesPath = RUN_TIMES_PATH
	}

	return nil
}

func (s *EtcdStore) AddJob(j agscheduler.Job) error {
	state, err := agscheduler.StateDump(j)
	if err != nil {
		return err
	}

	jPath := path.Join(s.JobsPath, j.Id)
	rPath := path.Join(s.RunTimesPath, j.Id)

	txn := s.Cli.Txn(ctx).If().Then(
		clientv3.OpPut(jPath, string(state)),
		clientv3.OpPut(rPath, strconv.Itoa(int(j.NextRunTime.UTC().Unix()))),
	)
	if _, err := txn.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *EtcdStore) GetJob(id string) (agscheduler.Job, error) {
	jPath := path.Join(s.JobsPath, id)

	resp, err := s.Cli.Get(ctx, jPath)
	if err != nil {
		return agscheduler.Job{}, err
	}
	if len(resp.Kvs) == 0 {
		return agscheduler.Job{}, agscheduler.JobNotFoundError(id)
	}

	state := resp.Kvs[0].Value
	return agscheduler.StateLoad(state)
}

func (s *EtcdStore) GetAllJobs() ([]agscheduler.Job, error) {
	resp, err := s.Cli.Get(ctx, s.JobsPath, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	var jobList []agscheduler.Job
	for _, kv := range resp.Kvs {
		j, err := agscheduler.StateLoad(kv.Value)
		if err != nil {
			return nil, err
		}
		jobList = append(jobList, j)
	}

	return jobList, nil
}

func (s *EtcdStore) UpdateJob(j agscheduler.Job) error {
	state, err := agscheduler.StateDump(j)
	if err != nil {
		return err
	}

	jPath := path.Join(s.JobsPath, j.Id)
	rPath := path.Join(s.RunTimesPath, j.Id)

	txn := s.Cli.Txn(ctx).If(clientv3.Compare(clientv3.Version(jPath), ">", 0)).Then(
		clientv3.OpPut(jPath, string(state)),
		clientv3.OpPut(rPath, strconv.Itoa(int(j.NextRunTime.UTC().Unix()))),
	)
	if _, err := txn.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *EtcdStore) DeleteJob(id string) error {
	jPath := path.Join(s.JobsPath, id)
	rPath := path.Join(s.RunTimesPath, id)

	txn := s.Cli.Txn(ctx).If(clientv3.Compare(clientv3.Version(jPath), ">", 0)).Then(
		clientv3.OpDelete(jPath),
		clientv3.OpDelete(rPath),
	)
	if _, err := txn.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *EtcdStore) DeleteAllJobs() error {
	txn := s.Cli.Txn(ctx).If().Then(
		clientv3.OpDelete(s.JobsPath, clientv3.WithPrefix()),
		clientv3.OpDelete(s.RunTimesPath, clientv3.WithPrefix()),
	)
	if _, err := txn.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *EtcdStore) GetNextRunTime() (time.Time, error) {
	resp, err := s.Cli.Get(ctx, s.RunTimesPath,
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByValue, clientv3.SortAscend),
		clientv3.WithLimit(1),
	)
	if err != nil || len(resp.Kvs) == 0 {
		return time.Time{}, err
	}

	nextRunTimeMinUnix, err := strconv.Atoi(string(resp.Kvs[0].Value))
	if err != nil {
		return time.Time{}, err
	}

	nextRunTimeMin := time.Unix(int64(nextRunTimeMinUnix), 0).UTC()
	return nextRunTimeMin, nil
}

func (s *EtcdStore) Clear() error {
	return s.DeleteAllJobs()
}
