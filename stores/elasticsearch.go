package stores

import (
	"encoding/json"
	"fmt"
	"time"

	es8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"

	"github.com/agscheduler/agscheduler"
)

const (
	INDEX = "agscheduler_jobs"
)

// Stores jobs in a Elasticsearch database.
type ElasticsearchStore struct {
	TClient *es8.TypedClient
	Index   string
}

type doc struct {
	NextRunTime int64  `json:"next_run_time"`
	State       []byte `json:"state"`
}

func (s *ElasticsearchStore) Init() error {
	if s.Index == "" {
		s.Index = INDEX
	}

	exists, err := s.TClient.Indices.Exists(s.Index).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check index exist: %s", err)
	}
	if !exists {
		_, err := s.TClient.Indices.Create(s.Index).Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to create index: %s", err)
		}
	}

	return nil
}

func (s *ElasticsearchStore) AddJob(j agscheduler.Job) error {
	state, err := agscheduler.StateDump(j)
	if err != nil {
		return err
	}

	_, err = s.TClient.Index(s.Index).Id(j.Id).Request(
		doc{
			j.NextRunTime.UTC().Unix(),
			state,
		},
	).Refresh(refresh.True).Do(ctx)

	return err
}

func (s *ElasticsearchStore) GetJob(id string) (agscheduler.Job, error) {
	resp, err := s.TClient.Get(s.Index, id).Do(ctx)
	if err != nil {
		return agscheduler.Job{}, err
	}
	if !resp.Found {
		return agscheduler.Job{}, agscheduler.JobNotFoundError(id)
	}

	var d doc
	err = json.Unmarshal(resp.Source_, &d)
	if err != nil {
		return agscheduler.Job{}, err
	}

	return agscheduler.StateLoad(d.State)
}

func (s *ElasticsearchStore) GetAllJobs() ([]agscheduler.Job, error) {
	resp, err := s.TClient.Search().Index(s.Index).Request(
		&search.Request{
			Query: &types.Query{MatchAll: &types.MatchAllQuery{}},
		},
	).Do(ctx)
	if err != nil {
		return nil, err
	}

	var jobList []agscheduler.Job
	for _, h := range resp.Hits.Hits {
		var d doc
		err = json.Unmarshal(h.Source_, &d)
		if err != nil {
			return nil, err
		}
		aj, err := agscheduler.StateLoad(d.State)
		if err != nil {
			return nil, err
		}
		jobList = append(jobList, aj)
	}

	return jobList, nil
}

func (s *ElasticsearchStore) UpdateJob(j agscheduler.Job) error {
	state, err := agscheduler.StateDump(j)
	if err != nil {
		return err
	}

	_, err = s.TClient.Update(s.Index, j.Id).Doc(
		doc{
			j.NextRunTime.UTC().Unix(),
			state,
		},
	).Refresh(refresh.True).Do(ctx)

	return err
}

func (s *ElasticsearchStore) DeleteJob(id string) error {
	_, err := s.TClient.Delete(s.Index, id).Refresh(refresh.True).Do(ctx)
	return err
}

func (s *ElasticsearchStore) DeleteAllJobs() error {
	_, err := s.TClient.DeleteByQuery(s.Index).Query(
		&types.Query{MatchAll: &types.MatchAllQuery{}},
	).Refresh(true).Do(ctx)

	return err
}

func (s *ElasticsearchStore) GetNextRunTime() (time.Time, error) {
	resp, err := s.TClient.Search().Index(s.Index).Request(
		&search.Request{
			Query: &types.Query{MatchAll: &types.MatchAllQuery{}},
			Sort: []types.SortCombinations{
				&types.SortOptions{
					SortOptions: map[string]types.FieldSort{
						"next_run_time": {Order: &sortorder.Asc},
					},
				},
			},
		},
	).Size(1).Do(ctx)
	if err != nil {
		return time.Time{}, err
	}
	if len(resp.Hits.Hits) == 0 {
		return time.Time{}, nil
	}

	var d doc
	err = json.Unmarshal(resp.Hits.Hits[0].Source_, &d)
	if err != nil {
		return time.Time{}, err
	}

	nextRunTimeMin := time.Unix(d.NextRunTime, 0).UTC()
	return nextRunTimeMin, nil
}

func (s *ElasticsearchStore) Clear() error {
	_, err := s.TClient.Indices.Delete(s.Index).Do(ctx)
	return err
}
