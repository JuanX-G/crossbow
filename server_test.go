/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package crossbow

import (
	"context"
	"errors"
	"testing"
)


func TestCallServer(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
	srv := setUpEchoServer(t, ctx, true)
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
	srv := setUpEchoSendServer(t, ctx, true, resCh)
	defer cancel()

	err := srv.Send(ctx, 42)
	if res := <- resCh; res != "42" {
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

	cfg := MakeDefaultServerConfig[*EchoSend]()
	_, err = NewServer(handler, cfg)
	if err == nil {
		t.Fatalf("Server passed initialization, even though it should have stopped because handler.Init() returned an error")
	}
}


func TestServerShutdown(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
	srv := setUpEchoServer(t, ctx, true)

	go srv.Run(ctx)
	cancel()
	err := srv.Send(ctx, -42)
	if err == nil {
		t.Fatalf("Server context canceled, should have errored when trying to send data")
	}
}

func TestServerShutdownReturnsTerminated(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
	srv := setUpHangerServer(t, ctx, true)

	go srv.Run(ctx)
	err := srv.Send(context.Background(), 42)
	cancel()
	_, err = srv.Call(context.Background(), 42)
	if err == nil {
		t.Fatalf("Server context canceled, should have errored to a message waiting in the queue")
	}
	if !errors.Is(err, ErrServerTerminated) {
		t.Fatalf("Server context canceled, should have reported error 'ErrServerTerminated' when trying to send data, found %#v", err)
	}
}
