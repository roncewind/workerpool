package workerpool

import (
	"context"
	"testing"
	"time"
)

func TestWorkerPool_NewWorkerPool(t *testing.T) {
	jobQ := make(chan Job, 5)

	if _, err := NewWorkerPool(0, jobQ); err != ErrNoWorkers {
		t.Fatalf("expected error when creating worker pool with no workers: %v", err)
	}
	if _, err := NewWorkerPool(-1, jobQ); err != ErrNoWorkers {
		t.Fatalf("expected error when creating worker pool with -1 workers: %v", err)
	}
	if _, err := NewWorkerPool(1, nil); err != ErrNilJobQueue {
		t.Fatalf("expected error when creating worker pool with nil job queue: %v", err)
	}

	if _, err := NewWorkerPool(5, jobQ); err != nil {
		t.Fatalf("expected no error creating worker pool: %v", err)
	}
}

func TestWorkerPool_GetWorkerCount(t *testing.T) {
	numberOfWorkers := 5
	jobQ := make(chan Job, numberOfWorkers)

	wp, err := NewWorkerPool(numberOfWorkers, jobQ)
	if err != nil {
		t.Fatal("error creating worker pool:", err)
	}

	workerCount := wp.GetWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}
}

func TestWorkerPool_ContextCancelRestart(t *testing.T) {
	numberOfWorkers := 1
	jobQ := make(chan Job, numberOfWorkers)
	ctx, cancel := context.WithCancel(context.Background())

	wp, err := NewWorkerPool(numberOfWorkers, jobQ)
	if err != nil {
		t.Fatal("error creating worker pool:", err)
	}

	wp.Start(ctx)
	time.Sleep(1 * time.Second)
	cancel()
	time.Sleep(1 * time.Second)
	workerCount := wp.GetRunningWorkerCount()
	if workerCount != 0 {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, 0)
	}

	ctx, cancel = context.WithCancel(context.Background())
	wp.Start(ctx)
	time.Sleep(1 * time.Second)
	workerCount = wp.GetRunningWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}
	cancel()
}

func TestWorkerPool_MultipleStart(t *testing.T) {
	numberOfWorkers := 5
	jobQ := make(chan Job, numberOfWorkers)
	ctx, cancel := context.WithCancel(context.Background())

	wp, err := NewWorkerPool(numberOfWorkers, jobQ)
	if err != nil {
		t.Fatal("error creating worker pool:", err)
	}

	// Check multiple calls to Start don't cause an issue
	wp.Start(ctx)
	wp.Start(ctx)
	wp.Start(ctx)
	wp.Start(ctx)
	workerCount := wp.GetWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}
	cancel()
}

func TestWorkerPool_AddWorker(t *testing.T) {
	numberOfWorkers := 5
	jobQ := make(chan Job, numberOfWorkers)
	ctx, cancel := context.WithCancel(context.Background())

	wp, err := NewWorkerPool(numberOfWorkers, jobQ)
	if err != nil {
		t.Fatal("error creating worker pool:", err)
	}

	wp.Start(ctx)
	workerCount := wp.GetWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}
	wp.AddWorker()
	numberOfWorkers++
	workerCount = wp.GetWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}
	cancel()
}

func TestWorkerPool_GetRunningWorkerCount(t *testing.T) {
	numberOfWorkers := 5
	jobQ := make(chan Job, numberOfWorkers)
	ctx, cancel := context.WithCancel(context.Background())

	wp, err := NewWorkerPool(numberOfWorkers, jobQ)
	if err != nil {
		t.Fatal("error creating worker pool:", err)
	}

	wp.Start(ctx)
	workerCount := wp.GetWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}

	// Wait a moment for the workers to start up:
	time.Sleep(1 * time.Second)

	workerCount = wp.GetRunningWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}
	cancel()
}

func TestWorkerPool_RemoveWorker(t *testing.T) {
	numberOfWorkers := 1
	jobQ := make(chan Job, numberOfWorkers)
	ctx, cancel := context.WithCancel(context.Background())

	wp, err := NewWorkerPool(numberOfWorkers, jobQ)
	if err != nil {
		t.Fatal("error creating worker pool:", err)
	}

	wp.Start(ctx)
	numberOfRunningWorkers := numberOfWorkers

	// Wait a moment for the workers to start up:
	time.Sleep(1 * time.Second)

	workerCount := wp.GetWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}

	// add a second worker, but don't start it
	wp.AddWorker()
	workerCount = wp.GetRunningWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d running workers, found %d", numberOfRunningWorkers, workerCount)
	}

	//remove the unstarted worker
	wp.RemoveWorker()
	workerCount = wp.GetRunningWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d running workers, found %d", numberOfRunningWorkers, workerCount)
	}

	//remove the only worker
	wp.RemoveWorker()
	numberOfWorkers--
	workerCount = wp.GetWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}

	//test removing a worker from an empty pool, expect an error
	wp.RemoveWorker()
	if err := wp.RemoveWorker(); err != ErrNoWorkersToRemove {
		t.Fatalf("expected error when creating worker pool with no workers: %v", err)
	}
	cancel()
}

// ----------------------------------------------------------------------------
// Trivial Job implementation

type TrivialJob struct {
}

// ----------------------------------------------------------------------------
// make sure TrivialJob implements the Job interface

var _ Job = (*TrivialJob)(nil)

// ----------------------------------------------------------------------------
// Job implementation

func (j *TrivialJob) Execute() error {
	myWork := 50
	// simulate doing some work... for "myWork" number of Milliseconds
	time.Sleep(time.Duration(myWork) * time.Millisecond)
	return nil
}

func (j *TrivialJob) OnError(err error) {
}

func TestWorkerPool_TrivialJobExecution(t *testing.T) {
	numberOfWorkers := 5
	jobQ := make(chan Job, numberOfWorkers)
	ctx, cancel := context.WithCancel(context.Background())
	jobQ <- &TrivialJob{}

	wp, err := NewWorkerPool(numberOfWorkers, jobQ)
	if err != nil {
		t.Fatal("error creating worker pool:", err)
	}

	wp.Start(ctx)
	workerCount := wp.GetWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}
	// Wait a moment for the workers to execute the job:
	time.Sleep(1 * time.Second)
	cancel()
}

// ----------------------------------------------------------------------------
// Panic Job implementation

type PanicJob struct {
}

// ----------------------------------------------------------------------------
// make sure PanicJob implements the Job interface

var _ Job = (*PanicJob)(nil)

// ----------------------------------------------------------------------------
// Job implementation

func (j *PanicJob) Execute() error {
	panic("panic job")
}

func (j *PanicJob) OnError(err error) {
}

func TestWorkerPool_PanicJobExecution(t *testing.T) {
	numberOfWorkers := 5
	jobQ := make(chan Job, numberOfWorkers)
	ctx, cancel := context.WithCancel(context.Background())
	jobQ <- &PanicJob{}

	wp, err := NewWorkerPool(numberOfWorkers, jobQ)
	if err != nil {
		t.Fatal("error creating worker pool:", err)
	}

	wp.Start(ctx)
	workerCount := wp.GetWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}
	// Wait a moment for the workers to execute the job:
	time.Sleep(1 * time.Second)
	cancel()
}
