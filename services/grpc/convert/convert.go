// Package convert translates the core scheduler types to and from the public
// protobuf transport types.
package convert

import (
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/agscheduler/agscheduler"
	pb "github.com/agscheduler/agscheduler/services/proto"
)

func JobToPbJobPtr(j agscheduler.Job) (*pb.Job, error) {
	args, err := structpb.NewStruct(j.Args)
	if err != nil {
		return &pb.Job{}, err
	}

	return &pb.Job{
		Id:           j.Id,
		Name:         j.Name,
		Type:         j.Type,
		StartAt:      j.StartAt,
		EndAt:        j.EndAt,
		Interval:     j.Interval,
		CronExpr:     j.CronExpr,
		Timezone:     j.Timezone,
		FuncName:     j.FuncName,
		Args:         args,
		Timeout:      j.Timeout,
		Queues:       j.Queues,
		MaxInstances: int32(j.MaxInstances),
		LastRunTime:  timestamppb.New(j.LastRunTime),
		NextRunTime:  timestamppb.New(j.NextRunTime),
		Status:       j.Status,
	}, nil
}

func PbJobPtrToJob(pbJob *pb.Job) agscheduler.Job {
	return agscheduler.Job{
		Id:           pbJob.GetId(),
		Name:         pbJob.GetName(),
		Type:         pbJob.GetType(),
		StartAt:      pbJob.GetStartAt(),
		EndAt:        pbJob.GetEndAt(),
		Interval:     pbJob.GetInterval(),
		CronExpr:     pbJob.GetCronExpr(),
		Timezone:     pbJob.GetTimezone(),
		FuncName:     pbJob.GetFuncName(),
		Args:         pbJob.GetArgs().AsMap(),
		Timeout:      pbJob.GetTimeout(),
		Queues:       pbJob.GetQueues(),
		MaxInstances: max(1, int(pbJob.GetMaxInstances())),
		LastRunTime:  pbJob.GetLastRunTime().AsTime(),
		NextRunTime:  pbJob.GetNextRunTime().AsTime(),
		Status:       pbJob.GetStatus(),
	}
}

func JobsToPbJobsPtr(jobs []agscheduler.Job) ([]*pb.Job, error) {
	pbJobs := make([]*pb.Job, 0, len(jobs))
	for _, job := range jobs {
		pbJob, err := JobToPbJobPtr(job)
		if err != nil {
			return nil, err
		}
		pbJobs = append(pbJobs, pbJob)
	}
	return pbJobs, nil
}

func PbJobsPtrToJobs(pbJobs []*pb.Job) []agscheduler.Job {
	jobs := make([]agscheduler.Job, 0, len(pbJobs))
	for _, pbJob := range pbJobs {
		jobs = append(jobs, PbJobPtrToJob(pbJob))
	}
	return jobs
}

func RecordToPbRecordPtr(record agscheduler.Record) (*pb.Record, error) {
	return &pb.Record{
		Id:      record.Id,
		JobId:   record.JobId,
		JobName: record.JobName,
		Status:  record.Status,
		Result:  record.Result,
		StartAt: timestamppb.New(record.StartAt),
		EndAt:   timestamppb.New(record.EndAt),
	}, nil
}

func PbRecordPtrToRecord(pbRecord *pb.Record) agscheduler.Record {
	return agscheduler.Record{
		Id:      pbRecord.GetId(),
		JobId:   pbRecord.GetJobId(),
		JobName: pbRecord.GetJobName(),
		Status:  pbRecord.GetStatus(),
		Result:  pbRecord.GetResult(),
		StartAt: pbRecord.GetStartAt().AsTime(),
		EndAt:   pbRecord.GetEndAt().AsTime(),
	}
}

func RecordsToPbRecordsPtr(records []agscheduler.Record) ([]*pb.Record, error) {
	pbRecords := make([]*pb.Record, 0, len(records))
	for _, record := range records {
		pbRecord, err := RecordToPbRecordPtr(record)
		if err != nil {
			return nil, err
		}
		pbRecords = append(pbRecords, pbRecord)
	}
	return pbRecords, nil
}

func PbRecordsPtrToRecords(pbRecords []*pb.Record) []agscheduler.Record {
	records := make([]agscheduler.Record, 0, len(pbRecords))
	for _, pbRecord := range pbRecords {
		records = append(records, PbRecordPtrToRecord(pbRecord))
	}
	return records
}

func QueueToPbQueuePtr(queue map[string]any) (*pb.Queue, error) {
	return &pb.Queue{
		Name:    queue["name"].(string),
		Type:    queue["type"].(string),
		Count:   int64(queue["count"].(int)),
		Workers: int32(queue["workers"].(int)),
	}, nil
}

func QueuesToPbQueuesPtr(queues []map[string]any) ([]*pb.Queue, error) {
	pbQueues := make([]*pb.Queue, 0, len(queues))
	for _, queue := range queues {
		pbQueue, err := QueueToPbQueuePtr(queue)
		if err != nil {
			return nil, err
		}
		pbQueues = append(pbQueues, pbQueue)
	}
	return pbQueues, nil
}

func NodeMapToPbNodesPtr(cn *agscheduler.ClusterNode) map[string]*pb.Node {
	pbNodes := make(map[string]*pb.Node)
	for endpoint, node := range cn.NodeMapCopy() {
		pbNodes[endpoint] = &pb.Node{
			EndpointMain:      node["endpoint_main"].(string),
			Endpoint:          node["endpoint"].(string),
			EndpointGrpc:      node["endpoint_grpc"].(string),
			EndpointHttp:      node["endpoint_http"].(string),
			Queue:             node["queue"].(string),
			Mode:              node["mode"].(string),
			Version:           agscheduler.Version,
			Health:            node["health"].(bool),
			RegisterTime:      timestamppb.New(node["register_time"].(time.Time)),
			LastHeartbeatTime: timestamppb.New(node["last_heartbeat_time"].(time.Time)),
		}
	}
	return pbNodes
}
