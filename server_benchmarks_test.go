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
	"testing"
)

func BenchmarkSend(b *testing.B) {
	ctx := b.Context()
	srv := setUpEchoServer(nil, ctx)
	for b.Loop() {
		srv.Send(ctx, 2137)
	}
}

func BenchmarkCall(b *testing.B) {
	ctx := b.Context()
	srv := setUpEchoServer(nil, ctx)
	for b.Loop() {
		srv.Call(ctx, 2137)
	}
}

// Benchmarking sending from multiple goroutines
func BenchmarkParallelSend(b *testing.B) {
	ctx := b.Context()
	srv := setUpEchoServer(nil, ctx)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = srv.Send(ctx, 2137)
		}
	})
}

func BenchmarkParallelCall(b *testing.B) {
	ctx := b.Context()
	srv := setUpEchoServer(nil, ctx)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = srv.Call(ctx, 2137)
		}
	})
}

const SERVICING_ROUTINES = 4 // Goroutines the parallel handler should be ran on.

// Benchmarking sending to handler that runs in parallel
func BenchmarkSendToParallel(b *testing.B) {
	ctx := b.Context()
	srv := setUpEchoServer(nil, ctx, SERVICING_ROUTINES)
	for b.Loop() {
		srv.Send(ctx, 2137)
	}
}

func BenchmarkCallToParallel(b *testing.B) {
	ctx := b.Context()
	srv := setUpEchoServer(nil, ctx, SERVICING_ROUTINES)
	for b.Loop() {
		srv.Call(ctx, 2137)
	}
}
