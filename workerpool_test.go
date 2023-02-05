package workerpool

import (
	"context"
	"errors"
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
	if workerCount != 0 {
		t.Fatalf("expected %d workers, found %d", 0, workerCount)
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
	if workerCount != 0 {
		t.Fatalf("expected %d workers, found %d", 0, workerCount)
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

	// Wait a moment for the workers to start up:
	time.Sleep(1 * time.Second)

	workerCount := wp.GetWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}

	// add a second worker, but don't start it
	wp.AddWorker()
	workerCount = wp.GetRunningWorkerCount()
	if workerCount != 0 {
		t.Fatalf("expected %d running workers, found %d", 0, workerCount)
	}

	//remove the unstarted worker
	wp.RemoveWorker()
	workerCount = wp.GetRunningWorkerCount()
	if workerCount != 0 {
		t.Fatalf("expected %d running workers, found %d", 0, workerCount)
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
	err interface{}
}

// ----------------------------------------------------------------------------
// make sure PanicJob implements the Job interface

var _ Job = (*PanicJob)(nil)

// ----------------------------------------------------------------------------
// Job implementation

func (j *PanicJob) Execute() error {
	panic(j.err)
}

func (j *PanicJob) OnError(err error) {
}

func TestWorkerPool_PanicJobExecution(t *testing.T) {
	numberOfWorkers := 5
	jobQ := make(chan Job, numberOfWorkers)
	ctx, cancel := context.WithCancel(context.Background())
	jobQ <- &PanicJob{
		err: errors.New("panic error type"),
	}
	jobQ <- &PanicJob{
		err: "panic string type",
	}
	jobQ <- &PanicJob{
		err: 100,
	}

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
// Error Job implementation

type ErrorJob struct {
}

// ----------------------------------------------------------------------------
// make sure ErrorJob implements the Job interface

var _ Job = (*ErrorJob)(nil)

// ----------------------------------------------------------------------------
// Job implementation

func (j *ErrorJob) Execute() error {
	return errors.New("error")
}

func (j *ErrorJob) OnError(err error) {
}

func TestWorkerPool_ErrorJobExecution(t *testing.T) {
	numberOfWorkers := 5
	jobQ := make(chan Job, numberOfWorkers)
	ctx, cancel := context.WithCancel(context.Background())
	jobQ <- &ErrorJob{}

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
// Log running Job implementation

type LongJob struct {
	seconds int
}

// ----------------------------------------------------------------------------
// make sure ErrorJob implements the Job interface

var _ Job = (*LongJob)(nil)

// ----------------------------------------------------------------------------
// Job implementation

func (j *LongJob) Execute() error {
	time.Sleep(time.Duration(j.seconds) * time.Second)
	return nil
}

func (j *LongJob) OnError(err error) {
}

func TestWorkerPool_LongJobExecution(t *testing.T) {
	numberOfWorkers := 1
	jobQ := make(chan Job, numberOfWorkers)
	ctx, cancel := context.WithCancel(context.Background())
	jobQ <- &LongJob{seconds: 3}

	wp, err := NewWorkerPool(numberOfWorkers, jobQ)
	if err != nil {
		t.Fatal("error creating worker pool:", err)
	}

	wp.Start(ctx)
	time.Sleep(1 * time.Second)
	workerCount := wp.GetRunningWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}
	// Wait a moment for the workers to execute the job:
	time.Sleep(1 * time.Second)
	cancel()
}

func TestWorkerPool_RemoveRunningWorker(t *testing.T) {
	numberOfWorkers := 1
	jobQ := make(chan Job, numberOfWorkers)
	ctx, cancel := context.WithCancel(context.Background())
	jobQ <- &LongJob{seconds: 3}

	wp, err := NewWorkerPool(numberOfWorkers, jobQ)
	if err != nil {
		t.Fatal("error creating worker pool:", err)
	}

	wp.Start(ctx)
	time.Sleep(1 * time.Second)
	workerCount := wp.GetRunningWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}
	//remove the only worker
	wp.RemoveWorker()
	numberOfWorkers--
	workerCount = wp.GetWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}
	cancel()
}
