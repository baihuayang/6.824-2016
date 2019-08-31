package mapreduce

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"sort"
)

// doReduce does the job of a reduce worker: it reads the intermediate
// key/value pairs (produced by the map phase) for this task, sorts the
// intermediate key/value pairs by key, calls the user-defined reduce function
// (reduceF) for each key, and writes the output to disk.
func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTaskNumber int, // which reduce task this is
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {
	// TODO:
	// You will need to write this function.
	// You can find the intermediate file for this reduce task from map task number
	// m using reduceName(jobName, m, reduceTaskNumber).
	// Remember that you've encoded the values in the intermediate files, so you
	// will need to decode them. If you chose to use JSON, you can read out
	// multiple decoded values by creating a decoder, and then repeatedly calling
	// .Decode() on it until Decode() returns an error.
	//
	// You should write the reduced output in as JSON encoded KeyValue
	// objects to a file named mergeName(jobName, reduceTaskNumber). We require
	// you to use JSON here because that is what the merger than combines the
	// output from all the reduce tasks expects. There is nothing "special" about
	// JSON -- it is just the marshalling format we chose to use. It will look
	// something like this:
	//
	// enc := json.NewEncoder(mergeFile)
	// for key in ... {
	// 	enc.Encode(KeyValue{key, reduceF(...)})
	// }
	// file.Close()

	//my code
	//read intermediate files and decode json
	//fmt.Printf("reduce: jobName:%s mapTaskNumber:%d reduce begin\n", jobName, reduceTaskNumber)
	smap := make(map[string][]string)
	for i := 0; i < nMap; i++ {
		imFileName := reduceName(jobName, i, reduceTaskNumber)
		//fmt.Printf("reduce number:%d imFileName:%s mapNumber:%d nMap:%d\n", reduceTaskNumber, imFileName, i, nMap)
		inputFile, error := os.Open(imFileName)
		if error != nil {
			//fmt.Printf("inputFile isn't exist with name:%s\n", imFileName)
			//it's so suck , master will cleaning all files range from 0-0 to 4-2 after mr done, will not pass if file not exist
			//so create file if not exist
			os.Create(imFileName)
			continue
		}
		defer inputFile.Close()
		inputReader := bufio.NewReader(inputFile)
		for {
			inputString, readerError := inputReader.ReadString('\n')
			if readerError == io.EOF {
				break
			}

			var t KeyValue
			json.Unmarshal([]byte(inputString), &t)
			if smap[t.Key] != nil {
				s := smap[t.Key]
				s = append(s, t.Value)
				smap[t.Key] = s
			} else {
				var s []string
				s = append(s, t.Value)
				smap[t.Key] = s
			}

		}
		//fmt.Printf("smap nMap:%d;len=%d\n", i, len(smap))
	}
	//fmt.Printf("smap total len=%d\n", len(smap))
	//fmt.Print(smap)
	//fmt.Println("final loop done")
	//sort
	keys := make([]string, len(smap))
	i := 0
	for k, _ := range smap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	//fmt.Println("sort done")
	mergeFile := mergeName(jobName, reduceTaskNumber)
	//fmt.Printf("mergeFileName=%s\n", mergeFile)
	file, _ := os.OpenFile(mergeFile, os.O_CREATE|os.O_WRONLY, 0666)
	defer file.Close()
	//enc := json.NewEncoder(file)
	for _, k := range keys {
		enc := json.NewEncoder(file)
		kv := KeyValue{k, reduceF(k, smap[k])}
		enc.Encode(&kv)
	}
	//fmt.Printf("reduce number:%d reduce done\n", reduceTaskNumber)
}
