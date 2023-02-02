package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"
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
	ctx              context.Context
	jobQ             chan Job
	idealWorkerCount int
	stopped          chan struct{}
	workers          map[string]*Worker
	workerIdNum      int
}

func (wp *WorkerPool) AddWorker() error {
	// TODO: protect so that workers isn't accessed by two goroutines

	wp.idealWorkerCount++
	wp.workerIdNum++
	id := fmt.Sprintf("worker-%d", wp.workerIdNum)
	wp.workers[id] = createWorker(wp.ctx, id, wp.jobQ)
	return nil
}

func (wp *WorkerPool) RemoveWorker() error {
	// TODO: protect so that workers isn't accessed by two goroutines
	wp.idealWorkerCount--
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

// ----------------------------------------------------------------------------
// gracefulShutdown waits for terminating syscalls then signals workers to shutdown
func (wp *WorkerPool) gracefulShutdown(cancel func(), timeout time.Duration) chan struct{} {
	sigShutdown := make(chan struct{})

	go func() {
		defer close(sigShutdown)
		sig := make(chan os.Signal, 1)
		defer close(sig)

		// PONDER: add any other syscalls?
		// SIGHUP - hang up, lost controlling terminal
		// SIGINT - interrupt (ctrl-c)
		// SIGQUIT - quit (ctrl-\)
		// SIGTERM - request to terminate
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
		killsig := <-sig
		switch killsig {
		case syscall.SIGINT:
			fmt.Println("Killed with ctrl-c")
		case syscall.SIGTERM:
			fmt.Println("Killed with request to terminate")
		case syscall.SIGQUIT:
			fmt.Println("Killed with ctrl-\\")
		case syscall.SIGHUP:
			fmt.Println("Killed with hang up")
		}

		// set timeout for the cleanup to be done to prevent system hang
		timeoutSignal := make(chan struct{})
		timeoutFunc := time.AfterFunc(timeout, func() {
			fmt.Printf("Timeout %.1fs have elapsed, force exit\n", timeout.Seconds())
			close(timeoutSignal)
		})

		defer timeoutFunc.Stop()

		// cancel the context
		cancel()
		fmt.Println("Shutdown signalled.")

		// wait for timeout to finish and exit
		<-timeoutSignal

		// remove all workers
		for k := range wp.workers {
			fmt.Println("deleting worker", wp.workers[k].id, wp.workers[k].started)
			delete(wp.workers, k)
		}

		sigShutdown <- struct{}{}
	}()

	return sigShutdown
}

var _ WorkerPoolFunctions = (*WorkerPool)(nil)

func NewWorkerPool(ctx context.Context, cancel func(), workerCount int, jobQ chan Job, shutdownTimeout time.Duration) WorkerPool {

	workerPool := WorkerPool{
		ctx:     ctx,
		jobQ:    jobQ,
		workers: make(map[string]*Worker),
	}
	workerPool.stopped = workerPool.gracefulShutdown(cancel, shutdownTimeout*time.Second)

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
		}
		//TODO: how to restart or add a new worker
		// workerChan <- &worker
		w.Start()
	}()
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
	if !w.started {
		return
	}
	// TODO:  shutdown worker gracefully
	w.started = false
	fmt.Println(w.id, "says they're stopping.")
}
