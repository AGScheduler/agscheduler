package agscheduler_test

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/agscheduler/agscheduler"
)

func TestRecordSort(t *testing.T) {
	rs := []agscheduler.Record{}

	r1 := agscheduler.Record{}
	r1.StartAt = time.Now().UTC()
	rs = append(rs, r1)

	r2 := agscheduler.Record{}
	r2.StartAt = time.Now().UTC()
	rs = append(rs, r2)

	assert.Equal(t, r1.StartAt, rs[0].StartAt)

	sort.Sort(agscheduler.RecordSlice(rs))

	assert.Equal(t, r2.StartAt, rs[0].StartAt)
}

func TestRecorderRecordMetadata(t *testing.T) {
	j := agscheduler.Job{Id: "1"}
	rec := getRecorder()
	s := &agscheduler.Scheduler{}
	err := s.SetRecorder(rec)
	assert.NoError(t, err)

	_, err = rec.RecordMetadata(j)
	assert.NoError(t, err)

	rs, _, err := rec.GetRecords(j.Id, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, j.Id, rs[0].JobId)
	assert.Equal(t, agscheduler.RECORD_STATUS_RUNNING, rs[0].Status)
}

func TestRecorderRecordResult(t *testing.T) {
	j := agscheduler.Job{Id: "1"}
	rec := getRecorder()
	s := &agscheduler.Scheduler{}
	err := s.SetRecorder(rec)
	assert.NoError(t, err)

	id, err := rec.RecordMetadata(j)
	assert.NoError(t, err)

	err = rec.RecordResult(id, agscheduler.RECORD_STATUS_COMPLETED, "")
	assert.NoError(t, err)

	rs, _, err := rec.GetRecords(j.Id, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.RECORD_STATUS_COMPLETED, rs[0].Status)
}

func TestRecorderGetRecords(t *testing.T) {
	j := agscheduler.Job{Id: "1"}
	rec := getRecorder()
	s := &agscheduler.Scheduler{}
	err := s.SetRecorder(rec)
	assert.NoError(t, err)

	rs, total, err := rec.GetRecords(j.Id, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, rs, 0)
	assert.Equal(t, 0, int(total))

	_, err = rec.RecordMetadata(j)
	assert.NoError(t, err)

	rs, total, err = rec.GetRecords(j.Id, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, rs, 1)
	assert.Equal(t, 1, int(total))
}

func TestRecorderGetAllRecords(t *testing.T) {
	j := agscheduler.Job{Id: "1"}
	j2 := agscheduler.Job{Id: "2"}
	rec := getRecorder()
	s := &agscheduler.Scheduler{}
	err := s.SetRecorder(rec)
	assert.NoError(t, err)

	rs, total, err := rec.GetRecords(j.Id, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, rs, 0)
	assert.Equal(t, 0, int(total))

	_, err = rec.RecordMetadata(j)
	assert.NoError(t, err)
	_, err = rec.RecordMetadata(j)
	assert.NoError(t, err)
	_, err = rec.RecordMetadata(j2)
	assert.NoError(t, err)

	rs, total, err = rec.GetRecords(j.Id, 1, 1)
	assert.NoError(t, err)
	assert.Len(t, rs, 1)
	assert.Equal(t, 2, int(total))
	rs, total, err = rec.GetAllRecords(2, 2)
	assert.NoError(t, err)
	assert.Len(t, rs, 1)
	assert.Equal(t, 3, int(total))
	rs, total, err = rec.GetAllRecords(1, 10)
	assert.NoError(t, err)
	assert.Len(t, rs, 3)
	assert.Equal(t, 3, int(total))
	rs, _, err = rec.GetAllRecords(10, 10)
	assert.NoError(t, err)
}

func TestRecorderDeleteRecords(t *testing.T) {
	j := agscheduler.Job{Id: "1"}
	rec := getRecorder()
	s := &agscheduler.Scheduler{}
	err := s.SetRecorder(rec)
	assert.NoError(t, err)

	_, err = rec.RecordMetadata(j)
	assert.NoError(t, err)
	rs, _, err := rec.GetRecords(j.Id, 1, 10)
	assert.Len(t, rs, 1)

	err = rec.DeleteRecords(j.Id)
	assert.NoError(t, err)
	rs, _, err = rec.GetRecords(j.Id, 1, 10)
	assert.Len(t, rs, 0)
}

func TestRecorderDeleteAllRecords(t *testing.T) {
	j := agscheduler.Job{Id: "1"}
	rec := getRecorder()
	s := &agscheduler.Scheduler{}
	err := s.SetRecorder(rec)
	assert.NoError(t, err)

	_, err = rec.RecordMetadata(j)
	assert.NoError(t, err)
	rs, _, err := rec.GetRecords(j.Id, 1, 10)
	assert.Len(t, rs, 1)

	err = rec.DeleteAllRecords()
	assert.NoError(t, err)
	rs, _, err = rec.GetRecords(j.Id, 1, 10)
	assert.Len(t, rs, 0)
}

func TestRecorderClear(t *testing.T) {
	j := agscheduler.Job{Id: "1"}
	rec := getRecorder()
	s := &agscheduler.Scheduler{}
	err := s.SetRecorder(rec)
	assert.NoError(t, err)

	_, err = rec.RecordMetadata(j)
	assert.NoError(t, err)

	rs, _, err := rec.GetRecords(j.Id, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, rs, 1)

	err = rec.Clear()
	assert.NoError(t, err)

	rs, _, err = rec.GetRecords(j.Id, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, rs, 0)
}
