#!/bin/sh
>result.txt
if [ "$#" -ne 1 ]; then
    echo "Usage: args number must be 1"
    exit 1
fi
max=100
succeedCount=0
for i in `seq 1 $max`
do
   >tmp.txt
   #go test -run $1 | grep 'ok' | grep 'raft' >> result.txt
   #res=`go test -run $1 | grep 'ok' | grep 'raft'` 
   #goRes=`go test -run $1`
   go test -run $1>tmp.txt
   
   okRes=`grep 'ok' tmp.txt | grep 'raft'`
   #echo $okRes 
   if [ ! -z "$okRes" ]; then
       successTxt=$1_success_$i.log
       cp tmp.txt logs/success/$successTxt
       succeedCount=$(($succeedCount+1))
   else
       errTxt=$1_error_$i.log
       cp tmp.txt logs/fail/$errTxt
   fi
   echo "Test $i done"
done
rm tmp.txt
echo "total succeed $succeedCount in $max"
exit 1

