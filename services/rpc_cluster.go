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
	cm  *agscheduler.ClusterMain
	cw  *agscheduler.ClusterWorker
}

func (crs *CRPCService) Register(w *agscheduler.Worker, m *agscheduler.Main) error {
	err := crs.cm.Register(w, m)
	if err != nil {
		return err
	}

	return nil
}

type ClusterRPCService struct {
	Srs *SchedulerRPCService
	Cm  *agscheduler.ClusterMain
	Cw  *agscheduler.ClusterWorker
}

func (s *ClusterRPCService) Start() error {
	address := "127.0.0.1:36364"
	if s.Cm != nil && s.Cm.Endpoint != "" {
		address = s.Cm.Endpoint
	}
	if s.Cw != nil && s.Cw.Endpoint != "" {
		address = s.Cw.Endpoint
	}

	crs := &CRPCService{srs: s.Srs, cm: s.Cm, cw: s.Cw}
	rpc.Register(crs)
	rpc.HandleHTTP()

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("cluster RPC Service listen failure: %s", err)
	}

	err = s.Srs.Start()
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Cluster RPC Service listening at: %s", lis.Addr()))
	go http.Serve(lis, nil)

	return nil
}
