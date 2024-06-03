package agscheduler

import (
	"sync"
	"time"

	"github.com/sony/sonyflake"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/agscheduler/agscheduler/services/proto"
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
	Id uint64 `json:"id"`
	// Job id
	JobId string `json:"job_id"`
	// Job name
	JobName string `json:"job_name"`
	// Optional: `RECORD_STATUS_RUNNING` | `RECORD_STATUS_COMPLETED` | `RECORD_STATUS_ERROR` | `RECORD_STATUS_TIMEOUT`
	Status string `json:"status"`
	// The result of the job run
	Result string `json:"result"`
	// Start time
	StartAt time.Time `json:"start_at"`
	// End time
	EndAt time.Time `json:"end_at"`
}

// `sort.Interface`, sorted by 'StartAt', descend.
type RecordSlice []Record

func (rs RecordSlice) Len() int           { return len(rs) }
func (rs RecordSlice) Less(i, j int) bool { return rs[i].StartAt.After(rs[j].StartAt) }
func (rs RecordSlice) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

// Used to gRPC Protobuf
func RecordToPbRecordPtr(r Record) (*pb.Record, error) {
	pbR := &pb.Record{
		Id:      r.Id,
		JobId:   r.JobId,
		JobName: r.JobName,
		Status:  r.Status,
		Result:  r.Result,
		StartAt: timestamppb.New(r.StartAt),
		EndAt:   timestamppb.New(r.EndAt),
	}

	return pbR, nil
}

// Used to gRPC Protobuf
func PbRecordPtrToRecord(pbRecord *pb.Record) Record {
	return Record{
		Id:      pbRecord.GetId(),
		JobId:   pbRecord.GetJobId(),
		JobName: pbRecord.GetJobName(),
		Status:  pbRecord.GetStatus(),
		Result:  pbRecord.GetResult(),
		StartAt: pbRecord.GetStartAt().AsTime(),
		EndAt:   pbRecord.GetEndAt().AsTime(),
	}
}

// Used to gRPC Protobuf
func RecordsToPbRecordsPtr(rs []Record) ([]*pb.Record, error) {
	pbRs := []*pb.Record{}

	for _, r := range rs {
		pbR, err := RecordToPbRecordPtr(r)
		if err != nil {
			return []*pb.Record{}, err
		}

		pbRs = append(pbRs, pbR)
	}

	return pbRs, nil
}

// Used to gRPC Protobuf
func PbRecordsPtrToRecords(pbRs []*pb.Record) []Record {
	rs := []Record{}

	for _, pbR := range pbRs {
		rs = append(rs, PbRecordPtrToRecord(pbR))
	}

	return rs
}

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

	t := time.Now().UTC()
	err = r.Backend.RecordMetadata(Record{
		Id:      id,
		JobId:   j.Id,
		JobName: j.Name,
		Status:  RECORD_STATUS_RUNNING,
		StartAt: t,
		EndAt:   t,
	})

	return id, err
}

func (r *Recorder) RecordResult(id uint64, status string, result string) error {
	r.backendM.Lock()
	defer r.backendM.Unlock()

	return r.Backend.RecordResult(id, status, result)
}

func (r *Recorder) GetRecords(jId string, page, pageSize int) ([]Record, int64, error) {
	r.backendM.RLock()
	defer r.backendM.RUnlock()

	return r.Backend.GetRecords(jId, page, pageSize)
}

func (r *Recorder) GetAllRecords(page, pageSize int) ([]Record, int64, error) {
	r.backendM.RLock()
	defer r.backendM.RUnlock()

	return r.Backend.GetAllRecords(page, pageSize)
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
