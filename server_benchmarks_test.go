/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package crossbow

import (
	"context"
	"testing"
)

func BenchmarkSend(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := setUpEchoServer(nil, ctx)
	for b.Loop() {
		srv.Send(ctx, 2137)
	}
}

func BenchmarkCall(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := setUpEchoServer(nil, ctx)
	for b.Loop() {
		srv.Call(ctx, 2137)
	}
}

// Benchmarking sending from multiple goroutines
func BenchmarkParallelSend(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := setUpEchoServer(nil, ctx)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = srv.Send(ctx, 2137)
		}
	})
}

func BenchmarkParallelCall(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := setUpEchoServer(nil, ctx, SERVICING_ROUTINES)
	for b.Loop() {
		srv.Send(ctx, 2137)
	}
}

func BenchmarkCallToParallel(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := setUpEchoServer(nil, ctx, SERVICING_ROUTINES)
	for b.Loop() {
		srv.Call(ctx, 2137)
	}
}
