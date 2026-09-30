// Copyright 2026 Maciej "juan_em (JuanX-G)" Woźniak
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Or use the bundled copy in the LICENSE file.

package queue

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

var mockch = make(chan error, 10000)

// Test push-pop sequence correctness
func TestPushPop(t *testing.T) {
	ctx := t.Context()
	q := NewQueue(2, 4, BlockPolicy[int]{})
	q.Push(ctx, 1, mockch)
	q.Push(ctx, 2, mockch)

	v, ok := q.Pop()
	if !ok {
		t.Fatalf("pop returned false but value inside were expected. Value returned: %d", v)
	}
	if v != 1 {
		t.Fatalf("pop returned wrong value. expected: 1, found: %d", v)
	}
}

// Test sending the notifications of new items.
func TestNotify(t *testing.T) {
	ctx := t.Context()
	q := NewQueue(2, 6, BlockPolicy[int]{})
	q.Push(ctx, 42, mockch)
	select {
	case <-q.Notify():
	default:
		t.Fatalf("pushed to queue but did not recive a notification")
	}
}

// Test the queue's behaviour on close
func TestClose(t *testing.T) {
	ctx := t.Context()
	q := NewQueue(2, 6, BlockPolicy[int]{})
	q.Push(ctx, 42, mockch)
	<-q.Notify()
	q.Close()
	_, ok := <-q.Notify()
	if ok {
		t.Fatalf("called close on queue but the notification channel remained open even after draining it")
	}
	if !q.IsClosed() {
		t.Fatalf("called close on queue but IsClosed() returned 'false'")
	}
}

// Test that the blocking policy actually blocks
func TestBlockPolicy(t *testing.T) {
	ctx := t.Context()
	q := NewQueue(2, 2, BlockPolicy[int]{})
	q.Push(ctx, 1, mockch)
	q.Push(ctx, 2, mockch)
	doneCh := make(chan struct{})
	pushFn := func() {
		q.Push(ctx, 3, mockch)
		doneCh <- struct{}{}
	}
	go pushFn()

	select {
	case <-doneCh:
		t.Fatalf("Should have been blocked")
	default:
	}

	q.Pop()
	time.Sleep(time.Millisecond * 25)
	select {
	case <-doneCh:
	default:
		t.Fatalf("Should have unblocked now")
	}

}

type policyTest struct {
	input    []int
	endState []int
}

// Test that drop policies drop correct items.
func TestPolicies(t *testing.T) {
	tests := map[MailboxPolicy[int]]policyTest{
		DropNewestPolicy[int]{}: policyTest{input: []int{1, 2, 3}, endState: []int{1, 2}},
		DropOldestPolicy[int]{}: policyTest{input: []int{1, 2, 3}, endState: []int{2, 3}},
	}
	ctx := t.Context()
	for policy, test := range tests {
		q := NewQueue(2, 2, policy)
		for _, v := range test.input {
			q.Push(ctx, v, mockch)
		}
		for _, vs := range test.endState {
			if vp, _ := q.Pop(); vp != vs {
				t.Fatalf("Queue with policy %#v returned a wrong value; expcted: %d, found: %d", policy, vs, vp)
			}
		}
	}
}

func FuzzQueue(f *testing.F) {
	q := NewQueue(1, 4, DropNewestPolicy[int]{})
	f.Add(1)
	f.Fuzz(func(t *testing.T, in int) {
		err := q.Push(t.Context(), in, mockch)
		if err != nil {
			panic("1")
		}
		if err != nil && !errors.Is(err, ErrDropped) {
			t.Errorf("Expected ErrDropped, found: %#v", err)
		}
	})
}

// Test concurrent popping and pushing to the queue
func TestQueueConcurrentPushPop(t *testing.T) {
	const (
		PRODUCERS          = 32
		ITEMS_PER_PRODUCER = 1000
	)

	q := NewQueue[int](16, 128, BlockPolicy[int]{})

	var wg sync.WaitGroup
	wg.Add(PRODUCERS)

	for p := 0; p < PRODUCERS; p++ {
		p := p
		go func() {
			defer wg.Done()

			for i := 0; i < ITEMS_PER_PRODUCER; i++ {
				v := p*ITEMS_PER_PRODUCER + i

				if err := q.Push(t.Context(), v, mockch); err != nil {
					t.Errorf("Push(%d): %v", v, err)
					return
				}
			}
		}()
	}

	seen := make(map[int]bool)
	var mu sync.Mutex

	done := make(chan struct{})

	go func() {
		defer close(done)

		expected := PRODUCERS * ITEMS_PER_PRODUCER

		for len(seen) < expected {
			v, ok := q.Pop()
			if !ok {
				runtime.Gosched()
				continue
			}

			mu.Lock()
			if seen[v] {
				t.Errorf("value %d popped twice", v)
			}
			seen[v] = true
			mu.Unlock()
		}
	}()

	wg.Wait()
	<-done

	if len(seen) != PRODUCERS*ITEMS_PER_PRODUCER {
		t.Fatalf("got %d items, expected %d",
			len(seen), PRODUCERS*ITEMS_PER_PRODUCER)
	}
}

// Test the queue's behaviour when unblocking multiples goroutines waiting to
// push.
func TestQueueConcurrentBlockedPushers(t *testing.T) {
	const pushers = 100

	q := NewQueue[int](2, 2, BlockPolicy[int]{})

	if err := q.Push(t.Context(), 1, mockch); err != nil {
		t.Fatal(err)
	}
	if err := q.Push(t.Context(), 2, mockch); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(pushers)

	errCh := make(chan error, pushers)

	for i := 0; i < pushers; i++ {
		go func(i int) {
			defer wg.Done()

			if err := q.Push(t.Context(), i+1000, mockch); err != nil {
				errCh <- fmt.Errorf("pusher %d: %w", i, err)
			}
		}(i)
	}

	for popped := 0; popped < pushers; {
		_, ok := q.Pop()
		if !ok {
			runtime.Gosched()
			continue
		}

		popped++
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(errCh)

		for err := range errCh {
			t.Error(err)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("blocked pushers did not all make progress")
	}
}

// Test whether the queue is safe under two routines competing for either
// sending or closing.
func TestQueuePushCloseRace(t *testing.T) {
	for iteration := 0; iteration < 1000; iteration++ {
		q := NewQueue[int](2, 2, BlockPolicy[int]{})

		start := make(chan struct{})

		var wg sync.WaitGroup

		wg.Add(2)

		go func() {
			defer wg.Done()
			<-start
			_ = q.Push(t.Context(), 42, mockch)
		}()

		go func() {
			defer wg.Done()
			<-start
			q.Close()
		}()

		close(start)
		wg.Wait()

		if !q.IsClosed() {
			t.Fatal("queue should be closed")
		}
	}
}
