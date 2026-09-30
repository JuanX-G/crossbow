package eventstream

import (
	"context"
	"sync/atomic"

	"github.com/JuanX-G/crossbow/internal/queue"

	"github.com/JuanX-G/crossbow"
)

type Subscription[E any] struct {
	mailbox     *queue.Queue[E]
	canceled    atomic.Bool
	closeNotify chan struct{}
	eventNotify chan struct{}
}

func NewSubscription[E any](maxCap int, policy crossbow.MailboxPolicy) (*Subscription[E], error) {
	if maxCap < 1 {
		return nil, ErrEventStreamClosed // TODO: fix
	}

	cap := (maxCap + 1) / 2

	subscription := &Subscription[E]{
		mailbox:     queue.NewQueue(cap, maxCap, crossbow.MakePolicy[E](policy)),
		closeNotify: make(chan struct{}, 1),
		eventNotify: make(chan struct{}, 1),
	}
	return subscription, nil
}

func (s *Subscription[E]) shutdown(reason error) {
	select {
	case s.closeNotify <- struct{}{}:
	default:
	}
	close(s.closeNotify)
	close(s.eventNotify)
	s.mailbox.Close()
}

func (s *Subscription[E]) Next() (E, error) {
	var zero E
	if s.canceled.Load() {
		return zero, ErrSubscriptionCanceled
	}

	elem, err := s.mailbox.Pop()
	if !err {
		return elem, ErrMailboxEmpty
	}

	return elem, nil
}

func (s *Subscription[E]) Recieve(ctx context.Context) (E, error) {
	var zero E
	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	case <-s.eventNotify:
		elem, err := s.mailbox.Pop()
		if !err {
			return elem, ErrMailboxEmpty
		}

		return elem, nil
	case <-s.closeNotify:
		return zero, ErrSubscriptionCanceled
	}
}

func (s *Subscription[E]) PublishTo(ctx context.Context, event E) error {
	if s.canceled.Load() {
		return ErrSubscriptionCanceled
	}

	err := s.mailbox.Push(ctx, event, make(chan error, 3))
	if err != nil {
		return err
	}

	if !s.canceled.Load() {
		select {
		case s.eventNotify <- struct{}{}:
		default:
		}
	}
	return nil
}

func (s *Subscription[E]) tryPublishTo(event E) error {
	if s.canceled.Load() {
		return ErrSubscriptionCanceled
	}

	err := s.mailbox.TryPush(event, nil)
	if err != nil {
		return err
	}

	if !s.canceled.Load() {
		select {
		case s.eventNotify <- struct{}{}:
		default:
		}
	}
	return nil
}

func (s *Subscription[E]) Close(reason error) {
	swapped := s.canceled.CompareAndSwap(false, true)
	if swapped {
		s.shutdown(reason)
	}
}

func (s *Subscription[E]) GetCloseNotifyChan() chan struct{} {
	return s.closeNotify
}

func (s *Subscription[E]) GetEventNotifyChan() chan struct{} {
	return s.eventNotify
}

func (s *Subscription[E]) Canceled() bool {
	return s.canceled.Load()
}
