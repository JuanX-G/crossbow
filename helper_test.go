/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package crossbow

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

// Example of a simple stateless handler.
type Echo struct{}

func (Echo) Handle(msg ContextMessage[int, string]) (string, error) {
	return strconv.Itoa(msg.Value), nil
}

func (Echo) Init() error {
	return nil
}

func (Echo) Terminate(err error) {
}

// Sets up a server using the Echo handler. Testing informs the handler if t will be nil.
// For benchmarks for example set testing = false and pass nil as the t parameter.
// If testing = false the function panics on sever intialization error.
// The function runs the server for you in a goroutine.
// The Argument threads is passed, unrolled, into [MakeDefaultServerConfig].
func setUpEchoServer(t testing.TB, ctx context.Context, threads ...uint) *Server[Echo, int, string] {
	handler := Echo{}
	cfg := MakeDefaultServerConfig(threads...)

	srv, err := NewServer(handler, cfg, DefaultPanicRecover)
	if err != nil {
		t.Fatalf("Error %s at server initialization", err)
	}
	go srv.Run(ctx)

	return srv
}

// Simple stateful handler that send back data through the provided channel
type EchoSend struct {
	BackCh chan string
}

func (e *EchoSend) Handle(msg ContextMessage[int, string]) (string, error) {
	e.BackCh <- strconv.Itoa(msg.Value)
	return "", nil
}

func (e EchoSend) Init() error {
	if e.BackCh == nil {
		return fmt.Errorf("send back channel cannot be nil")
	}
	return nil
}

func (e EchoSend) Terminate(err error) {
	close(e.BackCh)
}

// See [setUpEchoServer].
func setUpEchoSendServer(t testing.TB, ctx context.Context, resCh chan string, threads ...uint) *Server[*EchoSend, int, string] {
	handler := &EchoSend{BackCh: resCh}
	cfg := MakeDefaultServerConfig(threads...)

	srv, err := NewServer(handler, cfg, DefaultPanicRecover)
	if err != nil {
		t.Fatalf("Error %s at server initialization", err)
	}
	go srv.Run(ctx)
	return srv
}

// Example of a handler that hangs on the first message it recives
type Hanger struct{}

func (Hanger) Handle(msg ContextMessage[int, string]) (string, error) {
	for {

	}
}

func (Hanger) Init() error {
	return nil
}

func (Hanger) Terminate(err error) {
}

// Sets up a server using the Echo handler. Testing informs the handler if t will be nil.
// For benchmarks for example set testing = false and pass nil as the t parameter.
// If testing = false the function panics on sever intialization error.
// The function runs the server for you in a goroutine.
// The Argument threads is passed, unrolled, into [MakeDefaultServerConfig].
func setUpHangerServer(t testing.TB, ctx context.Context, threads ...uint) *Server[Hanger, int, string] {
	handler := Hanger{}
	cfg := MakeDefaultServerConfig(threads...)

	srv, err := NewServer(handler, cfg, DefaultPanicRecover)
	if err != nil {
		t.Fatalf("Error %s at server initialization", err)
	}
	go srv.Run(ctx)
	return srv
}

type Panicker struct{}

func (Panicker) Handle(msg ContextMessage[int, string]) (string, error) {
	panic("")
}

func (Panicker) Init() error {
	return nil
}

func (Panicker) Terminate(err error) {
}

func setUpResponderServer(t testing.TB, ctx context.Context, size int, threads ...uint) (*Server[*Responder, int, int], chan int) {
	ch := make(chan int, size)
	handler := &Responder{res: ch}
	cfg := MakeDefaultServerConfig(threads...)

	srv, err := NewServer(handler, cfg, DefaultPanicRecover)
	if err != nil {
		t.Fatalf("Error %s at server initialization", err)
	}
	go srv.Run(ctx)
	return srv, ch
}

type Responder struct {
	res     chan int
	size    int
	working atomic.Int32
}

func (t *Responder) Handle(msg ContextMessage[int, int]) (int, error) {
	t.working.Add(1)
	defer t.working.Add(-1)
	t.res <- msg.Value
	return 0, nil
}

func (t *Responder) Init() error {
	if t.res == nil {
		return fmt.Errorf("res cannot be nil")
	}
	return nil
}

func (t *Responder) Terminate(err error) {
	defer close(t.res)
	if t.working.Load() != 0 {
		time.Sleep(time.Millisecond * 50)
	}
}

func setUpCounterServer(t testing.TB, ctx context.Context, threads ...uint) (*Server[*Counter, int, int], *atomic.Uint64) {
	handler := &Counter{}
	cfg := MakeDefaultServerConfig(threads...)

	srv, err := NewServer(handler, cfg, DefaultPanicRecover)
	if err != nil {
		t.Fatalf("Error %s at server initialization", err)
	}
	go srv.Run(ctx)
	return srv, &handler.Counter
}

type Counter struct {
	Counter atomic.Uint64
	working atomic.Int32
}

func (c *Counter) Handle(msg ContextMessage[int, int]) (int, error) {
	c.working.Add(1)
	defer c.working.Add(-1)
	c.Counter.Add(1)
	return 0, nil
}

func (c *Counter) Init() error {
	return nil
}

func (c *Counter) Terminate(err error) {
}
