# Crossbow
Is a simple actor-like, inbox based worker-pool. 
It provides: lightweight lifecycle managment, panic recovery, adaptable mailbox policies and resizing, 
synchronous and asynchronous communications with the server; optional handler level parallelism 
and classic FIFO processing as default.

Crossbow makes heavy use of generics to keep whole thing type safe. It is kept lean and lightweight on purpose.
That is why it lacks a central registry, managers or process trees. 

## Versioning
Crossbow is intended to strcitly follow SemVer from version v0.1 onwards. Releases marked 
as v0-alpha.x.y or v0-beta.x.y do not make any compatibility guarantees.
We intends to use the standard go support cycle, each version of crossbow will be 
maintained for two version of the Go language.

## Usage
All examples can be found in ./examples

To use Crossbow you need to defined a handler that implements the ServerHandler[M any, O any] interface. 
Mainly you need to have a Handle(ContextMessage[M, O]) method, this is the handling loop of your server;
any data sent ot it is popped from the inbox and passed to it.
```go
(...)

// Simple stateless handler that send back data through the provided channel
type Foo struct{
    x int
}

func (e *Foo) Handle(msg ContextMessage[int, int]) (int, error) {
    if e.x == 2137 {
        fmt.Println("21:37")
        e.x = 0
        return 0, nil
    } else if msg.ID != "" { // ContextMessage carries related data for use by the handler (see message.go for details)
        return 0, fmt.Errorf("message cannot have an ID!") // This error will be passed on to anyone using the synchronous Call() method
    } else {
        e.x = msg.Value // Update internal state. Safe if the thread count is 0
        return msg.Value, nil
    }
}

// Simple init stub. Usually used to initialize connection and validate fields
func (e EchoSend) Init() error {
	return nil
}

// Handles the closing of connections etc.
func (e EchoSend) Terminate(err error) {
    if e.x == 42 {
        fmt.Println("Ah! 42")
    }
}

func main() {
    handler := &Foo{}
    cfg := MakeDefaultServerConfig[*Foo](1)
    srv, _ := NewServer(handler, cfg)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go srv.Run(ctx) // this must be done before sending anything to the server

    rqCtx, _ := context.WithCancel(ctx) // canceling a requests context won't stop the whole server
    _ = srv.Send(rqCtx, 2137) // Send() does not block awaiting a response and just discards any

    out, err := srv.Call(rqCtx, 11)
    // Output will be 0; err will be nil; and the string "21:37" will be printed
}
    
```
