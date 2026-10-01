package queue

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"
)

// Test the behaviour of the blocking policy under preemption and a context
// cancelation.
func TestQueueBlockedPushCancellation(t *testing.T) {
	q := NewQueue[int](2, 2, BlockPolicy[int]{})

	ctx := t.Context()

	if err := q.Push(ctx, 1, mockch); err != nil {
		t.Fatal(err)
	}
	if err := q.Push(ctx, 2, mockch); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)

	go func() {
		done <- q.Push(ctx, 3, mockch)
	}()

	// Make sure the goroutine actually gets a chance to block.
	runtime.Gosched()

	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}

	case <-time.After(time.Second):
		t.Fatal("blocked Push did not respond to cancellation")
	}
}
