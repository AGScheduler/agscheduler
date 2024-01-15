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

func (crs *CRPCService) Register(args *agscheduler.Node, reply *agscheduler.Node) error {
	crs.cn.RPCRegister(args, reply)
	return nil
}

func (crs *CRPCService) Ping(args *agscheduler.Node, reply *agscheduler.Node) error {
	crs.cn.RPCPing(args, reply)
	return nil
}

func (crs *CRPCService) Nodes(filters map[string]any, reply *map[string]map[string]map[string]any) error {
	*reply = crs.cn.NodeMap()
	return nil
}

func (crs *CRPCService) RunJob(j agscheduler.Job, reply *any) error {
	return crs.cn.Scheduler.RunJob(j)
}

func (crs *CRPCService) RaftRequestVote(args agscheduler.VoteArgs, reply *agscheduler.VoteReply) error {
	crs.cn.Raft.RPCRequestVote(args, reply)
	return nil
}

func (crs *CRPCService) RaftHeartbeat(args agscheduler.HeartbeatArgs, reply *agscheduler.HeartbeatReply) error {
	crs.cn.Raft.RPCHeartbeat(args, reply)
	return nil
}

type clusterRPCService struct {
	Cn *agscheduler.ClusterNode

	srv *http.Server
}

func (s *clusterRPCService) Start() error {
	gob.Register(time.Time{})

	crs := &CRPCService{cn: s.Cn}
	rpc.Register(crs)
	rpc.HandleHTTP()

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
	if err := s.srv.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("failed to stop service: %s", err)
	}

	return nil
}
