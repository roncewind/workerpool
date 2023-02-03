package workerpool

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

// ----------------------------------------------------------------------------
// Internal Type
// ----------------------------------------------------------------------------

type WorkerPoolImpl struct {
	jobQ             chan Job
	idealWorkerCount int
	lock             sync.Mutex
	workers          map[string]*Worker
	workerIdNum      int
}

type Worker struct {
	currentJob Job
	id         string
	jobQ       <-chan Job
	quit       chan struct{}
	running    bool
}

var ErrNoWorkers = fmt.Errorf("attempted to create a WorkerPool with no workers")
var ErrNilJobQueue = fmt.Errorf("attempted to create a WorkerPool without a job queue")

// ----------------------------------------------------------------------------

// Create a new WorkerPool
func NewWorkerPool(workerCount int, jobQ chan Job) (WorkerPool, error) {

	if workerCount < 1 {
		return nil, ErrNoWorkers
	}

	if jobQ == nil {
		return nil, ErrNilJobQueue
	}

	workerPool := &WorkerPoolImpl{
		jobQ:    jobQ,
		workers: make(map[string]*Worker),
	}

	// create a number of workers.
	for i := 0; i < workerCount; i++ {
		workerPool.AddWorker()
	}
	return workerPool, nil
}

// ----------------------------------------------------------------------------

func (wp *WorkerPoolImpl) AddWorker() error {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	wp.idealWorkerCount++
	wp.workerIdNum++
	id := fmt.Sprintf("worker-%d", wp.workerIdNum)
	worker := wp.createWorker(id)
	wp.workers[id] = worker
	return nil
}

// ----------------------------------------------------------------------------

func (wp *WorkerPoolImpl) GetWorkerCount() int {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	return len(wp.workers)
}

// ----------------------------------------------------------------------------

func (wp *WorkerPoolImpl) GetRunningWorkerCount() int {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	count := 0
	for _, worker := range wp.workers {
		if worker.running {
			count++
		}
	}
	return count
}

// ----------------------------------------------------------------------------

func (wp *WorkerPoolImpl) RemoveWorker() error {
	wp.idealWorkerCount--
	wp.lock.Lock()
	defer wp.lock.Unlock()
	workerCount := len(wp.workers)
	if workerCount <= 0 {
		return nil
	}

	keys := make([]string, 0, workerCount)
	stoppedId := ""
	for k := range wp.workers {
		keys = append(keys, k)
		if !wp.workers[k].running {
			stoppedId = k
			break
		}
	}
	var id string
	if stoppedId != "" {
		id = stoppedId
	} else {
		sort.Strings(keys)
		id = keys[0]
	}
	wp.workers[id].Stop()
	close(wp.workers[id].quit)
	delete(wp.workers, id)
	return nil
}

// ----------------------------------------------------------------------------

// Method starts all the defined workers in the WorkerPool
func (wp *WorkerPoolImpl) Start(ctx context.Context) error {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	for _, worker := range wp.workers {
		if !worker.running {
			go worker.Start(ctx)
		}
	}
	return nil
}

// ----------------------------------------------------------------------------

// internal method for creating workers.
func (wp *WorkerPoolImpl) createWorker(id string) *Worker {
	return &Worker{
		id:      id,
		jobQ:    wp.jobQ,
		quit:    make(chan struct{}),
		running: false,
	}
}

// ----------------------------------------------------------------------------

var _ WorkerPool = (*WorkerPoolImpl)(nil)

// ----------------------------------------------------------------------------

func (w *Worker) Start(ctx context.Context) {

	defer func() {
		// make sure the worker is stopped when leaving this function
		w.Stop()
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				w.currentJob.OnError(err)
			} else {
				w.currentJob.OnError(errors.New("job failed with panic"))
			}

			// restart this worker after panic
			w.Start(ctx)
		}
	}()
	w.running = true
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.quit:
			return
		case w.currentJob = <-w.jobQ:
			err := w.currentJob.Execute()
			if err != nil {
				w.currentJob.OnError(err)
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
}
