package workerpool

import (
	"testing"
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
