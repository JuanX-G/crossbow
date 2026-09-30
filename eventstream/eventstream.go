package eventstream

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JuanX-G/crossbow"
	"github.com/JuanX-G/crossbow/internal/queue"
)

type EventBus[E any] interface {
	Publish(context.Context, E) error
	Subscribe(crossbow.MailboxPolicy, ...int) (*Subscription[E], error)
	Close()
}

type subscriptionData[E any] struct {
	sub       *Subscription[E]
	failCount atomic.Uint32
}

type EventStream[E any] struct {
	mu                sync.RWMutex
	defaultMailboxCap int
	closed            atomic.Bool
	subscriptions     atomic.Pointer[map[*subscriptionData[E]]struct{}]
	slowTimeout       int // How many seconds a slow subscriber can block before the event is dropped.
	slowFailureCap    int // How many events may be dropped by a slow subscriber before it is kicked out. -1 means infinite, 0 means instant kickout.
}

const DEFAULT_MAILBOX_CAP = 6

func NewEventStream[E any](slowTimeout, slowFailureCap int, defaultMailboxCap ...int) *EventStream[E] {
	es := &EventStream[E]{
		subscriptions:  atomic.Pointer[map[*subscriptionData[E]]struct{}]{},
		slowTimeout:    slowTimeout,
		slowFailureCap: slowFailureCap,
	}
	new := make(map[*subscriptionData[E]]struct{})
	es.subscriptions.Store(&new)

	cap := DEFAULT_MAILBOX_CAP
	if len(defaultMailboxCap) >= 1 {
		if defaultMailboxCap[0] > 0 {
			cap = defaultMailboxCap[0]
		}
	}
	es.defaultMailboxCap = cap

	return es
}

const SLOW_PUSH_FANOUT_ROUTINES = 3

func (es *EventStream[E]) Publish(ctx context.Context, event E) error {
	if es.closed.Load() {
		return ErrEventStreamClosed
	}
	snap := es.subscriptions.Load()
	if snap == nil {
		return nil
	}

	var slow []*subscriptionData[E]
	var errs []error

	for s, _ := range *snap {
		if err := s.sub.tryPublishTo(event); errors.Is(err, queue.ErrFull) {
			slow = append(slow, s)
		} else if err != nil {
			errs = append(errs, err)
		} else {
			s.sub.PublishTo(ctx, event)
		}
	}

	if len(slow) == 0 {
		return errors.Join(errs...)
	}

	var timeout time.Duration
	var hasTimeout bool
	if es.slowTimeout > 0 {
		hasTimeout = true
		timeout = time.Second * time.Duration(es.slowTimeout)
	}

	var cancel context.CancelFunc
	defer cancel()

	tokens := make(chan struct{}, SLOW_PUSH_FANOUT_ROUTINES)
	wg := sync.WaitGroup{}

	for _, s := range slow {
		<-tokens
		wg.Go(func() {
			defer func() {
				tokens <- struct{}{}
			}()
			var pubCtx context.Context
			if hasTimeout {
				pubCtx, cancel = context.WithTimeout(ctx, timeout)
			}
			if err := s.sub.PublishTo(pubCtx, event); err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					if es.slowFailureCap == 0 || es.slowFailureCap < int(s.failCount.Load()) {
						s.sub.Close(ErrSlowConsumer)
						es.remove(s)
					} else {
						s.failCount.Add(1)
					}
				}
				errs = append(errs, err)
			}
		})
	}

	done := make(chan struct{}, 1)
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		errs = append(errs, ctx.Err())
		return errors.Join(errs...)

	case <-done:
	}

	return errors.Join(errs...)
}

func (es *EventStream[E]) Subscribe(policy crossbow.MailboxPolicy, maxCap ...int) (*Subscription[E], error) {
	cap := es.defaultMailboxCap
	if len(maxCap) >= 1 {
		if maxCap[0] > 0 {
			cap = maxCap[0]
		}
	}

	sub, err := NewSubscription[E](cap, policy)
	if err != nil {
		return nil, err
	}

	es.mu.Lock()
	defer es.mu.Unlock()
	if es.closed.Load() {
		return nil, ErrEventStreamClosed
	}

	new := *es.subscriptions.Load()
	new[&subscriptionData[E]{sub: sub}] = struct{}{}

	es.subscriptions.Swap(&new)

	return sub, nil
}

func (es *EventStream[E]) Close() {
	es.mu.Lock()
	defer es.mu.Unlock()

	if !es.closed.CompareAndSwap(false, true) {
		return
	}

	for subdata, _ := range *es.subscriptions.Load() {
		s := subdata.sub
		s.Close(nil)
	}
}

func (es *EventStream[E]) Unsubscribe(s *subscriptionData[E]) *queue.Queue[E] {
	es.remove(s)
	return s.sub.mailbox
}

func (es *EventStream[E]) remove(s *subscriptionData[E]) {
	es.mu.Lock()
	defer es.mu.Unlock()

	old := *es.subscriptions.Load()
	delete(old, s)

	es.subscriptions.Swap(&old)
}
