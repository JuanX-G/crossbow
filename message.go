/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package crossbow

import (
	"context"
	"time"
)

// ContextMessage holds a message, a channel where a response should be sent
// and general context such as.
type ContextMessage[M any, O any] struct {
	Value M // The message being sent to the handler
	reply chan Response[O]

	Context   context.Context // Context to be passed to handlers
	Timestamp time.Time       // Timestamp,  populated automatically if makeTimestamp is true
}

func newContextMessage[M any, O any](ctx context.Context, value M, resCh chan Response[O], makeTimestamp bool) ContextMessage[M, O] {
	if makeTimestamp {
		return ContextMessage[M, O]{Value: value, Timestamp: time.Now(), Context: ctx, reply: resCh}
	} else {
		return ContextMessage[M, O]{Value: value, Context: ctx, reply: resCh}
	}
}

// Response from a handler that carries the error. It is up to the programmers discrection
// when an error should be returned and if Value can be valid if Err != nil.
type Response[O any] struct {
	Value O
	Err   error
}
