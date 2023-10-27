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
	err := crs.cn.RPCRegister(args, reply)
	if err != nil {
		return err
	}

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

	crs := &CRPCService{srs: s.Srs, cn: s.Cn}
	rpc.Register(crs)
	rpc.HandleHTTP()

	lis, err := net.Listen("tcp", s.Cn.Endpoint)
	if err != nil {
		return fmt.Errorf("cluster RPC Service listen failure: %s", err)
	}

	s.Srs.Address = s.Cn.SchedulerEndpoint
	s.Srs.Queue = s.Cn.SchedulerQueue
	err = s.Srs.Start()
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Cluster RPC Service listening at: %s", lis.Addr()))
	go http.Serve(lis, nil)

	return nil
}
