package registry

import (
	"context"

	. "github.com/JuanX-G/crossbow"
)

func (s *Registry[K, T, M, O]) Send(ctx context.Context, k K, msg M) (bool, error) {
	s.mu.RLock()
	managedSrv, ok := s.servers[k]
	s.mu.RUnlock()
	if !ok {
		return ok, nil
	}
	srv := managedSrv.server
	err := srv.Send(ctx, msg)
	return true, err
}

func (s *Registry[K, T, M, O]) Call(ctx context.Context, k K, msg M) (O, bool, error) {
	s.mu.RLock()
	managedSrv, ok := s.servers[k]
	s.mu.RUnlock()
	if !ok {
		var zero O
		return zero, ok, nil
	}
	srv := managedSrv.server
	val, err := srv.Call(ctx, msg)
	return val, true, err
}

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
