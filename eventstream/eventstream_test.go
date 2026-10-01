package eventstream

import (
	"testing"

	"github.com/JuanX-G/crossbow"
)

func TestNewEventStream(t *testing.T) {
	strm := NewEventStream[int](5, 0, 6)
	if strm.defaultMailboxCap != 6 {
		t.Fatalf("expected mailbox cap 6, found: %d", strm.defaultMailboxCap)
	}
}

const TEST_PUBLISH_EVENT = 42

func TestEventStreamPublish(t *testing.T) {
	strm := NewEventStream[int](5, 0, 4)

	sub, err := strm.Subscribe(crossbow.PolicyDropNewest)
	if err != nil {
		t.Fatalf("expected `err == nil` for subscribing, found: %+v", err)
	}

	_ = strm.Publish(t.Context(), TEST_PUBLISH_EVENT)

	recv, err := sub.Next()
	if err != nil {
		t.Fatalf("expected `err == nil` for reading a sent event, found: %+v", err)
	}
	if recv != TEST_PUBLISH_EVENT {
		t.Fatalf("expected for: %d to be sent, found: %d", TEST_PUBLISH_EVENT, recv)
	}
}

const TEST_MULTIPLE_SUBSCRIPTIONS_COUNT = 8

func TestEventStreamMultipleSubscriptions(t *testing.T) {
	strm := NewEventStream[int](5, 0, 4)

	subs := make([]*Subscription[int], 0, 8)

	for _ = range TEST_MULTIPLE_SUBSCRIPTIONS_COUNT {
		sub, err := strm.Subscribe(crossbow.PolicyDropNewest)
		if err != nil {
			t.Fatalf("expected `err == nil` for subscribing, found: %+v", err)
		}
		subs = append(subs, sub)
	}

	_ = strm.Publish(t.Context(), TEST_PUBLISH_EVENT)

	for i := range TEST_MULTIPLE_SUBSCRIPTIONS_COUNT {
		recv, _ := subs[i].Recieve(t.Context())
		if recv != TEST_PUBLISH_EVENT {
			t.Fatalf("expected for: %d to be sent, found: %d", TEST_PUBLISH_EVENT, recv)
		}
	}
}
