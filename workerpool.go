package main

import (
	"context"
	"fmt"
	"sort"
)

// ----------------------------------------------------------------------------

type Job interface {
	Execute() error
	OnError(error)
}

type Worker struct {
	ctx     context.Context
	id      string
	jobQ    <-chan Job
	quit    chan struct{}
	started bool
}

type WorkerPoolFunctions interface {
	AddWorker() error
	RemoveWorker() error
	Start() error
}

type WorkerPool struct {
	ctx         context.Context
	jobQ        chan Job
	workers     map[string]*Worker
	workerIdNum int
}

func (wp *WorkerPool) AddWorker() error {
	// TODO: protect so that workers isn't accessed by two goroutines

	wp.workerIdNum++
	id := fmt.Sprintf("worker-%d", wp.workerIdNum)
	wp.workers[id] = createWorker(wp.ctx, id, wp.jobQ)
	return nil
}

func (wp *WorkerPool) RemoveWorker() error {
	// TODO: protect so that workers isn't accessed by two goroutines
	workerCount := len(wp.workers)
	fmt.Println("count", len(wp.workers))
	if workerCount <= 0 {
		return nil
	}
	// PONDER:  always kill the oldest worker?
	keys := make([]string, 0, workerCount)

	for k := range wp.workers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	id := keys[0]
	close(wp.workers[id].quit)
	delete(wp.workers, id)
	fmt.Println("count", len(wp.workers))
	return nil
}

func (wp *WorkerPool) Start() error {
	for id, worker := range wp.workers {
		if !worker.started {
			fmt.Println("Starting", id)
			go worker.Start()
		}
	}
	return nil
}

var _ WorkerPoolFunctions = (*WorkerPool)(nil)

func NewWorkerPool(ctx context.Context, cancel func(), workerCount int, jobQ chan Job) WorkerPool {

	workerPool := WorkerPool{
		ctx:     ctx,
		jobQ:    jobQ,
		workers: make(map[string]*Worker),
	}

	// create a number of workers.
	for i := 0; i < workerCount; i++ {
		workerPool.AddWorker()
	}
	return workerPool
}

func createWorker(ctx context.Context, id string, jobQ chan Job) *Worker {
	fmt.Println("Creating", id)
	return &Worker{
		ctx:     ctx,
		id:      id,
		jobQ:    jobQ,
		quit:    make(chan struct{}),
		started: false,
	}
}

func (w *Worker) Start() {
	w.started = true
	fmt.Println(w.id, "says they're starting.")
	for {
		// use select to test if our context has completed
		select {
		case <-w.ctx.Done():
			w.Stop()
			return
		case <-w.quit:
			w.Stop()
			return
		case j := <-w.jobQ:
			fmt.Printf("%s:", w.id)
			err := j.Execute()
			if err != nil {
				j.OnError(err)
			}
		}
	}
}

func (w *Worker) Stop() {
	// TODO:  shutdown worker gracefully
	w.started = false
	fmt.Println(w.id, "says they're stopping.")
}
