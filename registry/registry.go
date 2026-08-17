package registry

import (
	"sync"

	. "github.com/JuanX-G/crossbow"
)

type managedServer[T ServerHandler[M, O], M, O any] struct {
	server *Server[T, M, O]
	cancel func()
}

type Registry[K comparable, T ServerHandler[M, O], M, O any] struct {
	mu      sync.RWMutex
	servers map[K]*managedServer[T, M, O]
}

func NewRegistry[K comparable, T ServerHandler[M, O], M any, O any]() *Registry[K, T, M, O] {
	return &Registry[K, T, M, O]{
		servers: make(map[K]*managedServer[T, M, O]),
	}
}

func (s *Registry[K, T, M, O]) RegisterServer(cancel func(), k K, srv *Server[T, M, O]) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.servers[k]; ok {
		return ErrKeyTaken
	}
	if cancel == nil {
		return ErrCancelFuncNil
	}
	s.servers[k] = &managedServer[T, M, O]{server: srv, cancel: cancel}
	return nil
}

func (s *Registry[K, T, M, O]) Remove(k K) bool {
	s.mu.Lock()
	managedSrv, ok := s.servers[k]
	delete(s.servers, k)
	s.mu.Unlock()
	if !ok {
		return ok
	}
	managedSrv.cancel()
	return true
}

func (s *Registry[K, T, M, O]) Resolve(k K) (*Server[T, M, O], bool) {
	s.mu.Lock()
	managedSrv, ok := s.servers[k]
	s.mu.Unlock()
	if !ok {
		return nil, ok
	}
	return managedSrv.server, true
}
