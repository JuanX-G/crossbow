package registry

import (
	"context"
	"strconv"
	"testing"

	"github.com/JuanX-G/crossbow"
)

// Example of a simple handler.
type Echo struct {
	Mark     string
	MarkHook func(string)
}

func (e *Echo) Handle(msg crossbow.ContextMessage[int, string]) (string, error) {
	e.MarkHook(e.Mark)
	return strconv.Itoa(msg.Value), nil
}

// Default MarkHook to a no-op when not set
func (e *Echo) Init() error {
	f := func(string) {}
	if e.MarkHook == nil {
		e.MarkHook = f
	}
	return nil
}

func (Echo) Terminate(err error) {}

// Test new registry generation
func TestNewRegistry(t *testing.T) {
	reg := NewRegistry[string, *Echo]()
	if reg.servers == nil {
		t.Fatalf("new registry starts with a nil map")
	}
}

func TestRegistration(t *testing.T) {
	reg := NewRegistry[string, *Echo]()
	hndl := &Echo{}
	cfg := crossbow.MakeDefaultServerConfig()
	srv, _ := crossbow.NewServer(hndl, cfg, crossbow.DefaultPanicRecover[*Echo, int, string])

	ctx, cancel := context.WithCancel(t.Context())
	go srv.Run(ctx)

	// regular registration
	err := reg.RegisterServer(cancel, "SRV", srv)
	if err != nil {
		t.Fatalf("attempted to register server: %v, key: %s. Error: %s was returned", srv, "SRV", err.Error())
	}

	// registering under a taken key
	err = reg.RegisterServer(cancel, "SRV", srv)
	if err == nil {
		t.Fatalf("attempted to register server under a already registered key, nil error returned: server %v, key: %s", srv, "SRV")
	}

	// registering nil cancel func
	err = reg.RegisterServer(nil, "tmpSRV", srv)
	if err == nil {
		t.Fatalf("attempted to register server with a nil cancel func, nil error returned: server %v, key: %s", srv, "tmpSRV")
	}
}

func TestResolving(t *testing.T) {
	var mark string
	markHook := func(s string) {
		mark = s
	}

	expectedMark := "2137"
	hndl := &Echo{
		Mark:     expectedMark,
		MarkHook: markHook,
	}

	reg := NewRegistry[string, *Echo]()
	cfg := crossbow.MakeDefaultServerConfig()
	srv, _ := crossbow.NewServer(hndl, cfg, crossbow.DefaultPanicRecover[*Echo, int, string])

	ctx, cancel := context.WithCancel(t.Context())
	go srv.Run(ctx)

	key := "SRV"
	_ = reg.RegisterServer(cancel, key, srv)

	res, ok := reg.Resolve(key)
	if !ok {
		t.Fatalf("registered a server under the key %s, did not resolve with the same key.", key)
	}

	// check if the same server is retrived
	_, _ = res.Call(ctx, 0)
	if mark != expectedMark {
		t.Fatalf("expected the server to set mark: %s, found: %s", expectedMark, mark)
	}

	notsrv, ok := reg.Resolve(key + "NOT")
	if ok {
		t.Fatalf("tried to resolve a sever at key not registering one before, found ok == true, key used: %s, server returned: %v", key, notsrv)
	} else if notsrv != nil {
		t.Fatalf("tried to resolve a sever at key not registering one before, found server != nil, key: %s, server: %v", key, notsrv)
	}
}

func TestRemoving(t *testing.T) {
	hndl := &Echo{}
	reg := NewRegistry[string, *Echo]()
	cfg := crossbow.MakeDefaultServerConfig()
	srv, _ := crossbow.NewServer(hndl, cfg, crossbow.DefaultPanicRecover[*Echo, int, string])

	ctx, cancel := context.WithCancel(t.Context())
	go srv.Run(ctx)

	canceled := false
	cf := func() {
		canceled = true
		cancel()
	}
	key := "SRV"
	err := reg.RegisterServer(cf, key, srv)
	if err != nil {
		t.Fatalf("error: %s returned at server registartion", err.Error())
	}

	ok := reg.Remove(key)
	if !ok {
		t.Fatalf("failed to remove server at key: %s, but registration succeeded", key)
	}

	if !canceled {
		t.Fatalf("removing server did not call cancel")
	}

	ok = reg.Remove(key + "NOT")
	if ok {
		t.Fatalf("tried to remove a sever at key not registering one before, found ok == true, key used: %s", key)
	}
}
