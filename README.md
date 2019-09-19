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
