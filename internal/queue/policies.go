/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package queue

import "context"

type BlockPolicy[T any] struct{}

func (BlockPolicy[T]) Enqueue(ctx context.Context, q *Queue[T], item T) error {
	for {
		if q.closed {
			return ErrClosed
		}

		if q.ring.Push(item) {
			q.signal()
			return nil
		}

		if err := q.waitForSpace(ctx); err != nil {
			return err
		}
	}
}

type DropNewestPolicy[T any] struct{}

func (DropNewestPolicy[T]) Enqueue(ctx context.Context, q *Queue[T], item T) error {
	if q.closed {
		return ErrClosed
	}

	if !q.ring.Push(item) {
		return ErrDropped
	}

	q.signal()

	return nil
}

type DropOldestPolicy[T any] struct{}

func (DropOldestPolicy[T]) Enqueue(ctx context.Context, q *Queue[T], item T) error {
	if q.closed {
		return ErrClosed
	}

	if !q.ring.Push(item) {
		_, ok := q.ring.Pop()
		if !ok {
			return ErrFull
		}

		if !q.ring.Push(item) {
			return ErrFull
		}
	}

	q.signal()
	return nil
}

type UnboundedPolicy[T any] struct{}

func (UnboundedPolicy[T]) Enqueue(ctx context.Context, q *Queue[T], item T) error {
	if q.closed {
		return ErrClosed
	}

	if !q.ring.Push(item) {
		return ErrFull
	}

	q.signal()

	return nil
}
