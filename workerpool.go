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
	lock       sync.Mutex
	quit       chan struct{}
	running    bool
}

var ErrNoWorkers = fmt.Errorf("attempted to create a WorkerPool with no workers")
var ErrNilJobQueue = fmt.Errorf("attempted to create a WorkerPool without a job queue")
var ErrNoWorkersToRemove = fmt.Errorf("all workers removed, none to remove")

// ----------------------------------------------------------------------------

// Create a new WorkerPool with the given number of workers pulling jobs from
// the specified job queue
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

// Add a worker to the pool, the worker can then be started with
// `workerPool.Start(ctx)`.
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

// Get the current count of workers in the pool, they may not all be running.
// To start workers that are not running use the Start(ctx) method
func (wp *WorkerPoolImpl) GetWorkerCount() int {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	return len(wp.workers)
}

// ----------------------------------------------------------------------------

// Get the current count of workers that are running.
func (wp *WorkerPoolImpl) GetRunningWorkerCount() int {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	count := 0
	for _, worker := range wp.workers {
		// fmt.Println("running:", worker.id, worker.running)
		if worker.getRunning() {
			count++
		}
	}
	return count
}

// ----------------------------------------------------------------------------

// Remove a worker from the pool, this will remove non-running workers first
// followed by oldest workers next.
func (wp *WorkerPoolImpl) RemoveWorker() error {
	wp.idealWorkerCount--
	wp.lock.Lock()
	defer wp.lock.Unlock()
	workerCount := len(wp.workers)
	if workerCount <= 0 {
		return ErrNoWorkersToRemove
	}

	keys := make([]string, 0, workerCount)
	stoppedId := ""
	for k := range wp.workers {
		keys = append(keys, k)
		if !wp.workers[k].getRunning() {
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
		// fmt.Println(worker.id)
		if !worker.getRunning() {
			// fmt.Println("Starting ", worker.id)
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

// Check at compile time that the implementation adheres to the interface.
var _ WorkerPool = (*WorkerPoolImpl)(nil)

// ----------------------------------------------------------------------------

// Start the worker.  When this method exits it will call Stop()
func (w *Worker) Start(ctx context.Context) {

	defer func() {
		// make sure the worker is stopped when leaving this function
		w.Stop()
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				w.currentJob.OnError(err)
			} else if str, ok := r.(string); ok {
				w.currentJob.OnError(errors.New(str))
			} else {
				w.currentJob.OnError(errors.New("unknown error"))
			}

			// restart this worker after panic
			w.Start(ctx)
		}
	}()
	w.setRunning(true)
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

// Stop the workers, this only sets running to false.  For this working to quit,
// issue: `close(worker.quit)`
func (w *Worker) Stop() {
	if !w.getRunning() {
		return
	}
	// TODO:  shutdown worker gracefully?? is there anything else to do?
	w.setRunning(false)
}

// ----------------------------------------------------------------------------

// Thread safe setter for if this worker is running
func (w *Worker) setRunning(b bool) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.running = b
}

// ----------------------------------------------------------------------------

// Thread safe setter for if this worker is running
func (w *Worker) getRunning() bool {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.running
}
