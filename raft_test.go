package agscheduler

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newRaftTestNode(endpoint string, nodeCount int) *ClusterNode {
	cn := &ClusterNode{
		EndpointMain: endpoint,
		Endpoint:     endpoint,
		Mode:         "HA",
	}
	for i := 0; i < nodeCount; i++ {
		cn.registerNode(&ClusterNode{
			EndpointMain: endpoint,
			Endpoint:     fmt.Sprintf("node-%d", i),
			Mode:         "HA",
		})
	}
	cn.nodeMapM.Lock()
	for peer, node := range cn.nodeMap {
		if peer != endpoint {
			node["health"] = false
		}
	}
	cn.nodeMapM.Unlock()
	return cn
}

func TestRaftQuorumSize(t *testing.T) {
	tests := []struct {
		nodeCount int
		want      int
	}{
		{nodeCount: 1, want: 1},
		{nodeCount: 3, want: 2},
		{nodeCount: 5, want: 3},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d_nodes", tt.nodeCount), func(t *testing.T) {
			rf := &Raft{cn: newRaftTestNode("node-0", tt.nodeCount)}
			assert.Equal(t, tt.want, rf.quorumSize())
		})
	}
}

func TestRaftSingleNodeElection(t *testing.T) {
	rf := &Raft{cn: newRaftTestNode("node-0", 1)}
	rf.init()

	term := rf.beginElection()
	role, currentTerm := rf.state()
	assert.Equal(t, RAFT_LEADER, role)
	assert.Equal(t, term, currentTerm)

	select {
	case leaderTerm := <-rf.leaderC:
		assert.Equal(t, term, leaderTerm)
	default:
		t.Fatal("single-node election did not notify leader transition")
	}
}

func TestRaftThreeNodeElectionAfterMajority(t *testing.T) {
	rf := &Raft{cn: newRaftTestNode("node-0", 3)}
	rf.init()

	term := rf.beginElection()
	role, _ := rf.state()
	assert.Equal(t, RAFT_CANDIDATE, role)

	rf.recordVote(term, "node-1")
	role, currentTerm := rf.state()
	assert.Equal(t, RAFT_LEADER, role)
	assert.Equal(t, term, currentTerm)

	select {
	case leaderTerm := <-rf.leaderC:
		assert.Equal(t, term, leaderTerm)
	default:
		t.Fatal("majority vote did not notify leader transition")
	}

	rf.recordVote(term, "node-2")
	rf.mu.Lock()
	assert.Len(t, rf.votes, 2)
	rf.mu.Unlock()
}

func TestRaftConcurrentRequestVote(t *testing.T) {
	rf := &Raft{cn: newRaftTestNode("node-0", 3)}
	rf.init()

	const requests = 100
	replies := make(chan VoteReply, requests)
	var wg sync.WaitGroup
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var reply VoteReply
			err := rf.RPCRequestVote(VoteArgs{
				Term:              1,
				CandidateEndpoint: fmt.Sprintf("candidate-%d", i),
			}, &reply)
			assert.NoError(t, err)
			replies <- reply
		}(i)
	}
	wg.Wait()
	close(replies)

	granted := 0
	for reply := range replies {
		assert.Equal(t, 1, reply.Term)
		if reply.VoteGranted {
			granted++
		}
	}
	assert.Equal(t, 1, granted)
}

func TestRaftHigherTermVoteStepsDownLeader(t *testing.T) {
	rf := &Raft{cn: newRaftTestNode("node-0", 3)}
	rf.init()
	rf.mu.Lock()
	rf.role = RAFT_LEADER
	rf.currentTerm = 2
	rf.votedFor = rf.cn.Endpoint
	rf.mu.Unlock()

	var reply VoteReply
	err := rf.RPCRequestVote(VoteArgs{
		Term:              3,
		CandidateEndpoint: "node-1",
	}, &reply)
	assert.NoError(t, err)
	assert.True(t, reply.VoteGranted)
	assert.Equal(t, 3, reply.Term)

	role, term := rf.state()
	assert.Equal(t, RAFT_FOLLOWER, role)
	assert.Equal(t, 3, term)
	select {
	case <-rf.heartbeatC:
	default:
		t.Fatal("leader step-down did not notify the state machine")
	}
}

func TestRaftRejectsStaleRequestVote(t *testing.T) {
	rf := &Raft{cn: newRaftTestNode("node-0", 3)}
	rf.init()
	rf.mu.Lock()
	rf.role = RAFT_LEADER
	rf.currentTerm = 5
	rf.votedFor = rf.cn.Endpoint
	rf.mu.Unlock()

	var reply VoteReply
	err := rf.RPCRequestVote(VoteArgs{
		Term:              4,
		CandidateEndpoint: "node-1",
	}, &reply)
	assert.NoError(t, err)
	assert.False(t, reply.VoteGranted)
	assert.Equal(t, 5, reply.Term)

	role, term := rf.state()
	assert.Equal(t, RAFT_LEADER, role)
	assert.Equal(t, 5, term)
}

func TestRaftRejectsSameTermRequestVoteWithoutSteppingDown(t *testing.T) {
	rf := &Raft{cn: newRaftTestNode("node-0", 3)}
	rf.init()
	rf.mu.Lock()
	rf.role = RAFT_LEADER
	rf.currentTerm = 5
	rf.votedFor = rf.cn.Endpoint
	rf.mu.Unlock()

	var reply VoteReply
	err := rf.RPCRequestVote(VoteArgs{
		Term:              5,
		CandidateEndpoint: "node-1",
	}, &reply)
	assert.NoError(t, err)
	assert.False(t, reply.VoteGranted)
	assert.Equal(t, 5, reply.Term)

	role, term := rf.state()
	assert.Equal(t, RAFT_LEADER, role)
	assert.Equal(t, 5, term)
}

func TestRaftHeartbeatUpdatesFollower(t *testing.T) {
	cn := newRaftTestNode("node-0", 3)
	cn.EndpointMain = "old-main"
	rf := &Raft{cn: cn}
	rf.init()

	var reply HeartbeatReply
	err := rf.RPCHeartbeat(HeartbeatArgs{
		Term:              4,
		LeaderEndpoint:    "node-1",
		SchedulerCanStart: true,
	}, &reply)
	assert.NoError(t, err)
	assert.Equal(t, 4, reply.Term)
	assert.Equal(t, "node-1", cn.GetEndpointMain())

	rf.mu.Lock()
	assert.True(t, cn.SchedulerCanStart)
	rf.mu.Unlock()

	select {
	case <-rf.heartbeatC:
	default:
		t.Fatal("heartbeat did not reset the election timer")
	}
}

func TestRaftRejectsStaleHeartbeat(t *testing.T) {
	cn := newRaftTestNode("node-0", 3)
	cn.EndpointMain = "node-0"
	rf := &Raft{cn: cn}
	rf.init()
	rf.mu.Lock()
	rf.role = RAFT_LEADER
	rf.currentTerm = 5
	rf.mu.Unlock()

	var reply HeartbeatReply
	err := rf.RPCHeartbeat(HeartbeatArgs{
		Term:           4,
		LeaderEndpoint: "node-1",
	}, &reply)
	assert.NoError(t, err)
	assert.Equal(t, 5, reply.Term)
	assert.Equal(t, "node-0", cn.GetEndpointMain())

	role, term := rf.state()
	assert.Equal(t, RAFT_LEADER, role)
	assert.Equal(t, 5, term)
}
