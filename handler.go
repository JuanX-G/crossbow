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
	"crossbow/internal/queue"
)

type ServerHandler[M any, O any] interface {
	Init() error
	Handle(ContextMessage[M, O]) (O, error)
	Terminate(reason error)
}

// CustomMailboxPolicy defines rules for handling attempts to push to
// a full queue. The last argument is the item being pushed, the second
// one the the queue object [internal/queue]. The first argument is the
// context passed by the caller, you do not need to handle it in any
// special way. Return an apropriate error if the pushing fails,
// if so, you generally should not modify the underlying queues content so as
// to not confuse callers.
type CustomMailboxPolicy[T any] interface {
	queue.MailboxPolicy[T] // required function: ` Enqueue(context.Context, *Queue[T], T) error `
}
