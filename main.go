package main

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

// ----------------------------------------------------------------------------
// Job implementation

type SimulatedJob struct {
	id      string
	jobQ    chan<- Job
	timeout int
}

func (j *SimulatedJob) Execute() error {
	fmt.Println(j.id, "executing")
	myWork := rand.Intn(10)

	// Now do whatever work we should do.
	// t := time.Now()
	fmt.Println(j.id, "has", myWork, "work to do")
	// simulate doing some work... for "myWork" number of seconds
	time.Sleep(time.Duration(myWork) * time.Second)
	q := rand.Intn(100)
	// fmt.Println(worker.id, "q:", q, "since:", time.Since(t).Seconds(), "workerTimeout:", workerTimeout)
	if q < 10 {
		// failed work unit
		// re-queue the work unit for re-processing.
		// this blocks if the work queue is full.
		j.jobQ <- j
		// simulate 10% chance of panic
		panic(fmt.Sprintf("%v with %d", j.id, q))
	} else if q < 20 {
		// failed work unit
		// re-queue the work unit for re-processing.
		// this blocks if the work queue is full.
		j.jobQ <- j
		// simulate 10% chance of failure
		return fmt.Errorf("on %d", q)
		// } else if since := time.Since(t).Seconds(); since > float64(j.timeout) {
		// 	fmt.Println(j.id, "timeout:", since, ">", j.timeout)
		// 	// simulate timeout extension
		// 	j.timeout = j.timeout + 2
		// 	fmt.Println(j.id, "workerTimeout extended to ", j.timeout)
		// 	// PONDER:  It could be that our worker needs to be "richer"
		// 	// and simulate a cumulative amount of work time and continue
		// 	// to work until some "hard stop" amount of time.  It might be
		// 	// that we need to signal a queue to extend our time, but
		// 	// keep working.  all that sort of logic would need to go here.

		// 	// fail the work unit
		// 	// re-queue the work unit for re-processing.
		// 	// this blocks if the work queue is full.
		// 	j.jobQ <- j
		// 	// PONDER:  this will lead to a subtle error when shutting down
		// 	// the work unit will be lost, but since this is a simulated
		// 	// worker, there's no problem.

		// 	// if the work has taken more than allow timeout, return a timeout error
		// 	return errors.New("timeout")
	} else {
		fmt.Printf("%v completed with %d.\n", j.id, q)
		// emitRuntimeStats("Work complete")
	}

	return nil
}

func (j *SimulatedJob) OnError(err error) {
	fmt.Println(j.id, "error", err)
}

func simulateJobs(jobQ chan Job) {
	jid := 0
	// load some initial jobs
	fmt.Println("Add 20 jobs.")
	for i := 0; i < 20; i++ {
		jid++
		jobQ <- &SimulatedJob{
			id:      fmt.Sprintf("job-%d", jid),
			jobQ:    jobQ,
			timeout: 5,
		}
	}
	// time.Sleep(10 * time.Second)
	// fmt.Println("Add 10 more jobs")
	// for i := 0; i < 10; i++ {
	// 	jid++
	// 	jobQ <- &SimulatedJob{
	// 		id:      fmt.Sprintf("job-%d", jid),
	// 		jobQ:    jobQ,
	// 		timeout: 5,
	// 	}
	// }
	// time.Sleep(5 * time.Second)
	// fmt.Println("Add 10 more jobs")
	// for i := 0; i < 10; i++ {
	// 	jid++
	// 	jobQ <- &SimulatedJob{
	// 		id:      fmt.Sprintf("job-%d", jid),
	// 		jobQ:    jobQ,
	// 		timeout: 5,
	// 	}
	// }
	// time.Sleep(1 * time.Second)
	// fmt.Println("Add 10 more jobs")
	// for i := 0; i < 10; i++ {
	// 	jid++
	// 	jobQ <- &SimulatedJob{
	// 		id:      fmt.Sprintf("job-%d", jid),
	// 		jobQ:    jobQ,
	// 		timeout: 5,
	// 	}
	// }
	fmt.Println("Final jid:", jid)
}

// ----------------------------------------------------------------------------

func main() {

	fmt.Println("= Num Gorouting:", runtime.NumGoroutine())
	numWorkers := 5
	ctx, cancel := context.WithCancel(context.Background())
	jobQ := make(chan Job, numWorkers)

	go simulateJobs(jobQ)
	fmt.Println("= Num Gorouting:", runtime.NumGoroutine())

	wp := NewWorkerPool(ctx, cancel, numWorkers, jobQ, 60)
	wp.Start()
	fmt.Println("= Num Gorouting:", runtime.NumGoroutine())

	// time.Sleep(1 * time.Second)
	// fmt.Println("===add worker===")
	// if wp.AddWorker() != nil {
	// 	fmt.Println("Error adding worker")
	// }
	// wp.Start()
	// fmt.Println("= Num Gorouting:", runtime.NumGoroutine())

	// time.Sleep(1 * time.Second)
	// fmt.Println("===remore worker 1===")
	// if wp.RemoveWorker() != nil {
	// 	fmt.Println("Error removing worker")
	// }
	// fmt.Println("= Num Gorouting:", runtime.NumGoroutine())

	// time.Sleep(1 * time.Second)
	// fmt.Println("===remove worker 2===")
	// if wp.RemoveWorker() != nil {
	// 	fmt.Println("Error removing worker")
	// }
	// fmt.Println("= Num Gorouting:", runtime.NumGoroutine())
	// go func() {
	// 	ticker := time.NewTicker(30 * time.Second)
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			ticker.Stop()
	// 			return
	// 		case <-ticker.C:
	// 			// output some stats every few seconds
	// 			fmt.Println("= Jobs left:", len(jobQ))
	// 			fmt.Println("= Workers left:", len(wp.workers))
	// 			fmt.Println("= Num Gorouting:", runtime.NumGoroutine())
	// 		}
	// 	}
	// }()

	// go func() {
	// 	ticker := time.NewTicker(10 * time.Second)
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			ticker.Stop()
	// 			return
	// 		case <-ticker.C:
	// 			// output some stats every few seconds
	// 			fmt.Println("===============tick", runtime.NumGoroutine())
	// 		}
	// 	}
	// }()
	<-wp.stopped
	fmt.Println("===final===")
	fmt.Println("= Jobs left:", len(jobQ))
	fmt.Println("= Num Gorouting:", runtime.NumGoroutine())
}
