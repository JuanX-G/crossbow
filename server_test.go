/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package crossbow

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCallServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	srv := setUpEchoServer(t, ctx)
	defer cancel()

	res, err := srv.Call(ctx, 42)
	if res != "42" {
		t.Fatalf("Wrong value returned from Calling Echo; expected: 42; found: %s", res)
	}
	if err != nil {
		t.Fatalf("Wrong error returned from Calling Echo; expected: nil; found: %#v", err)
	}

	if !srv.terminated.Load() {
		srv.shutdown(nil)
	}
}

func TestSendServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	resCh := make(chan string)
	srv := setUpEchoSendServer(t, ctx, resCh)
	defer cancel()

	err := srv.Send(ctx, 42)
	if res := <-resCh; res != "42" {
		t.Fatalf("Wrong value returned from Sending to Echo; expected: 42; found: %s", res)
	}
	if err != nil {
		t.Fatalf("Wrong error returned from Calling Echo; expected: nil; found: %#v", err)
	}

	if !srv.terminated.Load() {
		srv.shutdown(nil)
	}
}

func TestHandlerInitFails(t *testing.T) {
	// Pre-check handler behaviour
	handler := &EchoSend{BackCh: nil}
	err := handler.Init() // it is safe to call this twice of EchoSend but handlers do not have to allow this
	if err == nil {
		t.Fatalf("Handler behaviour invalid, should have returned an error when given a nil channel.")
	}

	cfg := MakeDefaultServerConfig()
	_, err = NewServer(handler, cfg, DefaultPanicRecover)
	if err == nil {
		t.Fatalf("Server passed initialization, even though it should have stopped because handler.Init() returned an error")
	}
}

func TestServerShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	srv := setUpEchoServer(t, ctx)

	cancel()
	err := srv.Send(ctx, -42)
	if err == nil {
		t.Fatalf("Server context canceled, should have errored when trying to send data")
	}
}

func TestServerShutdownReturnsTerminated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	srv := setUpHangerServer(t, ctx)

	srv.Send(context.Background(), 42)
	cancel()
	_, err := srv.Call(context.Background(), 42)
	if err == nil {
		t.Fatalf("Server context canceled, should have errored to a message waiting in the queue")
	}
	if !errors.Is(err, ErrServerTerminated) {
		t.Fatalf("Server context canceled, should have reported error 'ErrServerTerminated' when trying to send data, found %#v", err)
	}
}

func TestServerPanicRecovery(t *testing.T) {
	handler := &Panicker{}

	cfg := MakeDefaultServerConfig()
	panicErr := errors.New("panicked")
	panicHandler := func(msg ContextMessage[int, string], h *Panicker, err error, s []byte) error {
		return panicErr
	}
	srv, err := NewServer(handler, cfg, panicHandler)
	if err != nil {
		t.Fatalf("server initialization failed")
	}
	go srv.Run(context.Background())
	_, err = srv.Call(context.Background(), 21)
	if err == nil {
		t.Fatalf("handler panicked, and recovery returned an error, but err == nil was returned")
	}
	if !srv.Terminated() {
		t.Fatalf("handler panicked, and recovery returned an error, but server reports terminated == false")
	}
}

func TestMultipleWorkersSendback(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Duration(time.Second*3))
	defer cancel()
	srv, ch := setUpResponderServer(t, ctx, 4, 4)

	for range 4 {
		srv.Send(ctx, 1)
	}

	counter := 0
	for v := range ch {
		counter++
		if v != 1 {
			t.Fatalf("expected 1 sent back, found: %d", v)
		}
	}
	if counter != 4 {
		t.Fatalf("expected 4 messages sent back, found: %d", counter)
	}
}

func TestMultipleWorkersCounter(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Duration(time.Second*4))
	defer cancel()
	srv, counter := setUpCounterServer(t, ctx, 4)

	for range 4 {
		srv.Send(ctx, 1)
	}
	time.Sleep(time.Duration(time.Millisecond * 600))
	if c := counter.Load(); c != 4 {
		t.Fatalf("expected 4 messages sent back, found: %d", c)
	}
}
