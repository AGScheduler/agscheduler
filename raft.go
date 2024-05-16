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
	slog.Info(fmt.Sprintf("Cluster node: `%s`, I'm Follower", rf.cn.Endpoint))

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

func (rf *Raft) sendRequestVote(address string, args VoteArgs) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("Address `%s` CRPCService.RaftRequestVote error: %s", address, err))
			slog.Debug(string(debug.Stack()))
		}
	}()

	rClient, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to cluster node while sending request vote: `%s`, error: %s", address, err))
		return
	}
	defer rClient.Close()

	var reply VoteReply
	err = rClient.Call("CRPCService.RaftRequestVote", args, &reply)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to send request vote to cluster node: `%s`, error: %s", address, err))
		return
	}

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
	args := VoteArgs{
		Term:              rf.currentTerm,
		CandidateEndpoint: rf.cn.Endpoint,
	}

	for endpoint, v := range rf.cn.HANodeMap() {
		if rf.cn.Endpoint == endpoint {
			continue
		}
		if !v["health"].(bool) {
			continue
		}
		go func(address string) {
			rf.sendRequestVote(address, args)
		}(endpoint)
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

	SchedulerCanStart bool
}

type HeartbeatReply struct {
	Term int
}

func (rf *Raft) sendHeartbeat(address string, args HeartbeatArgs) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(fmt.Sprintf("Address `%s` CRPCService.RaftHeartbeat error: %s", address, err))
			slog.Debug(string(debug.Stack()))
		}
	}()

	rClient, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		slog.Debug(fmt.Sprintf("Failed to connect to cluster node while sending heartbeat: `%s`, error: %s", address, err))
		return
	}
	defer rClient.Close()

	var reply HeartbeatReply
	err = rClient.Call("CRPCService.RaftHeartbeat", args, &reply)
	if err != nil {
		slog.Debug(fmt.Sprintf("Failed to send heartbeat to cluster node: `%s`, error: %s", address, err))
		return
	}

	if reply.Term > rf.currentTerm {
		rf.toFollower(reply.Term)
	}
}

func (rf *Raft) broadcastHeartbeat() {
	args := HeartbeatArgs{
		Term:           rf.currentTerm,
		LeaderEndpoint: rf.cn.Endpoint,

		SchedulerCanStart: rf.cn.Scheduler.IsRunning(),
	}

	for endpoint := range rf.cn.HANodeMap() {
		if rf.cn.Endpoint == endpoint {
			continue
		}
		go func(address string) {
			rf.sendHeartbeat(address, args)
		}(endpoint)
	}
}

func (rf *Raft) RPCHeartbeat(args HeartbeatArgs, reply *HeartbeatReply) error {
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		return nil
	}

	if args.Term > rf.currentTerm || rf.role == Leader {
		rf.toFollower(args.Term)
	}

	reply.Term = rf.currentTerm

	rf.cn.SetEndpointMain(args.LeaderEndpoint)
	rf.cn.SchedulerCanStart = args.SchedulerCanStart

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
						slog.Debug(fmt.Sprintf("Follower: `%s` received heartbeat", rf.cn.Endpoint))
					case <-time.After(time.Duration(rand.Intn(300)+500) * time.Millisecond):
						slog.Warn(fmt.Sprintf("Follower: `%s` timeout", rf.cn.Endpoint))
						rf.role = Candidate
					}
				case Candidate:
					slog.Info(fmt.Sprintf("Cluster node: `%s`, I'm candidate", rf.cn.Endpoint))

					rf.cn.Scheduler.Stop()

					rf.currentTerm++
					rf.votedFor = rf.cn.Endpoint
					rf.voteCount = 1
					go rf.broadcastRequestVote()

					select {
					case <-time.After(time.Duration(rand.Intn(300)+500) * time.Millisecond):
						rf.role = Follower
					case <-rf.toLeaderC:
						slog.Info(fmt.Sprintf("Cluster node: `%s`, I'm leader", rf.cn.Endpoint))
						rf.role = Leader

						rf.cn.SetEndpointMain(rf.cn.Endpoint)
						rf.cn.registerNode(rf.cn)
						if rf.cn.SchedulerCanStart {
							rf.cn.Scheduler.Start()
						}
					}
				case Leader:
					rf.broadcastHeartbeat()
					time.Sleep(300 * time.Millisecond)
				}
			}
		}
	}(ctx)
}
