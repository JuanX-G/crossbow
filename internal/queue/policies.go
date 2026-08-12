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
