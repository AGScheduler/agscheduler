package services

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/agscheduler/agscheduler"
	pb "github.com/agscheduler/agscheduler/services/proto"
)

type brkGRPCService struct {
	pb.UnimplementedBrokerServer

	broker *agscheduler.Broker
}

func (bgrs *brkGRPCService) GetQueues(ctx context.Context, in *emptypb.Empty) (*pb.QueuesResp, error) {
	qs := bgrs.broker.GetQueues()

	pbQs, err := agscheduler.QueuesToPbQueuesPtr(qs)
	if err != nil {
		return &pb.QueuesResp{}, err
	}

	return &pb.QueuesResp{Queues: pbQs}, err
}
