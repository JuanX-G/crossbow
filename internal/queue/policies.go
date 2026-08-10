package queue

type BlockPolicy[T any] struct{}

func (BlockPolicy[T]) Enqueue(q *Queue[T], item T) error {
	for {
		if q.closed {
			return ErrClosed
		}

		if q.ring.Push(item) {
			q.signal()
			return nil
		}

		q.notFull.Wait()
	}
}

type DropNewestPolicy[T any] struct{}

func (DropNewestPolicy[T]) Enqueue(q *Queue[T], item T) error {
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

func (DropOldestPolicy[T]) Enqueue(q *Queue[T], item T) error {
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

func (UnboundedPolicy[T]) Enqueue(q *Queue[T], item T) error {
	if q.closed {
		return ErrClosed
	}

	if !q.ring.Push(item) {
		return ErrFull
	}

	q.signal()

	return nil
}
