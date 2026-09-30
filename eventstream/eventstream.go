package eventstream

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/JuanX-G/crossbow"
)

type EventBus[E any] interface {
	Publish(context.Context, E) error
	Subscribe(...int) (Subscription[E], error)
	Close()
}

type EventStream[E any] struct {
	mu sync.RWMutex
	defaultMailboxCap int
	closed atomic.Bool
	subscriptions []*Subscription[E]
}

const DEFAULT_MAILBOX_CAP = 6

func NewEventStream[E any](defaultMailboxCap ...int) *EventStream[E] {
	es := &EventStream[E]{}

	cap := DEFAULT_MAILBOX_CAP
	if len(defaultMailboxCap) >= 1 {
		if defaultMailboxCap[0] > 0 {
			cap = defaultMailboxCap[0]
		}
	}
	es.defaultMailboxCap = cap

	return es
}

func(es *EventStream[E]) Publish(ctx context.Context, event E) error {
	es.mu.RLock()
	defer es.mu.RUnlock()
	for i := range len(es.subscriptions) {
		es.subscriptions[i].PublishTo(ctx, event)
	}

	return nil
}

func(es *EventStream[E]) Subscribe(policy crossbow.MailboxPolicy, maxCap ...int) (*Subscription[E], error) {
	cap := es.defaultMailboxCap
	if len(maxCap) >= 1 {
		if maxCap[0] > 0 {
			cap = maxCap[0]
		}
	}

	es.mu.Lock()
	defer es.mu.Unlock()
	if es.closed.Load() {
		return nil, ErrEventStreamClosed
	}

	subscription, err := NewSubscription[E](cap, policy)
	if err != nil {
		return nil, err
	}
	es.subscriptions = append(es.subscriptions, subscription)
	return subscription, nil
}

func(es *EventStream[E]) Close() {
    es.mu.Lock()
    defer es.mu.Unlock()

    if !es.closed.CompareAndSwap(false, true) {
        return
    }

    for _, sub := range es.subscriptions {
        sub.Close()
    }
}
