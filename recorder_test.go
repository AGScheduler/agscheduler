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

	rs, err := rec.GetRecords(j.Id)
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

	rs, err := rec.GetRecords(j.Id)
	assert.NoError(t, err)
	assert.Equal(t, agscheduler.RECORD_STATUS_COMPLETED, rs[0].Status)
}

func TestRecorderGetRecords(t *testing.T) {
	j := agscheduler.Job{Id: "1"}
	rec := getRecorder()
	s := &agscheduler.Scheduler{}
	err := s.SetRecorder(rec)
	assert.NoError(t, err)

	rs, err := rec.GetRecords(j.Id)
	assert.Len(t, rs, 0)

	_, err = rec.RecordMetadata(j)
	assert.NoError(t, err)

	rs, err = rec.GetRecords(j.Id)
	assert.Len(t, rs, 1)
}

func TestRecorderGetAllRecords(t *testing.T) {
	j := agscheduler.Job{Id: "1"}
	j2 := agscheduler.Job{Id: "2"}
	rec := getRecorder()
	s := &agscheduler.Scheduler{}
	err := s.SetRecorder(rec)
	assert.NoError(t, err)

	rs, err := rec.GetRecords(j.Id)
	assert.Len(t, rs, 0)

	_, err = rec.RecordMetadata(j)
	assert.NoError(t, err)
	_, err = rec.RecordMetadata(j2)
	assert.NoError(t, err)

	rs, err = rec.GetRecords(j.Id)
	assert.Len(t, rs, 1)
	rs, err = rec.GetAllRecords()
	assert.Len(t, rs, 2)
}

func TestRecorderDeleteRecords(t *testing.T) {
	j := agscheduler.Job{Id: "1"}
	rec := getRecorder()
	s := &agscheduler.Scheduler{}
	err := s.SetRecorder(rec)
	assert.NoError(t, err)

	_, err = rec.RecordMetadata(j)
	assert.NoError(t, err)
	rs, err := rec.GetRecords(j.Id)
	assert.Len(t, rs, 1)

	err = rec.DeleteRecords(j.Id)
	assert.NoError(t, err)
	rs, err = rec.GetRecords(j.Id)
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
	rs, err := rec.GetRecords(j.Id)
	assert.Len(t, rs, 1)

	err = rec.DeleteAllRecords()
	assert.NoError(t, err)
	rs, err = rec.GetRecords(j.Id)
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

	rs, err := rec.GetRecords(j.Id)
	assert.Len(t, rs, 1)

	err = rec.Clear()
	assert.NoError(t, err)

	rs, err = rec.GetRecords(j.Id)
	assert.Len(t, rs, 0)
}
