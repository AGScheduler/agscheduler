package services

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/kwkwc/agscheduler"
	pb "github.com/kwkwc/agscheduler/services/proto"
)

type cGRPCService struct {
	pb.UnimplementedClusterServer

	cn *agscheduler.ClusterNode
}

func (cgrs *cGRPCService) GetNodes(ctx context.Context, in *emptypb.Empty) (*pb.Nodes, error) {
	return cgrs.cn.NodeMapToPbNodesPtr(), nil
}
