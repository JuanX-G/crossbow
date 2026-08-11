/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package queue

import (
	"testing"
	"time"
)

func TestPushPop(t *testing.T) {
	ctx := t.Context()
	q := NewQueue(2, 4, BlockPolicy[int]{})
	q.Push(ctx, 1)
	q.Push(ctx, 2)

	v, ok := q.Pop()
	if !ok {
		t.Fatalf("pop returned false but value inside were expected. Value returned: %d", v)
	}
	if v != 1 {
		t.Fatalf("pop returned wrong value. expected: 1, found: %d", v)
	}
}

func TestNotify(t *testing.T) {
	ctx := t.Context()
	q := NewQueue(2, 6, BlockPolicy[int]{})
	q.Push(ctx, 42)
	select {
	case <-q.Notify():
	default:
		t.Fatalf("pushed to queue but did not recive a notification")
	}
}

func TestClose(t *testing.T) {
	ctx := t.Context()
	q := NewQueue[int](2, 6, BlockPolicy[int]{})
	q.Push(ctx, 42)
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

func TestBlockPolicy(t *testing.T) {
	ctx := t.Context()
	q := NewQueue(2, 2, BlockPolicy[int]{})
	q.Push(ctx, 1)
	q.Push(ctx, 2)
	doneCh := make(chan struct{})
	pushFn := func() {
		q.Push(ctx, 3)
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

func TestPolicies(t *testing.T) {
	tests := map[MailboxPolicy[int]]policyTest{
		DropNewestPolicy[int]{}: policyTest{input: []int{1, 2, 3}, endState: []int{1, 2}},
		DropOldestPolicy[int]{}: policyTest{input: []int{1, 2, 3}, endState: []int{2, 3}},
	}
	ctx := t.Context()
	for policy, test := range tests {
		q := NewQueue(2, 2, policy)
		for _, v := range test.input {
			q.Push(ctx, v)
		}
		for _, vs := range test.endState {
			if vp, _ := q.Pop(); vp != vs {
				t.Fatalf("Queue with policy %#v returned a wrong value; expcted: %d, found: %d", policy, vs, vp)
			}
		}
	}
}
