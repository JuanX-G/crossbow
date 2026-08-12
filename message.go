// Copyright 2026 Maciej "juan_em (JuanX-G)" Woźniak
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Or use the bundled copy in the LICENSE file.

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
