/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package crossbow

import (
	"crossbow/internal/queue"
	"fmt"
)

// DefaultPanicRecover is a basic handler for panics inside of the handler. It will return
// en error causing the server to terminate.
func DefaultPanicRecover[T ServerHandler[M, O], M any, O any](msg ContextMessage[M, O], handler T, err error, stack []byte) error {
	return fmt.Errorf("handler panicked with error: %s. Stack: %s; handler: %+v; message: %+v", err, string(stack), handler, msg)
}

// MailboxPolicy represents one of the predefined MailboxPolicies provided by the library.
type MailboxPolicy int

const (
	PolicyUnbounded  MailboxPolicy = iota // The queue is allowed to grow indefnitely, generally dicouraged for most use cases
	PolicyBlock                           // Default policy, if the queue is full Send() or Call() will block until there is space in the queue
	PolicyDropNewest                      // If the queue if full Send() or Call() will drop the message and return an appropriate error
	PolicyDropOldest                      // If the queue if full Send() or Call() will drop the oldest message (first in line) and the enqueue the just passed message. If a failure occurs ErrFull will be returned
)

func makePolicy[T any](mp MailboxPolicy) queue.MailboxPolicy[T] {
	switch mp {
	case PolicyBlock:
		return queue.BlockPolicy[T]{}
	case PolicyUnbounded:
		return queue.UnboundedPolicy[T]{}
	case PolicyDropNewest:
		return queue.DropNewestPolicy[T]{}
	case PolicyDropOldest:
		return queue.DropOldestPolicy[T]{}
	default:
		return queue.BlockPolicy[T]{}
	}
}

// ServerConfig provides a recovery function settings for a server and facilitates easy validation
type ServerConfig struct {
	Workers            uint          // The amount of workers to run the handler over, 1 for standard FIFO ordering
	MailboxSize        uint          // Maximum size of a the servers mailbox, must be greater than 0
	Policy             MailboxPolicy // MailboxPolicy according to [MailboxPolicy]
	GenerateTimestamps bool          // Switch for autogeneration of timestamps for the Timestamp field on ContextMessage [ContextMessage]
	Unbounded          bool
}

type RecoverFn[T ServerHandler[M, O], M any, O any] = func(ContextMessage[M, O], T, error, []byte) error // function called when the handler panics, if it returns an erros the server terminates

// MakeDefaultServerConfig provides a sesible default configuration for a server.
// If no wokers argument is provided a default of 1 is used. Arguments past the first
// one are ignored, but the library reserves a right to use them for any purpose in the future
// will cause the server to terminate. It also uses the PolicyBlock for its mailbox.
// The mailbox size is set to (4 + Workers * 4).
func MakeDefaultServerConfig(workers ...uint) ServerConfig {
	var workerCount uint
	if len(workers) == 1 {
		if workers[0] != 0 {
			workerCount = workers[0]
		} else {
			workerCount = 1
		}
	} else {
		workerCount = 1
	}

	return ServerConfig{
		Workers:            workerCount,
		MailboxSize:        4 + workerCount*4,
		Policy:             PolicyBlock,
		GenerateTimestamps: false,
	}
}

// Validate checks the fields of [ServerConfig] for disallowed values, returns an apropriate erros when
// a disallowed value is found.
func (s *ServerConfig) Validate() error {
	if s.Workers == 0 {
		return ServerConfigError{ErrType: ConfigErrorWorkersZero}
	}
	if s.MailboxSize == 0 && !s.Unbounded {
		return ServerConfigError{ErrType: ConfigErrorMailboxSizeZero}
	}
	return nil
}
