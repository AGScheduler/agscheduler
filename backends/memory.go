package backends

import (
	"sort"
	"time"

	"github.com/agscheduler/agscheduler"
)

// Store job records in an array in RAM.
// Provides no persistence support.
// Cluster mode is not supported.
type MemoryBackend struct {
	records []agscheduler.Record
}

func (b *MemoryBackend) Init() error {
	return nil
}

func (b *MemoryBackend) RecordMetadata(r agscheduler.Record) error {
	b.records = append(b.records, r)
	return nil
}

func (b *MemoryBackend) RecordResult(id uint64, status string, result []byte) error {
	for i, r := range b.records {
		if r.Id == id {
			b.records[i].Status = status
			b.records[i].Result = result
			b.records[i].EndAt = time.Now().UTC()
			return nil
		}
	}

	return nil
}

func (b *MemoryBackend) GetRecords(jId string) ([]agscheduler.Record, error) {
	rs := []agscheduler.Record{}
	for _, r := range b.records {
		if r.JobId == jId {
			rs = append(rs, r)
		}
	}
	sort.Sort(agscheduler.RecordSlice(rs))

	return rs, nil
}

func (b *MemoryBackend) GetAllRecords() ([]agscheduler.Record, error) {
	rs := make([]agscheduler.Record, len(b.records))
	copy(rs, b.records)
	sort.Sort(agscheduler.RecordSlice(rs))

	return rs, nil
}

func (b *MemoryBackend) DeleteRecords(jId string) error {
	j := 0
	for _, r := range b.records {
		if r.JobId != jId {
			b.records[j] = r
			j++
		}
	}
	b.records = b.records[:j]

	return nil
}

func (b *MemoryBackend) DeleteAllRecords() error {
	b.records = nil
	return nil
}

func (b *MemoryBackend) Clear() error {
	return b.DeleteAllRecords()
}
