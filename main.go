package workerpool

import "context"

// ----------------------------------------------------------------------------
// Types
// ----------------------------------------------------------------------------

type Job interface {
	Execute() error
	OnError(error)
}

type WorkerPool interface {
	AddWorker() error
	GetRunningWorkerCount() int
	GetWorkerCount() int
	RemoveWorker() error
	Start(context.Context) error
}
