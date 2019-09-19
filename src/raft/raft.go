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
	"math/rand"
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

	//leader:2  candidate:1  follower:0
	status int
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	//var term int
	//var isleader bool
	// Your code here.
	return rf.currentTerm, rf.status == 2
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
	//fmt.Printf("RequestVote before lock: from %v for %v vote\n", args.CandidateId, rf.me)
	//fmt.Printf("[required Lock] sendAppendEntries server %v\n", rf.me)
	rf.mu.Lock()
	//fmt.Printf("[Locked] RequestVote server %v\n", rf.me)
	defer rf.mu.Unlock()
	//defer fmt.Printf("[Unlocked] RequestVote server %v\n", rf.me)
	//fmt.Printf("RequestVote locked: from %v for %v vote\n", args.CandidateId, rf.me)
	//fmt.Println(args)
	// Your code here.
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		fmt.Printf("[get vote failed 1] server %v request with LastLogIndex = %v\n", args.CandidateId, args.LastLogIndex)
		return
	}
	//candidate begin a new term vote
	if args.Term > rf.currentTerm {
		fmt.Printf("[get vote success] server %v request with LastLogIndex = %v\n", args.CandidateId, args.LastLogIndex)
		reply.Term = rf.currentTerm
		reply.VoteGranted = true
		rf.votedFor = args.CandidateId
		rf.currentTerm = args.Term
		// become follower(0)
		rf.status = 0
		return
	}
	//vote for same term
	if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
		if args.Term >= rf.currentTerm && args.LastLogIndex >= len(rf.log)-1 {
			rf.votedFor = args.CandidateId
			reply.Term = rf.currentTerm
			reply.VoteGranted = true
			fmt.Printf("[get vote success] server %v request with LastLogIndex = %v\n", args.CandidateId, args.LastLogIndex)
			return
		} else {
			fmt.Printf("[get vote failed 2] server %v request with LastLogIndex = %v\n", args.CandidateId, args.LastLogIndex)
		}
	}else{
		fmt.Printf("[get vote failed 3] server %v args.term %v vote for server %v rf.currentTerm %v request with LastLogIndex = %v and votor votedFor = %v\n",
			args.CandidateId, args.Term, rf.me, rf.currentTerm, args.LastLogIndex, rf.votedFor)
	}
	return
}

// appendEntres RPC handler
//todo think need lock???
func (rf *Raft) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) {
	if len(args.Entries) == 0 {
		//heartbeats
		//reset time for candidate
		if args.Term < rf.currentTerm {
			fmt.Printf("[heartBeat: server %v hb failed] for term is not latest, args.term %v, rf.term %v\n",
				args.LeaderId, args.Term, rf.currentTerm)
			reply.Success = false
			return
		}
		fmt.Printf("[heartBeat: server %v term %v receive hb] chan sending\n", rf.me, rf.currentTerm)
		// todo can rf.heartbeatChan <- true use go routines???
		rf.heartbeatChan <- true
		reply.Term = rf.currentTerm
		reply.Success = true
		if rf.currentTerm < args.Term {
			rf.currentTerm = args.Term
		}
		//anyway become follower
		rf.mu.Lock()
		fmt.Printf("server %v become a follower\n", rf.me)
		rf.status = 0
		rf.mu.Unlock()
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
	fmt.Printf("OK1?????? from server %v to server %v : %v\n", args.CandidateId, server, ok)
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
func  Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.heartbeatChan = make(chan bool)
	// Your initialization code here.
	rf.votedFor = -1
	// if can not receive info (heartbeat) from others(leader) change to candidate for vote
	go func() {
		for {
			if !(rf.status == 2) {
				rand.Seed(time.Now().UnixNano())
				// 150ms - 300ms random
				ranN := time.Duration(rand.Intn(150) + 150)
				//candidate or follower code
				select {
				case <-rf.heartbeatChan:
					fmt.Printf("server %d term %v receive hb\n", me, rf.currentTerm)
					break
				case <-time.After(ranN * time.Millisecond):
					fmt.Printf("server %v term %v not receive hb\n", me, rf.currentTerm)
					//go to candidate
					rf.mu.Lock()
					rf.currentTerm += 1
					rf.votedFor = -1
					//become candidate(1)
					rf.status = 1
					rf.voteNumber = 0
					rf.mu.Unlock()
					fmt.Printf("server %v term %v votefor wait %v time\n", me, rf.currentTerm, ranN * time.Millisecond)
					for i := 0; i < len(peers); i++ {
						lastLogTerm := 0
						lastLogIndex := len(rf.log) - 1
						if lastLogIndex >= 0 {
							lastLog := rf.log[lastLogIndex].(LogEntry)
							lastLogTerm = lastLog.term
						}
						req := RequestVoteArgs{rf.currentTerm, me, lastLogIndex, lastLogTerm}
						reply := RequestVoteReply{}
						fmt.Printf("server %v send vote request to server %v\n", me, i)
						start := time.Now()
						responseOk := rf.sendRequestVote(i, req, &reply)
						fmt.Printf("server %v send vote request to server %v done\n", me, i)
						if responseOk {
							fmt.Printf("%s took %v\n", "<sendRequestVote Successed>", time.Since(start))
							if reply.VoteGranted {
								//fmt.Printf("[required Lock] sendAppendEntries server %v\n", rf.me)
								rf.mu.Lock()
								//fmt.Printf("[Locked] convert leader to follower server %v\n", rf.me)
								rf.voteNumber += 1
								fmt.Printf("server %v get vote from server %v with rf.term = %v, reply.term = %v\n", me, i, rf.currentTerm, reply.Term)
								fmt.Printf("server %v votenum = %v\n", me, rf.voteNumber)
								if rf.voteNumber > len(peers)/2 {
									//become leader  status = 2
									rf.status = 2
									fmt.Printf("server %v term %v become leader with vote num = %v\n", me, rf.currentTerm, rf.voteNumber)
									rf.mu.Unlock()
									//fmt.Printf("[Unlocked] convert leader to follower server %v\n", rf.me)
									break
								}
								rf.mu.Unlock()
								//fmt.Printf("[Unlocked] convert leader to follower server %v\n", rf.me)
							} else {
								fmt.Printf("server %v get reply term = %v, voteGranted = %v\n", me, reply.Term, reply.VoteGranted)
							}
						} else {
							fmt.Printf("%s took %v\n", "<sendRequestVote Failed>", time.Since(start))
						}
					}
					break
				}
			} else {
				//leader code
				time.Sleep(10 * time.Millisecond)
				//
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
						if rf.status == 2 {
							fmt.Printf("server %v term %v send hb to server %v and rf.status = %v\n", rf.me, rf.currentTerm, i, rf.status)
							start := time.Now()
							responseOk := rf.sendAppendEntries(i, req, &reply)
							if responseOk {
								fmt.Printf("%s took %v\n", "<sendAppendEntries Failed>", time.Since(start))
								fmt.Printf("server %v term %v send hb to server %v done\n", me, rf.currentTerm, i)
								if !reply.Success {
									fmt.Printf("server %v become follower from leader\n", me)
									rf.mu.Lock()
									rf.status = 0
									rf.mu.Unlock()
								}
							}else{
								fmt.Printf("%s took %v\n", "<sendAppendEntries Failed>", time.Since(start))
							}
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
