package raft

// Server for Raft Consensus Module. Exposes raft to the network
// and enable RPCs between Raft peers.

// Server wraps ConsensusModule
type Server struct {
	id int              // Server Id
	cm *ConsensusModule // consensus module communicating with Log and sync StateMachine
}
