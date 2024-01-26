package services

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kwkwc/agscheduler"
	pb "github.com/kwkwc/agscheduler/services/proto"
)

type cGRPCService struct {
	pb.UnimplementedClusterServer

	cn *agscheduler.ClusterNode
}

func (cgrs *cGRPCService) GetNodes(ctx context.Context, in *emptypb.Empty) (*pb.Nodes, error) {
	pbN := pb.Nodes{}
	pbN.Nodes = make(map[string]*pb.Node)

	for k, v := range cgrs.cn.NodeMapCopy() {
		pbN.Nodes[k] = &pb.Node{
			MainEndpoint:      v["main_endpoint"].(string),
			Endpoint:          v["endpoint"].(string),
			EndpointGrpc:      v["endpoint_grpc"].(string),
			EndpointHttp:      v["endpoint_http"].(string),
			Queue:             v["queue"].(string),
			Mode:              v["mode"].(string),
			Health:            v["health"].(bool),
			RegisterTime:      timestamppb.New(v["register_time"].(time.Time)),
			LastHeartbeatTime: timestamppb.New(v["last_heartbeat_time"].(time.Time)),
		}
	}

	return &pbN, nil
}
