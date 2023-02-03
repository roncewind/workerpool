package workerpool

import (
	"context"
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
	id      string
	jobQ    <-chan Job
	quit    chan struct{}
	running bool
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
	fmt.Println("Remove Worker before count", len(wp.workers))
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
	fmt.Println("Remove Worker", id, "worker count:", len(wp.workers))
	return nil
}

// ----------------------------------------------------------------------------

func (wp *WorkerPoolImpl) startAllWorkers(ctx context.Context) error {

	wp.lock.Lock()
	defer wp.lock.Unlock()
	for id, worker := range wp.workers {
		if !worker.running {
			fmt.Println("Starting", id)
			worker.Start(ctx)
		}
	}
	return nil
}

// ----------------------------------------------------------------------------

func (wp *WorkerPoolImpl) Start(ctx context.Context) error {

	wp.lock.Lock()
	defer wp.lock.Unlock()
	for id, worker := range wp.workers {
		if !worker.running {
			fmt.Println("Starting new", id)
			go worker.Start(ctx)
		}
	}
	return nil
}

// ----------------------------------------------------------------------------

func (wp *WorkerPoolImpl) createWorker(id string) *Worker {
	fmt.Println("Creating", id)
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
			w.Start(ctx)
		}
	}()
	w.running = true
	fmt.Println(w.id, "says they're running.")
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.quit:
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
