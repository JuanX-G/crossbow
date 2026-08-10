package queue

import (
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
	Enqueue(*Queue[T], T) error
}

type Queue[T any] struct {
	mu      sync.Mutex
	notFull *sync.Cond

	ring   *ring.RingBuffer[T]
	notify chan struct{}
	closed bool

	policy MailboxPolicy[T]
}

func NewQueue[T any](initialCap int, maxCap int, policy MailboxPolicy[T]) *Queue[T] {
	if initialCap < 2 {
		initialCap = 2
	}
	q := &Queue[T]{
		ring:   ring.NewRingBuffer[T](initialCap, maxCap),
		notify: make(chan struct{}, 1),
		policy: policy,
	}

	q.notFull = sync.NewCond(&q.mu)
	return q
}

func (q *Queue[T]) Push(item T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return ErrClosed
	}

	if q.policy == nil {
		return ErrFull
	}

	return q.policy.Enqueue(q, item)
}

func (q *Queue[T]) Pop() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	item, ok := q.ring.Pop()
	if !ok {
		return item, false
	}

	q.notFull.Signal()
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
	q.notFull.Broadcast()
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
