package agscheduler

import (
	"sync"
	"time"

	"github.com/sony/sonyflake"
)

// constant indicating the status of the job record
const (
	RECORD_STATUS_RUNNING   = "running"
	RECORD_STATUS_COMPLETED = "completed"
	RECORD_STATUS_ERROR     = "error"
	RECORD_STATUS_TIMEOUT   = "timeout"
)

// Carry the information of the job run.
type Record struct {
	// Unique Id
	Id uint64
	// Job id
	JobId string
	// Job name
	JobName string
	// Optional: `RECORD_STATUS_RUNNING` | `RECORD_STATUS_COMPLETED` | `RECORD_STATUS_ERROR` | `RECORD_STATUS_TIMEOUT`
	Status string
	// The result of the job run
	Result []byte
	// Start time
	StartAt time.Time
	// End time
	EndAt time.Time
}

// `sort.Interface`, sorted by 'StartAt', descend.
type RecordSlice []Record

func (rs RecordSlice) Len() int           { return len(rs) }
func (rs RecordSlice) Less(i, j int) bool { return rs[i].StartAt.After(rs[j].StartAt) }
func (rs RecordSlice) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

// When using a Recorder, the results of the job runs will be recorded to the specified backend.
type Recorder struct {
	// Record store
	// It should not be used directly.
	Backend Backend

	// Distributed unique Id generator
	// It should not be set manually.
	sf *sonyflake.Sonyflake

	backendM sync.RWMutex
}

// Initialization functions for each Recorder,
// called when the scheduler run `SetRecorder`.
func (r *Recorder) init() error {
	sf, err := sonyflake.New(sonyflake.Settings{
		StartTime:      time.Date(2014, 9, 1, 0, 0, 0, 0, time.UTC),
		MachineID:      nil,
		CheckMachineID: nil,
	})
	if err != nil {
		return err
	}
	r.sf = sf

	if err := r.Backend.Init(); err != nil {
		return err
	}

	return nil
}

func (r *Recorder) RecordMetadata(j Job) (id uint64, err error) {
	r.backendM.Lock()
	defer r.backendM.Unlock()

	id, err = r.sf.NextID()
	if err != nil {
		return id, err
	}

	err = r.Backend.RecordMetadata(Record{
		Id:      id,
		JobId:   j.Id,
		JobName: j.Name,
		Status:  RECORD_STATUS_RUNNING,
		StartAt: time.Now().UTC(),
	})

	return id, err
}

func (r *Recorder) RecordResult(id uint64, status string, result []byte) error {
	r.backendM.Lock()
	defer r.backendM.Unlock()

	return r.Backend.RecordResult(id, status, result)
}

func (r *Recorder) GetRecords(jId string) ([]Record, error) {
	r.backendM.RLock()
	defer r.backendM.RUnlock()

	return r.Backend.GetRecords(jId)
}

func (r *Recorder) GetAllRecords() ([]Record, error) {
	r.backendM.RLock()
	defer r.backendM.RUnlock()

	return r.Backend.GetAllRecords()
}

func (r *Recorder) DeleteRecords(jId string) error {
	r.backendM.Lock()
	defer r.backendM.Unlock()

	return r.Backend.DeleteRecords(jId)
}

func (r *Recorder) DeleteAllRecords() error {
	r.backendM.Lock()
	defer r.backendM.Unlock()

	return r.Backend.DeleteAllRecords()
}

func (r *Recorder) Clear() error {
	r.backendM.Lock()
	defer r.backendM.Unlock()

	return r.Backend.Clear()
}
