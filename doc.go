// Package crossbow Is a simple actor-like, inbox based worker-pool.
// It provides: lightweight lifecycle managment, panic recovery, adaptable mailbox policies and resizing,
// synchronous and asynchronous communications with the server; optional handler level parallelism
// and classic FIFO processing as default.
//
// Crossbow makes heavy use of generics to keep whole thing type safe. It is kept lean and lightweight on purpose.
// That is why it lacks a central registry, managers or process trees.
//
// # Why Crossbow
//
// Crossbow offers actor-like stage managment in a lightweight and type-safe fashion;
// while at the same time extending it with an optional worker pool.
//
// Crossbow offer configurable panic recovery that lets your provide your own handler
// to inspect the mssage causing the panic and the state of the handler.
//
// Included with the library is sensible set of mailbox backpressure policies, though
// by providing a simple interface you can write your own backpressure handlers.
//
// Crossbow allows you to limit the scope of every operation to an appropriate handler,
// making your code for modular and safer.
//
// # Versioning
//
// Crossbow is intended to strcitly follow SemVer from version v0.1 onwards. Releases marked
// as v0-alpha.x.y or v0-beta.x.y do not make any compatibility guarantees.
// We intends to use the standard go support cycle, each version of crossbow will be
// maintained for two version of the Go language.
//
// # Usage
//
// All examples can be found in ./examples
//
// To use Crossbow you need to define a handler that implements the ServerHandler[M any, O any] interface.
// Most importantly you need to have a Handle(ContextMessage[M, O]) method, this is the handling loop of your server;
// any data sent ot it is popped from the inbox and passed to it.
//
//		// Simple stateless handler that send back data through the provided channel
//		type Foo struct{
//		    x int
//		}
//
//		func (e *Foo) Handle(msg ContextMessage[int, int]) (int, error) {
//		    if e.x == 2137 {
//		        fmt.Println("21:37")
//		        e.x = 0
//		        return 0, nil
//		    } else {
//		        e.x = msg.Value // Update internal state. Safe if the thread count is 1
//		        return msg.Value, nil
//		    }
//		}
//
//		// Simple init stub. Usually used to initialize connections and validate fields
//		func (e *Foo) Init() error {
//			return nil
//		}
//
//		// Handles the closing of connections etc.
//		func (e *Foo) Terminate(err error) {
//		    if e.x == 42 {
//		        fmt.Println("Ah! 42")
//		    }
//		}
//
//		func main() {
//			handler := &Foo{}
//			cfg := MakeDefaultServerConfig(1) // one worker for sequential processing
//			srv, _ := NewServer[*Foo, int, int](handler, cfg, DefaultPanicRecover)
//
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			go srv.Run(ctx) // this must be done before sending anything to the server
//
//			rqCtx, _ := context.WithCancel(ctx) // canceling a requests context won't stop the whole server
//			_ = srv.Send(rqCtx, 2137) // Send() does not block awaiting a response and just discards any
//
//	    	out, err := srv.Call(rqCtx, 11)
//			// Output will be 0; err will be nil; and the string "21:37" will be printed
//		}
//
// # Toolchain
//
// Go 1.25 or later is required. The 'go.uber.org/goleak' package is used for tests built
// with `-tags=leaktest`, it is encouraged to run such test after making changes to the code base.
// This module requires 'go.uber.org/goleak' but, it is not neccesary for the library functions.
package crossbow

// Copyright 2026 Maciej "juan_em (JuanX-G)" Woźniak
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Or use the bundled copy in the LICENSE file.
