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
	"sync/atomic"
)

type StatsSnapshot struct {
	Failures uint64
	Panics   uint64
}

type serverStats struct {
	failures atomic.Uint64
	panics   atomic.Uint64
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
