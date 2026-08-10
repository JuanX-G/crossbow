/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package crossbow

import (
	"testing"
	"context"
	"strconv"
	"fmt"
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
func setUpEchoServer(t *testing.T, ctx context.Context, testing bool, threads ...uint) *Server[Echo, int, string] {
	handler := Echo{}
	cfg := MakeDefaultServerConfig[Echo](threads...)

	srv, err := NewServer(handler, cfg)
	if err != nil && testing {
		t.Fatalf("Error %s at server initialization", err)
	} else if err != nil {
		panic(fmt.Sprintf("Error %s at server initialization", err))
	}
	go srv.Run(ctx)
	return srv
}

// Simple stateful handler that send back data through the provided channel
type EchoSend struct{
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
func setUpEchoSendServer(t *testing.T, ctx context.Context, testing bool, resCh chan string, threads ...uint) (*Server[*EchoSend, int, string]) {
	handler := &EchoSend{BackCh: resCh}
	cfg := MakeDefaultServerConfig[*EchoSend](threads...)

	srv, err := NewServer(handler, cfg)
	if err != nil && testing {
		t.Fatalf("Error %s at server initialization", err)
	} else if err != nil {
		panic(fmt.Sprintf("Error %s at server initialization", err))
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
func setUpHangerServer(t *testing.T, ctx context.Context, testing bool, threads ...uint) *Server[Hanger, int, string] {
	handler := Hanger{}
	cfg := MakeDefaultServerConfig[Hanger](threads...)

	srv, err := NewServer(handler, cfg)
	if err != nil && testing {
		t.Fatalf("Error %s at server initialization", err)
	} else if err != nil {
		panic(fmt.Sprintf("Error %s at server initialization", err))
	}
	go srv.Run(ctx)
	return srv
}
