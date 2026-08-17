package registry

import (
	"sync"

	. "github.com/JuanX-G/crossbow"
)

// hold a сancel func and a server pointer.
type cancelableServer[T ServerHandler[M, O], M, O any] struct {
	server *Server[T, M, O]
	cancel func()
}

// Registry holds a map of like cancelable servers keyed by the type K.
type Registry[K comparable, T ServerHandler[M, O], M, O any] struct {
	mu      sync.RWMutex
	servers map[K]*cancelableServer[T, M, O]
}

// Returns an empty registry.
func NewRegistry[K comparable, T ServerHandler[M, O], M any, O any]() *Registry[K, T, M, O] {
	return &Registry[K, T, M, O]{
		servers: make(map[K]*cancelableServer[T, M, O]),
	}
}

// Register a server in a registry. Cancel should be a function
// that cancels the srv argument. Returns an error if the key
// already exists or if cancel is nil.
func (s *Registry[K, T, M, O]) RegisterServer(cancel func(), k K, srv *Server[T, M, O]) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.servers[k]; ok {
		return ErrKeyTaken
	}
	if cancel == nil {
		return ErrCancelFuncNil
	}
	s.servers[k] = &cancelableServer[T, M, O]{server: srv, cancel: cancel}
	return nil
}

// Removes the server keyed by k from the registry.
// Calls the server's cancel func. False is retuend
// if the key does not exist.
func (s *Registry[K, T, M, O]) Remove(k K) bool {
	s.mu.Lock()
	managedSrv, ok := s.servers[k]
	delete(s.servers, k)
	s.mu.Unlock()
	if !ok {
		return false
	}
	managedSrv.cancel()
	return true
}

// Looks up the server under the key k. If the key
// does not exist, returned is nil and false.
func (s *Registry[K, T, M, O]) Resolve(k K) (*Server[T, M, O], bool) {
	s.mu.Lock()
	managedSrv, ok := s.servers[k]
	s.mu.Unlock()
	return managedSrv.server, ok // .server will be nil if the lookup failes and ok == false.
}
