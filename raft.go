package raft

// ConsensusModule has a State Machine with 3 States
// 1 - Follower
// 2 - Candidate (timeout start an election)
// 3 - Leader (receive votes from majority)

// By default CMState is follower
type CMState int

const (
	Follower CMState	= 0
	Candidate 			= 1
	Leader 				= 2
	Dead 				= 3
)

func (state CMState) String() string  {
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

type ConsensusModule struct {
	id int // server id of this CM
	peersIds []int // lists the IDs of our peers in the cluster


}

// Election timer
// Timer every follower runs continuously, restarting it every time it hears from the
// current leader.
// Leader must sent period heartbeats to followers
