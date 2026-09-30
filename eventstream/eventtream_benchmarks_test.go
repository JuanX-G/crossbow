package eventstream

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanX-G/crossbow"
)

func BenchmarkPublishConsume(b *testing.B) {
	strm := NewEventStream[int](10, 0, 1024)

	sub, err := strm.Subscribe(crossbow.PolicyDropNewest)
	if err != nil {
		b.Fatal(err)
	}

	ctx := b.Context()

	done := make(chan struct{})

	go func() {
		defer close(done)

		for {
			_, err := sub.Next()
			if err == nil {
				continue
			}

			if sub.Canceled() {
				return
			}
		}
	}()

	b.ResetTimer()

	for b.Loop() {
		_ = strm.Publish(ctx, TEST_PUBLISH_EVENT)
	}

	b.StopTimer()
	sub.Close(nil)
	<-done
}

func BenchmarkConcurrentPublish(b *testing.B) {
	strm := NewEventStream[int](10, 0, 1024)

	sub, err := strm.Subscribe(crossbow.PolicyDropNewest)
	if err != nil {
		b.Fatal(err)
	}

	done := make(chan struct{}, 1)
	defer close(done)
	go func() {
		for {
			_, err := sub.Recieve(b.Context())
			if err == nil {
				continue
			}

			if sub.Canceled() {
				return
			}
			_, ok := <-done
			if ok {
				break
			}
		}
	}()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = strm.Publish(b.Context(), TEST_PUBLISH_EVENT)
		}
	})

	b.StopTimer()
	sub.Close(nil)

	done <- struct{}{}
}

const TEST_FAST_PUBLISHING_SUB_COUNT = 128 ^ 2
const TEST_FAST_PUBLISHING_PUB_COUNT = 128 * 64
const TEST_FAST_PUBLISHING_EVENT_COUNT = 4096 * 16

func BenchmarkEventStreamFastPublishing(t *testing.B) {
	strm := NewEventStream[int](5, 0, 4)

	subs := make([]*Subscription[int], 0, TEST_FAST_PUBLISHING_SUB_COUNT)

	arrived := sync.Map{}
	expected := make([]int, 0, TEST_FAST_PUBLISHING_EVENT_COUNT)

	for i := range TEST_FAST_PUBLISHING_EVENT_COUNT {
		expected = append(expected, i)
		arrived.Store(i, struct{}{})
	}

	done := atomic.Bool{}

	for _ = range TEST_FAST_PUBLISHING_SUB_COUNT {
		sub, err := strm.Subscribe(crossbow.PolicyDropNewest)
		if err != nil {
			t.Fatalf("expected `err == nil` for subscribing, found: %+v", err)
		}

		go func() {
			for done.Load() {
				ev, _ := sub.Recieve(t.Context())
				arrived.Store(ev, struct{}{})
			}
		}()
		subs = append(subs, sub)
	}

	doneCh := make(chan struct{}, 1)

	tokens := make(chan struct{}, TEST_FAST_PUBLISHING_PUB_COUNT)
	var num atomic.Int32
	num.Store(-1)
	go func() {
		for _ = range TEST_FAST_PUBLISHING_EVENT_COUNT {
			tokens <- struct{}{}
			go func() {
				num.Add(1)
				strm.Publish(t.Context(), int(num.Load()))
				defer func() { <-tokens }()
			}()
		}
		doneCh <- struct{}{}
	}()

	select {
	case <-t.Context().Done():
		break
	case <-doneCh:
		time.Sleep(1000 * time.Millisecond)
	}

	notFoundCount := 0
	for _, n := range expected {
		_, ok := arrived.Load(n)
		if !ok {
			notFoundCount++
		}
	}
	t.Logf("expected for: %d messages to arrive, found: %d", TEST_FAST_PUBLISHING_EVENT_COUNT, TEST_FAST_PUBLISHING_EVENT_COUNT-notFoundCount)
	t.Logf("%f percent have not arrived", float64(notFoundCount)/float64(TEST_FAST_PUBLISHING_EVENT_COUNT))
}
