package mapreduce

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// schedule starts and waits for all tasks in the given phase (Map or Reduce).
func (mr *Master) schedule(phase jobPhase) {
	var ntasks int
	var nios int // number of inputs (for reduce) or outputs (for map)
	switch phase {
	case mapPhase:
		ntasks = len(mr.files)
		nios = mr.nReduce
	case reducePhase:
		ntasks = mr.nReduce
		nios = len(mr.files)
	}

	fmt.Printf("Schedule: %v %v tasks (%d I/Os)\n", ntasks, phase, nios)

	// them have been completed successfully should the function return.
	// Remember that workers may fail, and that any given worker may finish
	// multiple tasks.	// All ntasks tasks have to be scheduled on workers, and only once all of

	//
	// TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO TODO
	// mycode
	var wg sync.WaitGroup
	wg.Add(ntasks)
	i := int32(-1)

	if phase == mapPhase {
		for {
			var worker string
			worker = <-mr.registerChannel
			fmt.Printf("mapwork %s\n", worker)
			if worker == "done" {
				break
			}
			go func() {
				for {
					t := atomic.AddInt32(&i, 1)
					if t > int32(ntasks) {
						break
					}
					if t == int32(ntasks) {
						mr.registerChannel <- "done"
						break
					}
					file := mr.files[t]
					args := DoTaskArgs{"test", file, phase, int(i), nios}
					//rpc worker possible have error
					if !call(worker, "Worker.DoTask", args, new(struct{})) {
						fmt.Printf("mapworkfailed %s\n", worker)
						t = atomic.AddInt32(&i, -1)
						fmt.Printf("failed t = %d\n", t)
						//recall after one seconds
						time.Sleep(1e9)
						// break
					} else {
						fmt.Printf("success t = %d\n", t)
						wg.Done()
					}
				}
			}()
		}
		wg.Wait()
	}

	if phase == reducePhase {
		fmt.Printf("workers num=%d\n", len(mr.workers))
		fmt.Println("reduce begin")
		for _, worker := range mr.workers {
			// fmt.Printf("reduce worker=%s\n", worker)
			go func() {
				for {
					t := atomic.AddInt32(&i, 1)
					// fmt.Printf("t = %d\n", t)
					if t >= int32(ntasks) {
						fmt.Println("reduce ntasks all done")
						break
					}
					file := mr.files[t]
					args := DoTaskArgs{"test", file, phase, int(i), nios}
					//rpc worker possible have error
					if !call(worker, "Worker.DoTask", args, new(struct{})) {
						// fmt.Printf("111 workers num=%d\n", len(mr.workers))
						//fmt.Println("worker is failed")
						t = atomic.AddInt32(&i, -1)
						fmt.Printf("reduce failed t = %d\n", t)
						time.Sleep(1e9)
						// break
					} else {
						wg.Done()
					}
				}
			}()
		}
		wg.Wait()
	}

	fmt.Printf("Schedule: %v phase done\n", phase)
}
