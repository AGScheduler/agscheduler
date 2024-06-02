package services

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/agscheduler/agscheduler"
	pb "github.com/agscheduler/agscheduler/services/proto"
)

type rGRPCService struct {
	pb.UnimplementedRecorderServer

	recorder *agscheduler.Recorder
}

func (rgrs *rGRPCService) _getRecords(jobId string, page int, pageSize int) (*pb.RecordsResp, error) {
	page = fixPositiveNum(page, 1)
	pageSize = fixPositiveNumMax(fixPositiveNum(pageSize, 10), 1000)

	var rs []agscheduler.Record
	var total int64
	var err error
	if jobId != "" {
		rs, total, err = rgrs.recorder.GetRecords(jobId, page, pageSize)
	} else {
		rs, total, err = rgrs.recorder.GetAllRecords(page, pageSize)
	}
	if err != nil {
		return &pb.RecordsResp{}, err
	}

	rsPtr, err := agscheduler.RecordsToPbRecordsPtr(rs)
	if err != nil {
		return &pb.RecordsResp{}, err
	}

	return &pb.RecordsResp{Records: rsPtr, Page: int32(page), PageSize: int32(pageSize), Total: total}, nil
}

func (rgrs *rGRPCService) GetRecords(ctx context.Context, req *pb.RecordsReq) (*pb.RecordsResp, error) {
	return rgrs._getRecords(req.GetJobId(), int(req.GetPage()), int(req.GetPageSize()))
}

func (rgrs *rGRPCService) GetAllRecords(ctx context.Context, req *pb.RecordsAllReq) (*pb.RecordsResp, error) {
	return rgrs._getRecords("", int(req.GetPage()), int(req.GetPageSize()))
}

func (rgrs *rGRPCService) DeleteRecords(ctx context.Context, req *pb.JobReq) (*emptypb.Empty, error) {
	err := rgrs.recorder.DeleteRecords(req.GetId())
	return &emptypb.Empty{}, err
}

func (rgrs *rGRPCService) DeleteAllRecords(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	err := rgrs.recorder.DeleteAllRecords()
	return &emptypb.Empty{}, err
}
