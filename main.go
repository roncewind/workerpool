package workerpool

import (
	"context"
	"fmt"
)

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

type Job interface {
	Execute(context.Context) error
	OnError(error)
}

type WorkerPool interface {
	AddWorker() error
	GetRunningWorkerCount() int
	GetWorkerCount() int
	RemoveWorker() error
	Start(context.Context) error
}

// ----------------------------------------------------------------------------
// Predefined errors
// ----------------------------------------------------------------------------

var ErrNoWorkers = fmt.Errorf("attempted to create a WorkerPool with no workers")
var ErrNilJobQueue = fmt.Errorf("attempted to create a WorkerPool without a job queue")
var ErrNoWorkersToRemove = fmt.Errorf("all workers removed, none to remove")
var ErrPanic = fmt.Errorf("job panicked")
