# 6.824-2016
mit 6.824 distribution

mepreduce: all done 5/5

gfs: paper 
     2.5 On the other hand ....    However ...
     2.6.3 The master recovers ... 
		how checkpoints implements???
			save checkpoint status?? and recover subsequence log??? 
	 2.7 
		consistent :  all clients will always see the same data, regardless of which replicas they read from
		defined : consistent + clients see what the mutation writes 
		consistent but not defined : clients not see every mutation 

gfs read done


raft: paper

lab: 
initialElection done
reElection processing  
	Q: if leader is down, how to define the majority of new reElection? 

2019.09.19
	reElection sometimes error sometimes succeed
	error1: disconnect two server expecting no leader but one
		reason: in former step, leader first disconnect and reElect new leader
			next recover the disconnected leader but it hasn't send hb to the higher term server , which 
			make it become follower. the other server disconnecting then ,and leader is the first disconnected server but wo need no leader.but recover before check leader
	error2: no leader




lab2  log entry

question
如果commitIndex > lastApplied，那么就 lastApplied 加一，并把log[lastApplied]应用到状态机中（5.3 节）
如果对于一个跟随者，最后日志条目的索引值大于等于 nextIndex，那么：发送从 nextIndex 开始的所有日志条目：
如果成功：更新相应跟随者的 nextIndex 和 matchIndex
如果因为日志不一致而失败，减少 nextIndex 重试
如果存在一个满足N > commitIndex的 N，并且大多数的matchIndex[i] ≥ N成立，并且log[N].term == currentTerm成立，那么令 commitIndex 等于这个 N （5.3 和 5.4 节）
last two are for leaders


question 

问题在于cfg.log 一直没有赋值，应该有一个地方会commit

先到这里告一段落，2019.11.19
完成内容
lab1 全部
lab2 选举
