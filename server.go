/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package crossbow

import (
	"crossbow/internal/queue"
	"sync"
	"sync/atomic"
)

// Server is the central object for crossbow. It represents a handler with a inbox attached. The handler processes messages from the inbox in a FIFO manner by default.
// The first type parameter is a user defined handler that implements the ServerHandler[M any, O any] interface. M is the message type that the handler acccepts
// wrapped with additional context by ContextMessage[M]. O is the output from the handler, once the handler finishes it must return O and error, both will be passed on to
// the caller when using the Call method, and disacarded if using Send. The server is threadsafe within its public APIs. The default config,
// provided by MakeDefaultServerConfig [MakeDefaultServerConfig], sets the max worker count to 1; this guarantees FIFO processing od messages sent like in regular actor systems.
type Server[T ServerHandler[M, O], M any, O any] struct {
	queue             *queue.Queue[ContextMessage[M, O]]
	handler           T
	sem               chan struct{}
	wg                sync.WaitGroup
	terminated        atomic.Bool
	stats             serverStats
	generateTimestamp bool
	recover           func(ContextMessage[M, O], T, error, []byte) error
}

// NewServer returns a ready to run server. Importantly, in this state the server may hang forever on any messages sent.
// Handling of deadlines and cancelations if only guaranteed after running Server.Run() [Server.Run()]. New Server accepts the user-defined handler and
// server config [ServerConfig] as its agruments. It run the Validate() [ServerConfig.Validate()] methods on the passed config,
// returning early with the errors and nil passed as the server pointer. NewServer is also responsible for running Init() on the handler. Upon encoutering an error
// it returns that error and nil as the server pointer. NewSerevr initializes the queue that back the mailbox with setting retrived from ServerConfig.
// It initializes the queue with a size of `1 + cfg.Worker`.
func NewServer[T ServerHandler[M, O], M any, O any](handler T, cfg ServerConfig, recover RecoverFn[T, M, O]) (*Server[T, M, O], error) {
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}
	if recover == nil {
		return nil, ServerConfigError{ErrType: ConfigErrorRecoverFuncNil}
	}

	if err := handler.Init(); err != nil {
		return nil, err
	}
	return &Server[T, M, O]{
		queue:             queue.NewQueue(1+int(cfg.Workers), int(cfg.MailboxSize), makePolicy[ContextMessage[M, O]](cfg.Policy)),
		handler:           handler,
		sem:               make(chan struct{}, cfg.Workers),
		recover:           recover,
		generateTimestamp: cfg.GenerateTimestamps,
	}, nil
}
