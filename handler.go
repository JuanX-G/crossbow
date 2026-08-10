/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package crossbow

type ServerHandler[M any, O any] interface {
	Init() error
	Handle(ContextMessage[M, O]) (O, error)
	Terminate(reason error)
}
