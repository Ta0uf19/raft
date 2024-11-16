package raft

import (
	"math/rand"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// ConsensusModule has a State Machine with 3 States
// 1 - Follower
// 2 - Candidate (timeout start an election)
// 3 - Leader (receive votes from majority)

// CMState by default is a follower
type CMState int

const (
	Follower  CMState = 0
	Candidate         = 1
	Leader            = 2
	Dead              = 3
)

func (state CMState) String() string {
	switch state {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	case Dead:
		return "Dead"
	default:
		panic("unreachable")
	}
}

// ConsensusModule receives commands from clients and adds them to its log.
// It communicates with other ConsensusModules to ensure that every log entry is safely replicated.
type ConsensusModule struct {
	id       int     // server id
	peersIds []int   // lists the IDs of our peers in the cluster
	server   *Server // used to send RPCs to other servers

	// Persistent state on all servers
	mu          sync.Mutex // Lock to protect shared access to this peer's state
	currentTerm int        // latest term server has seen (initialized to 0 on first boot, increases monotonically)
	votedFor    int        // candidateId that received vote in current term (or null if none)

	// Volatile state on all servers
	state           CMState   // current state of the server (follower, candidate, leader)
	leaderHeartbeat time.Time // time when we last heard from the leader
}

func NewConsensusModule(id int, peersIds []int, server *Server) *ConsensusModule {
	cm := &ConsensusModule{
		id:       id,
		peersIds: peersIds,
		state:    Follower,
		server:   server,
		votedFor: -1,
	}
	cm.leaderHeartbeat = time.Now()
	// Start the election timer
	go cm.runElectionTimer()
	return cm
}

// runElectionTimer runs the election timer in a separate goroutine
// Every follower runs an election timer that restarts every time it hears from the leader
// If the election timer expires, the follower starts an election
func (cm *ConsensusModule) runElectionTimer() {
	timeoutDuration := cm.electionTimeout()
	cm.mu.Lock()
	termStarted := cm.currentTerm
	cm.mu.Unlock()

	// When the election timer expires, start a new election
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cm.mu.Lock()
			if cm.state != Follower {
				log.Debugf("Server %v is not a follower", cm.id)
				cm.mu.Unlock()
				return
			}
			// If the term has changed, it means that a leader has been elected
			if termStarted != cm.currentTerm {
				log.Debugf("Server %v term has changed %d to %d", cm.id, termStarted, cm.currentTerm)
				cm.mu.Unlock()
				return
			}

			// Start an election for follower, if we haven't heard from the leader for a while or haven't voted for a candidate
			if time.Since(cm.leaderHeartbeat) > timeoutDuration {
				cm.startElection()
				return
			}
			cm.mu.Unlock()
		}
	}
}

func (cm *ConsensusModule) startElection() {
	cm.state = Candidate
	cm.currentTerm++
	cm.votedFor = cm.id
	cm.leaderHeartbeat = time.Now()
	cm.sendRequestVote()
}

func (cm *ConsensusModule) electionTimeout() time.Duration {
	// Election timeout is a random value between 150ms and 300ms as per the Raft paper
	return time.Duration(150+rand.Intn(150)) * time.Millisecond
}

func (cm *ConsensusModule) sendRequestVote() {
	// Send RequestVote RPCs to all other servers in the cluster
	for _, peerId := range cm.peersIds {
		if peerId == cm.id {
			continue
		}
		go func(peerId int) {
			// Send RequestVote RPC to peer
			// If peer votes for us, we increment the vote count
			// If we receive votes from majority, we become the leader
			// If we receive a heartbeat from the leader, we reset the election timer
		}(peerId)
	}
}

// startLeader changes the state of the server to Leader
func (cm *ConsensusModule) startLeader() {
	cm.state = Leader
	// Send heartbeats to all other servers in the cluster as long as we are the leader
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			cm.sendHeartbeats()
			select {
			case <-ticker.C:
				// check if we are still the leader
				cm.mu.Lock()
				if cm.state != Leader {
					cm.mu.Unlock()
					return
				}
				cm.mu.Unlock()
			}
		}
	}()
}

func (cm *ConsensusModule) sendHeartbeats() {
	// Send AppendEntries RPCs to all other servers in the cluster
	for _, peerId := range cm.peersIds {
		if peerId == cm.id {
			continue
		}
		go func(peerId int) {
			// Send AppendEntries RPC to peer
			// If peer receives the heartbeat, it resets the election timer
		}(peerId)
	}
}
