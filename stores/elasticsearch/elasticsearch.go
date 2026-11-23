package stores

import (
	"encoding/json"
	"fmt"
	"time"

	es8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/putmapping"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"

	"github.com/agscheduler/agscheduler"
)

const (
	ES_INDEX = "agscheduler_jobs"
)

// Stores jobs in a Elasticsearch database.
type ElasticsearchStore struct {
	TClient *es8.TypedClient
	Index   string
}

type doc struct {
	JobId       string `json:"job_id"`
	NextRunTime int64  `json:"next_run_time"`
	Data        []byte `json:"data"`
}

const esPageSize = 1000

func (s *ElasticsearchStore) Name() string {
	return "Elasticsearch"
}

func (s *ElasticsearchStore) Init() error {
	if s.Index == "" {
		s.Index = ES_INDEX
	}

	exists, err := s.TClient.Indices.Exists(s.Index).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check index exist: %s", err)
	}
	if !exists {
		_, err := s.TClient.Indices.Create(s.Index).Request(
			&create.Request{Mappings: esMappings()},
		).Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to create index: %s", err)
		}
	} else {
		_, err := s.TClient.Indices.PutMapping(s.Index).Request(
			&putmapping.Request{Properties: esMappings().Properties},
		).Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to update index mapping: %s", err)
		}
	}

	return nil
}

func (s *ElasticsearchStore) AddJob(j agscheduler.Job) error {
	bJ, err := agscheduler.JobMarshal(j)
	if err != nil {
		return err
	}

	_, err = s.TClient.Index(s.Index).Id(j.Id).Request(
		doc{
			JobId:       j.Id,
			NextRunTime: j.NextRunTime.UTC().Unix(),
			Data:        bJ,
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

	return agscheduler.JobUnmarshal(d.Data)
}

func (s *ElasticsearchStore) GetAllJobs() ([]agscheduler.Job, error) {
	var jobList []agscheduler.Job
	var searchAfter []types.FieldValue
	for {
		resp, err := s.TClient.Search().Index(s.Index).Request(
			&search.Request{
				Query:       &types.Query{MatchAll: &types.MatchAllQuery{}},
				SearchAfter: searchAfter,
				Sort:        jobSort(),
			},
		).Size(esPageSize).Do(ctx)
		if err != nil {
			return nil, err
		}
		for _, h := range resp.Hits.Hits {
			j, err := jobFromHit(h)
			if err != nil {
				return nil, err
			}
			jobList = append(jobList, j)
		}
		if len(resp.Hits.Hits) < esPageSize {
			break
		}
		searchAfter = resp.Hits.Hits[len(resp.Hits.Hits)-1].Sort
	}

	return jobList, nil
}

func (s *ElasticsearchStore) UpdateJob(j agscheduler.Job) error {
	bJ, err := agscheduler.JobMarshal(j)
	if err != nil {
		return err
	}

	_, err = s.TClient.Update(s.Index, j.Id).Doc(
		doc{
			JobId:       j.Id,
			NextRunTime: j.NextRunTime.UTC().Unix(),
			Data:        bJ,
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

	return time.Unix(d.NextRunTime, 0).UTC(), nil
}

func (s *ElasticsearchStore) Clear() error {
	_, err := s.TClient.Indices.Delete(s.Index).Do(ctx)
	return err
}

func esMappings() *types.TypeMapping {
	return &types.TypeMapping{Properties: map[string]types.Property{
		"job_id":        types.KeywordProperty{Type: "keyword"},
		"next_run_time": types.LongNumberProperty{Type: "long"},
	}}
}

func jobSort() []types.SortCombinations {
	return []types.SortCombinations{
		&types.SortOptions{SortOptions: map[string]types.FieldSort{
			"next_run_time": {Order: ptr(sortorder.Asc)},
		}},
		&types.SortOptions{SortOptions: map[string]types.FieldSort{
			"job_id": {Order: ptr(sortorder.Asc)},
		}},
	}
}

func jobFromHit(hit types.Hit) (agscheduler.Job, error) {
	var d doc
	if err := json.Unmarshal(hit.Source_, &d); err != nil {
		return agscheduler.Job{}, err
	}
	return agscheduler.JobUnmarshal(d.Data)
}

func ptr[T any](value T) *T {
	return &value
}
