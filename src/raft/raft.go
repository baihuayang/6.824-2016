package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"fmt"
	"sync"
	"time"
)
import "labrpc"

// import "bytes"
// import "encoding/gob"

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make().
//
type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}

type LogEntry struct {
	command interface{}
	term    int
}

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex
	peers     []*labrpc.ClientEnd
	persister *Persister
	me        int // index into peers[]

	// Your data here.
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	currentTerm int
	votedFor    int
	log         []interface{}

	commitIndex int
	lastApplied int

	//for leader
	nextIndex  []int
	matchIndex []int

	//heartbeat
	heartbeatChan chan bool

	//voted number by others
	voteNumber int

	//is leader
	isLeader bool
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	//var term int
	//var isleader bool
	// Your code here.
	return rf.currentTerm, rf.isLeader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here.
	// Example:
	// w := new(bytes.Buffer)
	// e := gob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	// Your code here.
	// Example:
	// r := bytes.NewBuffer(data)
	// d := gob.NewDecoder(r)
	// d.Decode(&rf.xxx)
	// d.Decode(&rf.yyy)
}

//
// example RequestVote RPC arguments structure.
//
type RequestVoteArgs struct {
	// Your data here.
	// candidate's term
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

//
// example RequestVote RPC reply structure.
//
type RequestVoteReply struct {
	// Your data here.
	//currentTeam
	Term        int
	VoteGranted bool
}

// append logs and heartbeats request
type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []interface{}
	LeaderCommit int
}

// append logs and heartbeats reply
type AppendEntriesReply struct {
	Term    int
	Success bool
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) {
	//fmt.Println("RequestVote: ")
	//fmt.Println(args)
	// Your code here.
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		fmt.Println("branch 1")
		return
	}
	//rf.votedFor == args.candidateId  (voted for itself)
	if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
		if args.Term >= rf.currentTerm && args.LastLogIndex >= len(rf.log)-1 {
			rf.votedFor = args.CandidateId
			reply.Term = rf.currentTerm
			reply.VoteGranted = true
			return
		}
	}
}

// appendEntres RPC handler
func (rf *Raft) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) {
	if len(args.Entries) == 0 {
		//heartbeats
		//reset time for candidate
		if args.Term < rf.currentTerm {
			reply.Success = false
			return
		}
		rf.heartbeatChan <- true
		reply.Term = rf.currentTerm
		reply.Success = true
		return
	} else {
		//append entries

	}
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// returns true if labrpc says the RPC was delivered.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	//Start(command) asks Raft to start the processing to append the command to the replicated log
	index := -1
	term := -1
	isLeader := true

	return index, term, isLeader
}

//
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// Your code here, if desired.
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.heartbeatChan = make(chan bool)
	// Your initialization code here.
	// if can not receive info (heartbeat) from others(leader) change to candidate for vote
	go func() {
		for {
			if !rf.isLeader {
				//candidate or follower code
				select {
				case <-rf.heartbeatChan:
					fmt.Printf("server %d receive hb\n", me)
					break
				case <-time.After(1000 * time.Millisecond):
					fmt.Printf("server %d not receive hb\n", me)
					//go to candidate
					for i := 0; i < len(peers); i++ {
						lastLogTerm := 0
						lastLogIndex := len(rf.log) - 1
						if lastLogIndex >= 0 {
							lastLog := rf.log[lastLogIndex].(LogEntry)
							lastLogTerm = lastLog.term
						}
						req := RequestVoteArgs{rf.currentTerm, me, lastLogIndex, lastLogTerm}
						reply := RequestVoteReply{}
						//fmt.Println("sendRequestVote")
						rf.sendRequestVote(i, req, &reply)
						if reply.VoteGranted {
							rf.mu.Lock()
							rf.voteNumber += 1
							if rf.voteNumber > len(peers) {
								//become leader
								rf.isLeader = true
								break
							}
							rf.mu.Unlock()
						}
					}
					break
				}
			} else {
				//leader code
				time.Sleep(100 * time.Millisecond)
				for i := 0; i < len(peers); i++ {
					if i != me {
						lastLogTerm := 0
						lastLogIndex := len(rf.log) - 1
						if lastLogIndex >= 0 {
							lastLog := rf.log[lastLogIndex].(LogEntry)
							lastLogTerm = lastLog.term
						}
						req := AppendEntriesArgs{
							rf.currentTerm,
							me,
							lastLogIndex,
							lastLogTerm,
							nil,
							rf.commitIndex}
						reply := AppendEntriesReply{}
						//fmt.Println("sendAppendEntries")
						rf.sendAppendEntries(i, req, &reply)
						if !reply.Success {
							rf.isLeader = false
						}
					}
				}
			}
		}
	}()

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	return rf
}
