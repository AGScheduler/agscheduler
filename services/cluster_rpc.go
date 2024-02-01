package services

import (
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"
	"net/http"
	"net/rpc"
	"time"

	"github.com/kwkwc/agscheduler"
)

type CRPCService struct {
	cn *agscheduler.ClusterNode
}

func (crs *CRPCService) GetInfo(filters map[string]any, reply *map[string]any) error {
	*reply = crs.cn.Scheduler.Info()
	return nil
}

func (crs *CRPCService) Register(args *agscheduler.Node, reply *agscheduler.Node) error {
	crs.cn.RPCRegister(args, reply)
	return nil
}

func (crs *CRPCService) Heartbeat(args *agscheduler.Node, reply *agscheduler.Node) error {
	crs.cn.RPCHeartbeat(args, reply)
	return nil
}

func (crs *CRPCService) RunJob(j agscheduler.Job, reply *any) error {
	return crs.cn.Scheduler.RunJob(j)
}

func (crs *CRPCService) RaftRequestVote(args agscheduler.VoteArgs, reply *agscheduler.VoteReply) error {
	var err error
	if crs.cn.Raft != nil {
		err = crs.cn.Raft.RPCRequestVote(args, reply)
	}
	return err
}

func (crs *CRPCService) RaftHeartbeat(args agscheduler.HeartbeatArgs, reply *agscheduler.HeartbeatReply) error {
	var err error
	if crs.cn.Raft != nil {
		err = crs.cn.Raft.RPCHeartbeat(args, reply)
	}
	return err
}

type clusterRPCService struct {
	Cn *agscheduler.ClusterNode

	srv *http.Server
}

func (s *clusterRPCService) Start() error {
	gob.Register(time.Time{})

	crs := &CRPCService{cn: s.Cn}
	rpcServer := rpc.NewServer()
	err := rpcServer.Register(crs)
	if err != nil {
		return fmt.Errorf("failed to register service: %s", err)
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				slog.Warn(fmt.Sprintf("Handle registered error: %s\n", err))
			}
		}()

		rpcServer.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	}()

	s.srv = &http.Server{
		Addr:    s.Cn.Endpoint,
		Handler: rpcServer,
	}

	slog.Info(fmt.Sprintf("Cluster RPC Service listening at: %s", s.Cn.Endpoint))

	s.srv = &http.Server{Addr: s.Cn.Endpoint}

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error(fmt.Sprintf("Cluster RPC Service Unavailable: %s", err))
		}
	}()

	return nil
}

func (s *clusterRPCService) Stop() error {
	slog.Info("Cluster RPC Service stop")

	if err := s.srv.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("failed to stop service: %s", err)
	}

	return nil
}
