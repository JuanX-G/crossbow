/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package crossbow

import (
	"fmt"
	"crossbow/internal/queue"
)

// DefaultPanicRecover is a basic handler for panics inside of the handler. It will return
// en error causing the server to terminate.
func DefaultPanicRecover[T ServerHandler[M, O], M any, O any](msg ContextMessage[M, O], handler T, err error, stack []byte) error {
	return fmt.Errorf("Handler panicked with error: %s. Stack: %s; handler: %+v; message: %+v", err, string(stack), handler, msg)
}

// MailboxPolicy represents one of the predefined MailboxPolicies provided by the library.
type MailboxPolicy int
const (
	PolicyUnbounded MailboxPolicy = iota // The queue is allowed to grow indefnitely, generally dicouraged for most use cases
	PolicyBlock // Default policy, if the queue is full Send() or Call() will block until there is space in the queue
	PolicyDropNewest // If the queue if full Send() or Call() will drop the message and return an appropriate error
	PolicyDropOldest // If the queue if full Send() or Call() will drop the oldest message (first in line) and the enqueue the just passed message. If a failure occurs ErrFull will be returned
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
type ServerConfig[T ServerHandler[M, O], M any, O any] struct {
	ThreadsCfg uint // The amount of threads to run the handler over, 1 for standard FIFO ordering
	MailboxSize uint // Maximum size of a the servers mailbox, must be greater than 0
	MailboxPolicy // MailboxPolicy according to [MailboxPolicy]
	GenerateUUIDs bool // Switch for autogeneration of uuids for the ID field on ContextMessage [ContextMessage]
	RecoverFn func(ContextMessage[M, O], T, error, []byte) error // function called when the handler panics, if it returns an erros the server terminates
}

// MakeDefaultServerConfig provides a sesible default configuration for a server.
// If no threads argument is provided a default of 1 is used. Arguments past the first
// one are ignored, but the library reserves a right to use them for any purpose in the future
// This server uses the [DefaultPanicRecover] recovery handler, so any panics inside the handler
// will cause the server to terminate. It also uses the PolicyBlock for its mailbox.
// The mailbox size is set to (4 + threadsCount * 4).
func MakeDefaultServerConfig[T ServerHandler[M, O], M any, O any](threads... uint) ServerConfig[T, M, O] {
	var threadsCount uint
	if len(threads) == 1 {
		if threads[0] != 0 {
			threadsCount = threads[0]
		} else {
			threadsCount = 1
		}
	} else {
		threadsCount = 1
	}

	return ServerConfig[T, M, O]{
		ThreadsCfg: threadsCount,
		MailboxSize: 4 + threadsCount * 4,
		MailboxPolicy: PolicyBlock,
		RecoverFn: DefaultPanicRecover[T, M, O],
		GenerateUUIDs: false,
	}
}

// Validate checks the fields of [ServerConfig] for disallowed values, returns an apropriate erros when
// a disallowed value is found.
func (s *ServerConfig[T, M, O]) Validate() error {
	if s.ThreadsCfg == 0 {
		return ServerConfigError{errType: ConfigErrorThreadsCfgZero}
	}
	if s.MailboxSize == 0 {
		return ServerConfigError{errType: ConfigErrorMailboxSizeZero}
	}
	if s.RecoverFn == nil {
		return ServerConfigError{errType: ConfigErrorRecoverFuncNil}
	}
	return nil
}
