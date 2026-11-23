package agscheduler

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/rpc"
	"runtime/debug"
	"sync"
	"time"
)

const (
	raftHeartbeatInterval = 300 * time.Millisecond
	raftElectionBase      = 500 * time.Millisecond
	raftElectionJitter    = 300 * time.Millisecond
)

type Role int

const (
	RAFT_FOLLOWER Role = iota + 1
	RAFT_CANDIDATE
	RAFT_LEADER
)

type Raft struct {
	cn *ClusterNode

	mu          sync.Mutex
	role        Role
	currentTerm int
	votedFor    string
	votes       map[string]struct{}

	heartbeatC chan struct{}
	leaderC    chan int
}

type VoteArgs struct {
	Term              int
	CandidateEndpoint string
}

type VoteReply struct {
	Term        int
	VoteGranted bool
}

type HeartbeatArgs struct {
	Term           int
	LeaderEndpoint string

	SchedulerCanStart bool
}

type HeartbeatReply struct {
	Term int
}

func raftElectionTimeout() time.Duration {
	return raftElectionBase + time.Duration(rand.Intn(int(raftElectionJitter/time.Millisecond)))*time.Millisecond
}

func (rf *Raft) init() {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	rf.role = RAFT_FOLLOWER
	rf.currentTerm = 0
	rf.votedFor = ""
	rf.votes = nil
	rf.heartbeatC = make(chan struct{}, 1)
	rf.leaderC = make(chan int, 1)
}

func (rf *Raft) state() (Role, int) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	return rf.role, rf.currentTerm
}

func (rf *Raft) quorumSize() int {
	return len(rf.cn.HANodeMap())/2 + 1
}

func (rf *Raft) notifyHeartbeat() {
	select {
	case rf.heartbeatC <- struct{}{}:
	default:
	}
}

func (rf *Raft) notifyLeader(term int) {
	select {
	case rf.leaderC <- term:
	default:
	}
}

// becomeFollower advances the term when necessary and reports whether the node
// was a leader before the transition, in which case its scheduler must stop.
func (rf *Raft) becomeFollower(term int) (stoppedLeading bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if term < rf.currentTerm {
		return false
	}
	if term > rf.currentTerm {
		rf.currentTerm = term
		rf.votedFor = ""
		rf.votes = nil
	}

	stoppedLeading = rf.role == RAFT_LEADER
	rf.role = RAFT_FOLLOWER
	return stoppedLeading
}

func (rf *Raft) stopSchedulerAfterStepDown() {
	if rf.cn.Scheduler != nil {
		go rf.cn.Scheduler.Stop()
	}
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
	defer func() {
		_ = rClient.Close()
	}()

	var reply VoteReply
	err = rClient.Call("CRPCService.RaftRequestVote", args, &reply)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to send request vote to cluster node: `%s`, error: %s", address, err))
		return
	}

	if reply.Term > args.Term {
		if rf.becomeFollower(reply.Term) {
			rf.stopSchedulerAfterStepDown()
		}
		rf.notifyHeartbeat()
		return
	}
	if reply.VoteGranted {
		rf.recordVote(args.Term, address)
	}
}

func (rf *Raft) recordVote(term int, address string) {
	rf.mu.Lock()
	if rf.role != RAFT_CANDIDATE || rf.currentTerm != term {
		rf.mu.Unlock()
		return
	}
	if _, ok := rf.votes[address]; ok {
		rf.mu.Unlock()
		return
	}

	rf.votes[address] = struct{}{}
	promoted := len(rf.votes) >= rf.quorumSize()
	if promoted {
		rf.role = RAFT_LEADER
	}
	rf.mu.Unlock()

	if promoted {
		rf.notifyLeader(term)
	}
}

func (rf *Raft) broadcastRequestVote(term int) {
	args := VoteArgs{
		Term:              term,
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

func (rf *Raft) beginElection() int {
	select {
	case <-rf.leaderC:
	default:
	}

	rf.mu.Lock()
	rf.role = RAFT_CANDIDATE
	rf.currentTerm++
	rf.votedFor = rf.cn.Endpoint
	rf.votes = map[string]struct{}{rf.cn.Endpoint: {}}
	term := rf.currentTerm
	promoted := len(rf.votes) >= rf.quorumSize()
	if promoted {
		rf.role = RAFT_LEADER
	}
	rf.mu.Unlock()

	if promoted {
		rf.notifyLeader(term)
		return term
	}

	if rf.cn.Scheduler != nil {
		go rf.cn.Scheduler.Stop()
	}
	rf.broadcastRequestVote(term)
	return term
}

func (rf *Raft) RPCRequestVote(args VoteArgs, reply *VoteReply) error {
	rf.mu.Lock()
	oldRole := rf.role
	oldTerm := rf.currentTerm
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		rf.mu.Unlock()
		return nil
	}
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.votedFor = ""
		rf.votes = nil
		rf.role = RAFT_FOLLOWER
	}

	reply.Term = rf.currentTerm
	reply.VoteGranted = rf.currentTerm == args.Term &&
		rf.role != RAFT_LEADER &&
		(rf.votedFor == "" || rf.votedFor == args.CandidateEndpoint)
	if reply.VoteGranted {
		rf.votedFor = args.CandidateEndpoint
	}
	rf.mu.Unlock()

	if oldRole != RAFT_FOLLOWER && args.Term > oldTerm {
		if oldRole == RAFT_LEADER {
			rf.stopSchedulerAfterStepDown()
		}
		rf.notifyHeartbeat()
	}

	return nil
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
	defer func() {
		_ = rClient.Close()
	}()

	var reply HeartbeatReply
	err = rClient.Call("CRPCService.RaftHeartbeat", args, &reply)
	if err != nil {
		slog.Debug(fmt.Sprintf("Failed to send heartbeat to cluster node: `%s`, error: %s", address, err))
		return
	}

	if reply.Term > args.Term {
		if rf.becomeFollower(reply.Term) {
			rf.stopSchedulerAfterStepDown()
		}
		rf.notifyHeartbeat()
	}
}

func (rf *Raft) broadcastHeartbeat(term int) {
	args := HeartbeatArgs{
		Term:           term,
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
	rf.mu.Lock()
	oldRole := rf.role
	oldTerm := rf.currentTerm
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		rf.mu.Unlock()
		return nil
	}
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.votedFor = ""
		rf.votes = nil
	}
	rf.role = RAFT_FOLLOWER
	rf.cn.SchedulerCanStart = args.SchedulerCanStart
	reply.Term = rf.currentTerm
	rf.mu.Unlock()

	if oldRole == RAFT_LEADER && args.Term >= oldTerm {
		rf.stopSchedulerAfterStepDown()
	}

	rf.cn.SetEndpointMain(args.LeaderEndpoint)
	rf.notifyHeartbeat()
	return nil
}

func (rf *Raft) activateLeader(term int) {
	rf.mu.Lock()
	if rf.role != RAFT_LEADER || rf.currentTerm != term {
		rf.mu.Unlock()
		return
	}
	canStart := rf.cn.SchedulerCanStart
	rf.mu.Unlock()

	slog.Info(fmt.Sprintf("Cluster node: `%s`, I'm leader", rf.cn.Endpoint))
	rf.cn.SetEndpointMain(rf.cn.Endpoint)
	rf.cn.registerNode(rf.cn)
	if canStart {
		rf.cn.Scheduler.Start()
	}
}

func (rf *Raft) runFollower(ctx context.Context) bool {
	timer := time.NewTimer(raftElectionTimeout())
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-rf.heartbeatC:
		return true
	case <-timer.C:
		slog.Warn(fmt.Sprintf("Follower: `%s` timeout", rf.cn.Endpoint))
		rf.beginElection()
		return true
	}
}

func (rf *Raft) runCandidate(ctx context.Context, term int) bool {
	slog.Info(fmt.Sprintf("Cluster node: `%s`, I'm candidate", rf.cn.Endpoint))
	timer := time.NewTimer(raftElectionTimeout())
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case leaderTerm := <-rf.leaderC:
			if leaderTerm == term {
				return true
			}
		case <-rf.heartbeatC:
			role, currentTerm := rf.state()
			if role != RAFT_CANDIDATE || currentTerm != term {
				return true
			}
		case <-timer.C:
			rf.mu.Lock()
			if rf.role == RAFT_CANDIDATE && rf.currentTerm == term {
				rf.role = RAFT_FOLLOWER
			}
			rf.mu.Unlock()
			return true
		}
	}
}

func (rf *Raft) runLeader(ctx context.Context, term int) bool {
	role, currentTerm := rf.state()
	if role != RAFT_LEADER || currentTerm != term {
		return true
	}

	rf.activateLeader(term)
	rf.broadcastHeartbeat(term)
	ticker := time.NewTicker(raftHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			role, currentTerm = rf.state()
			if role != RAFT_LEADER || currentTerm != term {
				return true
			}
			rf.broadcastHeartbeat(term)
		case <-rf.heartbeatC:
			role, currentTerm = rf.state()
			if role != RAFT_LEADER || currentTerm != term {
				return true
			}
		}
	}
}

func (rf *Raft) start(ctx context.Context) {
	slog.Info("Raft start.")
	rf.init()

	go func() {
		for {
			role, term := rf.state()
			var keepRunning bool
			switch role {
			case RAFT_FOLLOWER:
				keepRunning = rf.runFollower(ctx)
			case RAFT_CANDIDATE:
				keepRunning = rf.runCandidate(ctx, term)
			case RAFT_LEADER:
				keepRunning = rf.runLeader(ctx, term)
			default:
				keepRunning = false
			}
			if !keepRunning {
				return
			}
		}
	}()
}
