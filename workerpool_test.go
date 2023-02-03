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
	wp.RemoveWorker()
	numberOfWorkers--
	workerCount = wp.GetWorkerCount()
	if numberOfWorkers != workerCount {
		t.Fatalf("expected %d workers, found %d", numberOfWorkers, workerCount)
	}
	cancel()

}
