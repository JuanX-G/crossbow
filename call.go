/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package crossbow

import (
	"context"
	"fmt"
)

// Send is the basic method for signaling to the handler, it does not block waiting for a response.
// It returns a context error upon cancelation of ctx. An error is returned if creating a new
// ContextMessage fail. An error is also returned if enqueueing
// the message fails. If nil is returned that means that the message was successfully enqueued
func (s *Server[T, M, O]) Send(ctx context.Context, req M) error {
	if s.terminated.Load() {
		return ErrServerTerminated
	}

	msg := newContextMessage[M, O](ctx, req, nil, s.generateTimestamp)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if err := s.enqueue(msg); err != nil {
			return fmt.Errorf("queue error: %w", err)
		}
	}
	return nil
}

// Call is the basic method for talking to the handler, it blocks until the handler
// returns a response.
// It returns a context error upon cancelation of ctx. An error is returned if creating a new
// ContextMessage fail. An error is also returned if enqueueing
// the message fails. If nil is returned that means that the message was successfully enqueued
func (s *Server[T, M, O]) Call(ctx context.Context, req M) (O, error) {
	var zero O
	if s.terminated.Load() {
		return zero, ErrServerTerminated
	}

	resCh := s.replyPool.Get().(chan Response[O])
	msg := newContextMessage(ctx, req, resCh, s.generateTimestamp)

	if err := s.enqueue(msg); err != nil {
		s.replyPool.Put(resCh)
		return zero, fmt.Errorf("queue error: %w", err)
	}

	select {
	case <-ctx.Done():
		return zero, ctx.Err()
	case res := <-resCh:
		s.replyPool.Put(resCh)
		return res.Value, res.Err
	}
}

// Terminated reports whether the server has terminated.
func (s *Server[T, M, O]) Terminated() bool {
	return s.terminated.Load()
}

// MailboxLen returns the length of the queue that backs the Server's mailbox
func (s *Server[T, M, O]) MailboxLen() int {
	return s.queue.Len()
}

// Stats returns a StatsSnapshot [StatsSnapshot], it reads the atomic Uints in the dynamic stats
// and saves it into the structure. It is thus guranteed to be thread safe but, the stats may change from the time
// the values were saved and this function outputed the snapshot.
func (s *Server[T, M, O]) Stats() StatsSnapshot {
	return s.stats.Snapshot()
}
