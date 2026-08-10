/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package crossbow

import (
	"sync/atomic"
)

type StatsSnapshot struct {
    Failures uint64
    Panics uint64
}

type serverStats struct {
	failures atomic.Uint64
	panics atomic.Uint64
}

func (s *serverStats) AddFail() {
	s.failures.Add(1)
}

func (s *serverStats) FailCount() uint64 {
	return s.failures.Load()
}

func (s *serverStats) AddPanic() {
	s.panics.Add(1)
}

func (s *serverStats) PanicCount() uint64 {
	return s.panics.Load()
}

func (s *serverStats) Snapshot() StatsSnapshot {
	return StatsSnapshot{Failures: s.FailCount(), Panics: s.PanicCount()}
}
