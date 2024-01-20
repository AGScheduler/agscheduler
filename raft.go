package agscheduler

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/rpc"
	"runtime/debug"
	"time"
)

type Role int

const (
	Follower Role = iota + 1
	Candidate
	Leader
)

type Raft struct {
	cn *ClusterNode

	role        Role
	currentTerm int
	votedFor    string
	voteCount   int

	heartbeatC chan bool
	toLeaderC  chan bool
}

func (rf *Raft) toFollower(term int) {
	slog.Info(fmt.Sprintf("Cluster node: `%s`, I'm Follower\n", rf.cn.Endpoint))

	rf.currentTerm = term
	rf.role = Follower
	rf.votedFor = ""

	rf.cn.Scheduler.Stop()
}

type VoteArgs struct {
	Term              int
	CandidateEndpoint string
}

type VoteReply struct {
	Term        int
	VoteGranted bool
}

func (rf *Raft) sendRequestVote(address string, args VoteArgs, reply *VoteReply) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("Address `%s` CRPCService.RaftRequestVote error: %s\n", address, err))
			slog.Debug(fmt.Sprintf("%s\n", string(debug.Stack())))
		}
	}()

	rClient, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to cluster node while sending request vote: `%s`, error: %s\n", address, err))
		return
	}
	defer rClient.Close()

	rClient.Call("CRPCService.RaftRequestVote", args, reply)
	if reply.Term > rf.currentTerm {
		rf.toFollower(reply.Term)
		return
	}

	if reply.VoteGranted {
		rf.voteCount++
		if rf.voteCount > len(rf.cn.HANodeMap())/2+1 {
			rf.toLeaderC <- true
		}
	}
}

func (rf *Raft) broadcastRequestVote() {
	var args = VoteArgs{
		Term:              rf.currentTerm,
		CandidateEndpoint: rf.cn.Endpoint,
	}

	for _, v := range rf.cn.HANodeMap() {
		for endpoint, v2 := range v {
			if rf.cn.Endpoint == endpoint {
				continue
			}
			if !v2["health"].(bool) {
				continue
			}
			go func(address string) {
				var reply VoteReply
				rf.sendRequestVote(address, args, &reply)
			}(endpoint)
		}
	}
}

func (rf *Raft) RPCRequestVote(args VoteArgs, reply *VoteReply) error {
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		return nil
	}

	if args.Term > rf.currentTerm {
		rf.toFollower(args.Term)
	}

	if rf.votedFor == "" {
		rf.votedFor = args.CandidateEndpoint
		reply.Term = rf.currentTerm
		reply.VoteGranted = true
	}

	return nil
}

type HeartbeatArgs struct {
	Term           int
	LeaderEndpoint string

	MainEndpoint string
}

type HeartbeatReply struct {
	Term int
}

func (rf *Raft) sendHeartbeat(address string, args HeartbeatArgs, reply *HeartbeatReply) error {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("Address `%s` CRPCService.RaftHeartbeat error: %s\n", address, err))
			slog.Debug(fmt.Sprintf("%s\n", string(debug.Stack())))
		}
	}()

	rClient, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to cluster node: `%s`, error: %s", address, err)
	}
	defer rClient.Close()

	rClient.Call("CRPCService.RaftHeartbeat", args, reply)
	if reply.Term > rf.currentTerm {
		rf.toFollower(reply.Term)
	}

	return nil
}

func (rf *Raft) broadcastHeartbeat() {
	for _, v := range rf.cn.HANodeMap() {
		for endpoint := range v {
			if rf.cn.Endpoint == endpoint {
				continue
			}
			args := HeartbeatArgs{
				Term:           rf.currentTerm,
				LeaderEndpoint: rf.cn.Endpoint,
			}
			ch := make(chan error, 1)
			go func(address string) {
				var reply HeartbeatReply
				ch <- rf.sendHeartbeat(address, args, &reply)
			}(endpoint)
			select {
			case err := <-ch:
				if err != nil {
					slog.Debug(fmt.Sprintf("Failed to send heartbeat to cluster node: `%s`: %s\n", endpoint, err))
				}
			case <-time.After(50 * time.Millisecond):
				slog.Debug(fmt.Sprintf("Failed to send heartbeat to cluster node: `%s`, timeout\n", endpoint))
			}
		}
	}
}

func (rf *Raft) RPCHeartbeat(args HeartbeatArgs, reply *HeartbeatReply) error {
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		return nil
	}

	if args.Term > rf.currentTerm {
		rf.toFollower(args.Term)
	}

	reply.Term = rf.currentTerm

	rf.cn.SetMainEndpoint(args.LeaderEndpoint)

	rf.heartbeatC <- true

	return nil
}

func (rf *Raft) start(ctx context.Context) {
	rf.role = Follower
	rf.currentTerm = 0
	rf.votedFor = ""
	rf.heartbeatC = make(chan bool)
	rf.toLeaderC = make(chan bool)

	go func(ctx context.Context) {

		rand.New(rand.NewSource(time.Now().UnixNano()))

		for {
			select {
			case <-ctx.Done():
				return
			default:
				switch rf.role {
				case Follower:
					select {
					case <-rf.heartbeatC:
						slog.Debug(fmt.Sprintf("Follower: `%s` recived heartbeat\n", rf.cn.Endpoint))
					case <-time.After(time.Duration(rand.Intn(200)+400) * time.Millisecond):
						slog.Warn(fmt.Sprintf("Follower: `%s` timeout\n", rf.cn.Endpoint))
						rf.role = Candidate
					}
				case Candidate:
					slog.Info(fmt.Sprintf("Cluster node: `%s`, I'm candidate\n", rf.cn.Endpoint))

					rf.cn.Scheduler.Stop()

					rf.currentTerm++
					rf.votedFor = rf.cn.Endpoint
					rf.voteCount = 1
					go rf.broadcastRequestVote()

					select {
					case <-time.After(time.Duration(rand.Intn(5000)+400) * time.Millisecond):
						rf.role = Follower
					case <-rf.toLeaderC:
						slog.Info(fmt.Sprintf("Cluster node: `%s`, I'm leader\n", rf.cn.Endpoint))
						rf.role = Leader

						rf.cn.SetMainEndpoint(rf.cn.Endpoint)
						rf.cn.registerNode(rf.cn)
						rf.cn.Scheduler.Start()
					}
				case Leader:
					rf.broadcastHeartbeat()
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}(ctx)
}