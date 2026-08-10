/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package crossbow

import (
	"context"
	"time"
	"github.com/google/uuid"
)

// ContextMessage holds a message, a channel where a response should be sent
// and general context such as. ID may be used by the programmer for custom IDs
// if GenerateUUIDs is set to false. If GenerateUUIDs this field will be overwritten.
type ContextMessage[M any, O any] struct {
    Value M // The message being sent to the handler
	res chan Response[O]

    Context context.Context // Context to be passed to handlers
    ID string // ID of the message, reserved if GenerateUUIDs == true; available to the user for custom IDs in all other cases
    Timestamp time.Time // Timestamp,  populated automatically when wrappaing a message M in this struct
}

func newContextMessage[M any, O any](ctx context.Context, value M, resCh chan Response[O], makeUUID bool) (ContextMessage[M, O], error) {
	var zero ContextMessage[M, O]
	if makeUUID {
		id, err := uuid.NewRandom()
		if err != nil {
			return zero, err
		}

		return ContextMessage[M, O]{Value: value, Timestamp: time.Now(), Context: ctx, res: resCh, ID: string(id.String())}, nil
	}
	return ContextMessage[M, O]{Value: value, Timestamp: time.Now(), Context: ctx, res: resCh}, nil
}

// Response from a handler that carries the error. It is up to the programmers discrection
// when an error should be returned and if Value can be valid if Err != nil.
type Response[O any] struct {
    Value O
    Err   error
}
