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
	//command interface{}
	Command interface{}
	Term    int
	//term    int
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
	voteChan chan bool

	//become leader chan
	becomeLeaderChan chan bool

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
	//Entries      []interface{}
	Entries      []LogEntry
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
	rf.mu.Lock()
	defer rf.mu.Unlock()
	go func (){
		if rf.status == 0{
			//only follower reset time count
			//fmt.Printf("from %v to %v send voteChan begin\n", args.CandidateId, rf.me)
			//fmt.Printf("[RequestVote server receive %v now:%v]",rf.me,time.Now())
			rf.voteChan<-true
			//fmt.Printf("from %v to %vsend voteChan done\n", args.CandidateId, rf.me)
		}
	}()
	// Your code here.
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		//fmt.Printf("[get vote failed 1] server %v request with LastLogIndex = %v\n", args.CandidateId, args.LastLogIndex)
		return
	}
	//candidate begin a new term vote
	//rf.voteChan<-true
	if args.Term > rf.currentTerm {
		//fmt.Printf("[get vote success] server %v request with LastLogIndex = %v\n", args.CandidateId, args.LastLogIndex)
		reply.Term = rf.currentTerm
		reply.VoteGranted = true
		rf.votedFor = args.CandidateId
		rf.currentTerm = args.Term
		// become follower(0)
		//fmt.Printf("[back to follower] server %v\n",rf.me)
		rf.status = 0
		return
	}
	//vote for same term
	if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
		if args.Term >= rf.currentTerm && args.LastLogIndex >= len(rf.log)-1 {
			rf.votedFor = args.CandidateId
			reply.Term = rf.currentTerm
			reply.VoteGranted = true
			//fmt.Printf("[get vote success] server %v request with LastLogIndex = %v\n", args.CandidateId, args.LastLogIndex)
			return
		} else {
			//fmt.Printf("[get vote failed 2] server %v request with LastLogIndex = %v\n", args.CandidateId, args.LastLogIndex)
		}
	}else{
		//fmt.Printf("[get vote failed 3] server %v args.term %v vote for server %v rf.currentTerm %v request with LastLogIndex = %v and votor votedFor = %v\n",
		//	args.CandidateId, args.Term, rf.me, rf.currentTerm, args.LastLogIndex, rf.votedFor)
	}
	return
}

// appendEntres RPC handler
func (rf *Raft) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) {
	if len(args.Entries) == 0 {
		//fmt.Printf("[AppendEntries heartbeat] args : %v\n", args)
		rf.mu.Lock()
		go func() {
			rf.heartbeatChan <- true
		}()
		//heartbeats
		//reset time for candidate
		if args.Term < rf.currentTerm {
			fmt.Printf("[heartBeat: server %v hb failed] for term is not latest, args.term %v, rf.term %v\n",
				args.LeaderId, args.Term, rf.currentTerm)
			reply.Term = rf.currentTerm
			reply.Success = false
			return
		}
		//fmt.Printf("[heartBeat: server %v term %v receive hb] chan sending\n", rf.me, rf.currentTerm)
		reply.Term = rf.currentTerm
		reply.Success = true
		if rf.currentTerm < args.Term {
			rf.currentTerm = args.Term
		}
		//anyway become follower
		//fmt.Printf("server %v become a follower\n", rf.me)
		rf.status = 0
		rf.mu.Unlock()
		return
	} else {
		fmt.Printf("[AppendEntries logentry=%v", args)
		if args.Term < rf.currentTerm {
			fmt.Printf("[AppendEntries failed 1] [log entry: server %v failed] for term is not latest, args.term %v, rf.term %v\n",
				args.LeaderId, args.Term, rf.currentTerm)
			reply.Term = rf.currentTerm
			reply.Success = false
			return
		}
		//append entries
		//leader prevLogIndex is too large
		if len(rf.log) < args.PrevLogIndex{
			fmt.Println("[AppendEntries failed 2] len(rf.log) < args.PrevLogIndex and return")
			//do something
			//fmt.Println("do something")
			reply.Success = false
			return
		}
		fmt.Printf("args.PrevLogIndex = %v\n", args.PrevLogIndex)
		if args.PrevLogIndex < 0{
			//first log entry
			if len(rf.log) > args.PrevLogIndex+1{
				rf.log = rf.log[:args.PrevLogIndex+1]
			}
			for _, le := range args.Entries{
				rf.log = append(rf.log, le)
			}
			reply.Success = true
		}else{
			//fmt.Printf("rf.log[args.PrevLogIndex].(LogEntry) = %v\n", rf.log[args.PrevLogIndex])
			logEntry := rf.log[args.PrevLogIndex].(LogEntry)
			if logEntry.Term != args.PrevLogTerm{
				fmt.Println("[AppendEntries failed 3]false 1111")
				//remove PrevLogIndex -> ...
				rf.log = rf.log[:args.PrevLogIndex]
				//append new log
				reply.Success = false
				return
			}else{
				if len(rf.log) > args.PrevLogIndex+1{
					rf.log = rf.log[:args.PrevLogIndex+1]
				}
				for _, le := range args.Entries{
					rf.log = append(rf.log, le)
				}
				reply.Success = true
			}
		}
		//todo 如果 leaderCommit > commitIndex，令 commitIndex 等于 leaderCommit 和 新日志条目索引值中较小的一个
		if args.LeaderCommit > rf.commitIndex {
			if args.LeaderCommit > len(rf.log)-1 {
				rf.commitIndex = len(rf.log) - 1
			}else{
				rf.commitIndex = args.LeaderCommit
			}
		}
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
	//fmt.Printf("[sendRequestVote] at time=%v\n", time.Now())
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	//fmt.Printf("OK1?????? from server %v to server %v : %v\n", args.CandidateId, server, ok)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	//fmt.Printf("OK2?????? from server %v to server %v : %v\n", args.LeaderId, server, ok)
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
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	//Start(command) asks Raft to start the processing to append the command to the replicated log
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.status != 2{
		//not a leader
		return -1,-1,false
	}
	//add to leader first
	newLog := LogEntry{command, rf.currentTerm}
	rf.log = append(rf.log, newLog)
	index := -1
	for i := 0; i < len(rf.peers); i++ {
		//term := -1
		//isLeader := true
		fmt.Printf("rf.log len = %v\n",len(rf.log))
		if rf.me == i{
			continue
		}
		index = rf.nextIndex[i]
		index = rf.sendLog(index, command, i)
	}
	//todo what should first return value be
	return index, rf.currentTerm, true
}

func (rf *Raft) sendLog(index int, command interface{}, i int) int {
	fmt.Printf("sendLog with index %d\n", index)
	log := rf.log[index].(LogEntry)
	log.Term = rf.currentTerm
	prevLogIndex := index - 1
	prevLogTerm := -1
	if prevLogIndex >= 0 {
		prevLog := rf.log[prevLogIndex].(LogEntry)
		prevLogTerm = prevLog.Term
	}
	var logs []LogEntry
	n := LogEntry{Command: command, Term:rf.currentTerm}
	logs = append(logs, n)
	//logs := []interface{}{LogEntry{command, rf.currentTerm}}
	req := AppendEntriesArgs{
		rf.currentTerm,
		rf.me,
		prevLogIndex,
		prevLogTerm,
		logs,
		rf.commitIndex}
	reply := AppendEntriesReply{}
	rf.sendAppendEntries(i, req, &reply)
	if reply.Success {
		rf.nextIndex[i] = rf.nextIndex[i] + 1
	} else {
		//--nextIndex and retry
		fmt.Println("retrying")
		index = rf.sendLog(index-1, command, i)
	}
	return index
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
	//about election init
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.heartbeatChan = make(chan bool)
	rf.voteChan = make(chan bool)
	rf.becomeLeaderChan = make(chan bool)
	// Your initialization code here.
	rf.votedFor = -1
	//about log init
	rf.log = make([]interface{}, 0)
	rf.nextIndex = make([]int, len(peers))
	rf.matchIndex = make([]int, len(peers))
	// if can not receive info (heartbeat) from others(leader) change to candidate for vote
	go func() {
		for {
			if rf.status == 0 {
				//follower code
				rand.Seed(time.Now().UnixNano())
				// 500ms - 1000ms random
				ranN := time.Duration(rand.Intn(500) + 500)
				//candidate or follower code
				select {
				case <-rf.heartbeatChan:
					break
				case <-rf.voteChan:
					break
				case <-time.After(ranN * time.Millisecond):
					//go to candidate
					rf.mu.Lock()
					rf.currentTerm += 1
					rf.votedFor = -1
					//become candidate(1)
					rf.status = 1
					rf.voteNumber = 0
					rf.mu.Unlock()
					//fmt.Printf("server %v term %v votefor wait %v time\n", me, rf.currentTerm, ranN * time.Millisecond)
					for i := 0; i < len(peers); i++ {
						go beginVote(rf, me, i, peers)
					}
					break
				}
			} else if rf.status == 1{
				//candidate code
				//similar to follower code
				rand.Seed(time.Now().UnixNano())
				// 150ms - 300ms random
				ranN := time.Duration(rand.Intn(500) + 500)
				//candidate or follower code
				select {
				case <-rf.heartbeatChan:
					break
				case <-rf.becomeLeaderChan:
					break
				case <-time.After(ranN * time.Millisecond):
					fmt.Printf("[case revote2]server %v term %v not receive hb\n", me, rf.currentTerm)
					//go to candidate
					rf.mu.Lock()
					rf.currentTerm += 1
					rf.votedFor = -1
					//become candidate(1)
					rf.status = 1
					rf.voteNumber = 0
					rf.mu.Unlock()
					for i := 0; i < len(peers); i++ {
						go beginVote(rf, me, i, peers)
					}
					break
				}
			} else {
				//leader code
				time.Sleep(100 * time.Millisecond)
				for i := 0; i < len(peers); i++ {
					go beginHeartBeat(i, me, rf)
				}
			}
		}
	}()

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	return rf
}

func beginHeartBeat(i int, me int, rf *Raft) {
	if i != me {
		lastLogTerm := 0
		lastLogIndex := len(rf.log) - 1
		if lastLogIndex >= 0 {
			lastLog := rf.log[lastLogIndex].(LogEntry)
			lastLogTerm = lastLog.Term
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
			responseOk := rf.sendAppendEntries(i, req, &reply)
			if responseOk {
				if !reply.Success {
					rf.mu.Lock()
					rf.status = 0
					rf.currentTerm = reply.Term
					rf.mu.Unlock()
				}
			}
		}
	}
}

func beginVote(rf *Raft, me int, i int, peers []*labrpc.ClientEnd) {
	lastLogTerm := 0
	lastLogIndex := len(rf.log) - 1
	if lastLogIndex >= 0 {
		lastLog := rf.log[lastLogIndex].(LogEntry)
		lastLogTerm = lastLog.Term
	}
	req := RequestVoteArgs{rf.currentTerm, me, lastLogIndex, lastLogTerm}
	reply := RequestVoteReply{}
	responseOk := rf.sendRequestVote(i, req, &reply)
	if responseOk {
		if reply.VoteGranted {
			rf.mu.Lock()
			rf.voteNumber += 1
			if rf.voteNumber > len(peers)/2 {
				//become leader  status = 2
				rf.status = 2
				defer func() {
					rf.becomeLeaderChan<-true
				}()
			}
			rf.mu.Unlock()
		}
	}
}
