package registry

import (
	"context"

	. "github.com/JuanX-G/crossbow"
)

// Send to a server under they key k. Returns ErrServerNotFound
// when the key does not exist.
func (s *Registry[K, T, M, O]) Send(ctx context.Context, k K, msg M) error {
	s.mu.RLock()
	managedSrv, ok := s.servers[k]
	s.mu.RUnlock()
	if !ok {
		return ErrServerNotFound
	}
	srv := managedSrv.server
	err := srv.Send(ctx, msg)
	return err
}

// See [crossbow/registry/Send]. this is the Call equivalent.
func (s *Registry[K, T, M, O]) Call(ctx context.Context, k K, msg M) (O, error) {
	s.mu.RLock()
	managedSrv, ok := s.servers[k]
	s.mu.RUnlock()
	if !ok {
		var zero O
		return zero, ErrServerNotFound
	}
	srv := managedSrv.server
	val, err := srv.Call(ctx, msg)
	return val, err
}

// Call send with the same ctx and msg on all servers.
func (s *Registry[K, T, M, O]) Broadcast(ctx context.Context, msg M) []error {
	s.mu.RLock()
	servers := make([]*Server[T, M, O], 0, len(s.servers))

	for _, managed := range s.servers {
		servers = append(servers, managed.server)
	}

	s.mu.RUnlock()

	out := make([]error, 0, len(servers))
	for _, srv := range servers {
		out = append(out, srv.Send(ctx, msg))
	}
	return out
}
