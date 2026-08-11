package crossbow

import (
	"context"
	"testing"
	"time"
)

func TestMessageTimestamp(t *testing.T) {
	msg := newContextMessage[int, int](context.Background(), 10, nil, true)
	if msg.Timestamp.IsZero() {
		t.Fatalf("expected a non-zero timestamp, found %s", msg.Timestamp.Format(time.UnixDate))
	}

	msg = newContextMessage[int, int](context.Background(), 10, nil, false)
	if !msg.Timestamp.IsZero() {
		t.Fatalf("expected a zero timestamp, found %s", msg.Timestamp.Format(time.UnixDate))
	}
}
