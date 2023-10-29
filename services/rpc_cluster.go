package services

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/rpc"

	"github.com/kwkwc/agscheduler"
)

type CRPCService struct {
	srs *SchedulerRPCService
	cn  *agscheduler.ClusterNode
}

func (crs *CRPCService) Register(args *agscheduler.Node, reply *agscheduler.Node) error {
	crs.cn.RPCRegister(args, reply)
	return nil
}

func (crs *CRPCService) Ping(args *agscheduler.Node, reply *agscheduler.Node) error {
	crs.cn.RPCPing(args, reply)
	return nil
}

type ClusterRPCService struct {
	Srs *SchedulerRPCService
	Cn  *agscheduler.ClusterNode
}

func (s *ClusterRPCService) Start() error {
	if s.Cn.Endpoint == "" {
		s.Cn.Endpoint = "127.0.0.1:36364"
	}
	if s.Cn.MainEndpoint == "" {
		s.Cn.MainEndpoint = s.Cn.Endpoint
	}
	if s.Cn.SchedulerEndpoint == "" {
		s.Cn.SchedulerEndpoint = "127.0.0.1:36363"
	}
	if s.Cn.SchedulerQueue == "" {
		s.Cn.SchedulerQueue = "default"
	}

	s.Srs.Address = s.Cn.SchedulerEndpoint
	s.Srs.Queue = s.Cn.SchedulerQueue
	err := s.Srs.Start()
	if err != nil {
		return err
	}

	crs := &CRPCService{srs: s.Srs, cn: s.Cn}
	rpc.Register(crs)
	rpc.HandleHTTP()

	lis, err := net.Listen("tcp", s.Cn.Endpoint)
	if err != nil {
		return fmt.Errorf("cluster RPC Service listen failure: %s", err)
	}

	go http.Serve(lis, nil)
	slog.Info(fmt.Sprintf("Cluster RPC Service listening at: %s", lis.Addr()))

	return nil
}
