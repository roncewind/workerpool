package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// ----------------------------------------------------------------------------
// Job implementation

type SimulatedJob struct {
	id string
}

func (j *SimulatedJob) Execute() error {
	fmt.Println(j.id, "executing")
	time.Sleep(2 * time.Second)
	return nil
}

func (j *SimulatedJob) OnError(err error) {
	fmt.Println(j.id, "error", err)
}

// ----------------------------------------------------------------------------

func main() {

	fmt.Println("= Num Gorouting:", runtime.NumGoroutine())
	numWorkers := 5
	ctx, cancel := context.WithCancel(context.Background())
	jobQ := make(chan Job, numWorkers)

	go func() {
		for i := 0; i < 20; i++ {
			jobQ <- &SimulatedJob{
				id: fmt.Sprintf("job-%d", i),
			}
		}
	}()
	fmt.Println("= Num Gorouting:", runtime.NumGoroutine())

	wp := NewWorkerPool(ctx, cancel, numWorkers, jobQ)
	wp.Start()
	fmt.Println("= Num Gorouting:", runtime.NumGoroutine())

	time.Sleep(1 * time.Second)
	fmt.Println("===add worker===")
	if wp.AddWorker() != nil {
		fmt.Println("Error adding worker")
	}
	wp.Start()
	fmt.Println("= Num Gorouting:", runtime.NumGoroutine())

	time.Sleep(1 * time.Second)
	fmt.Println("===remore worker 1===")
	if wp.RemoveWorker() != nil {
		fmt.Println("Error removing worker")
	}
	fmt.Println("= Num Gorouting:", runtime.NumGoroutine())

	time.Sleep(1 * time.Second)
	fmt.Println("===remove worker 2===")
	if wp.RemoveWorker() != nil {
		fmt.Println("Error removing worker")
	}
	fmt.Println("= Num Gorouting:", runtime.NumGoroutine())

	time.Sleep(5 * time.Second)
	fmt.Println("===cancel===")
	cancel()
	time.Sleep(5 * time.Second)
	fmt.Println("= Num Gorouting:", runtime.NumGoroutine())
}
