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
	id      string
	jobQ    <-chan Job
	quit    chan struct{}
	running bool
}

type WorkerPoolFunctions interface {
	AddWorker() error
	GetWorkerCount() int
	RemoveWorker() error
	Start(context.Context) error
}

type WorkerPool struct {
	jobQ             chan Job
	idealWorkerCount int
	// stopped          chan struct{}
	workers     map[string]*Worker
	workerIdNum int
}

// ----------------------------------------------------------------------------

func (wp *WorkerPool) AddWorker() error {
	// TODO: protect so that workers isn't accessed by two goroutines

	wp.idealWorkerCount++
	wp.workerIdNum++
	id := fmt.Sprintf("worker-%d", wp.workerIdNum)
	wp.workers[id] = wp.createWorker(id)
	return nil
}

// ----------------------------------------------------------------------------

func (wp *WorkerPool) GetWorkerCount() int {
	return len(wp.workers)
}

// ----------------------------------------------------------------------------

func (wp *WorkerPool) RemoveWorker() error {
	// TODO: protect so that workers isn't accessed by two goroutines
	wp.idealWorkerCount--
	workerCount := len(wp.workers)
	fmt.Println("Remove Worker before count", len(wp.workers))
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
	wp.cleanup(wp.workers[id])
	// wp.workers[id].Stop()
	// close(wp.workers[id].quit)
	// delete(wp.workers, id)
	fmt.Println("Remove Worker after count", len(wp.workers))
	return nil
}

// ----------------------------------------------------------------------------

func (wp *WorkerPool) startAllWorkers(ctx context.Context) error {

	for id, worker := range wp.workers {
		if !worker.running {
			fmt.Println("Starting", id)
			worker.Start(ctx)
		}
	}
	return nil
}

// ----------------------------------------------------------------------------

func (wp *WorkerPool) cleanup(worker *Worker) {
	fmt.Println("cleanup", worker.id)
	worker.Stop()
	close(worker.quit)
	delete(wp.workers, worker.id)
}

// ----------------------------------------------------------------------------

func (wp *WorkerPool) Start(ctx context.Context) error {

	for id, worker := range wp.workers {
		if !worker.running {
			fmt.Println("Starting new", id)
			go worker.Start(ctx)
		}
	}
	return nil
}

// ----------------------------------------------------------------------------

func (wp *WorkerPool) createWorker(id string) *Worker {
	fmt.Println("Creating", id)
	return &Worker{
		id:      id,
		jobQ:    wp.jobQ,
		quit:    make(chan struct{}),
		running: false,
	}
}

// ----------------------------------------------------------------------------

var _ WorkerPoolFunctions = (*WorkerPool)(nil)

// ----------------------------------------------------------------------------

func NewWorkerPool(ctx context.Context, workerCount int, jobQ chan Job) WorkerPool {

	workerPool := WorkerPool{
		jobQ:    jobQ,
		workers: make(map[string]*Worker),
	}
	// workerPool.stopped = workerPool.gracefulShutdown(cancel, shutdownTimeout*time.Second)

	// create a number of workers.
	for i := 0; i < workerCount; i++ {
		workerPool.AddWorker()
	}
	return workerPool
}

// ----------------------------------------------------------------------------

func (w *Worker) Start(ctx context.Context) {
	// make the goroutine signal its death, whether it's a panic or a return
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				fmt.Println(w.id, "recover", err)
				// w.setError(err)
			} else {
				fmt.Println(w.id, "recover panic", r)
				// w.setError(fmt.Errorf("panic happened %v", r))
			}
			// } else {
			// a little tricky go code here.
			//  err is picked up from the doWork return
			// w.setError(err)

			// restart this worker after panic
			fmt.Println("Stop and re-start worker after panic")
			w.Stop()
			w.Start(ctx)
		}
	}()
	w.running = true
	fmt.Println(w.id, "says they're running.")
	for {
		// use select to test if our context has completed
		select {
		case <-ctx.Done():
			// stop when all work is cancelled
			w.Stop()
			return
		case <-w.quit:
			// stop when this worker is asked to quit
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

// ----------------------------------------------------------------------------

func (w *Worker) Stop() {
	if !w.running {
		return
	}
	// TODO:  shutdown worker gracefully
	w.running = false
	fmt.Println(w.id, "says they're stopping.")
}
