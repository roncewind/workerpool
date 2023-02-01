package main

import (
	"context"
	"fmt"
	"time"
)

// ----------------------------------------------------------------------------
// Job implementation

type SimulatedJob struct {
	id string
}

func (j *SimulatedJob) Execute() error {
	fmt.Println(j.id, "executing")
	return nil
}

func (j *SimulatedJob) OnError(err error) {
	fmt.Println(j.id, "error", err)
}

// ----------------------------------------------------------------------------

func main() {

	numWorkers := 5
	ctx, cancel := context.WithCancel(context.Background())
	jobQ := make(chan Job, numWorkers)

	go func() {
		for i := 0; i < 10; i++ {
			jobQ <- &SimulatedJob{
				id: fmt.Sprintf("job-%d", i),
			}
		}
	}()

	wp := NewWorkerPool(ctx, cancel, numWorkers, jobQ)
	wp.Start()

	time.Sleep(1 * time.Second)
	fmt.Println("===")
	if wp.AddWorker() != nil {
		fmt.Println("Error adding worker")
	}
	wp.Start()

	time.Sleep(1 * time.Second)
	fmt.Println("===")
	if wp.RemoveWorker() != nil {
		fmt.Println("Error removing worker")
	}

	time.Sleep(1 * time.Second)
	fmt.Println("===")
	if wp.RemoveWorker() != nil {
		fmt.Println("Error removing worker")
	}

	time.Sleep(1 * time.Second)
}
