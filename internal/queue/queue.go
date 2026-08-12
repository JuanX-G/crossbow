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
	"context"
	"errors"
	"sync"

	"crossbow/internal/ring"
)

var (
	ErrClosed  = errors.New("queue is closed")
	ErrFull    = errors.New("queue is full")
	ErrDropped = errors.New("message dropped")
)

type MailboxPolicy[T any] interface {
	Enqueue(context.Context, *Queue[T], T) error
}

type Queue[T any] struct {
	mu sync.Mutex

	ring      *ring.RingBuffer[T]
	notify    chan struct{}
	notFullCh chan struct{} // closed and replaced any time space might have freed up; lets waiters select on it alongside ctx.Done()
	closed    bool

	policy MailboxPolicy[T]
}

func NewQueue[T any](initialCap int, maxCap int, policy MailboxPolicy[T]) *Queue[T] {
	if initialCap < 2 {
		initialCap = 2
	}
	q := &Queue[T]{
		ring:      ring.NewRingBuffer[T](initialCap, maxCap),
		notify:    make(chan struct{}, 1),
		notFullCh: make(chan struct{}),
		policy:    policy,
	}

	return q
}

// waitForSpace blocks until either space might be available (notFullCh fires),
// the queue closes, or ctx is done. Must be called
// with q.mu held; it releases the lock while waiting and reacquires it before
// returning, so callers can safely re-check state afterwards.
func (q *Queue[T]) waitForSpace(ctx context.Context) error {
	ch := q.notFullCh
	q.mu.Unlock()
	defer q.mu.Lock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
		return nil
	}
}

// wakeWaiters releases anything blocked in waitForSpace. Must be called with q.mu held.
func (q *Queue[T]) wakeWaiters() {
	close(q.notFullCh)
	q.notFullCh = make(chan struct{})
}

func (q *Queue[T]) Push(ctx context.Context, item T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return ErrClosed
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if q.policy == nil {
		return ErrFull
	}

	return q.policy.Enqueue(ctx, q, item)
}

func (q *Queue[T]) Pop() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, ok := q.ring.Pop()
	if !ok {
		return item, false
	}

	q.wakeWaiters()
	return item, true
}

func (q *Queue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.ring.Len()
}

func (q *Queue[T]) Notify() <-chan struct{} {
	return q.notify
}

func (q *Queue[T]) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return
	}

	q.closed = true
	q.wakeWaiters()
	close(q.notify)
}

func (q *Queue[T]) IsClosed() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.closed
}

func (q *Queue[T]) signal() {
	select {
	case q.notify <- struct{}{}:
	default:
	}
}
