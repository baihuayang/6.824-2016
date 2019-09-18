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

