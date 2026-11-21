package services

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/agscheduler/agscheduler"
	"github.com/agscheduler/agscheduler/services/grpc/convert"
	pb "github.com/agscheduler/agscheduler/services/proto"
)

type cGRPCService struct {
	pb.UnimplementedClusterServer

	cn *agscheduler.ClusterNode
}

func (cgrs *cGRPCService) GetNodes(ctx context.Context, in *emptypb.Empty) (*pb.NodesResp, error) {
	pbNs := convert.NodeMapToPbNodesPtr(cgrs.cn)
	return &pb.NodesResp{Nodes: pbNs}, nil
}
