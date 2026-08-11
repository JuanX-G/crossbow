/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package crossbow

import (
	"context"
	"fmt"
	"runtime/debug"
)

func (s *Server[T, M, O]) enqueue(msg ContextMessage[M, O]) error {
	if s.terminated.Load() {
		return ErrServerTerminated
	}
	if err := s.queue.Push(msg.Context, msg); err != nil {
		return err
	}
	return nil
}

func (s *Server[T, M, O]) dispatch(msg ContextMessage[M, O]) {
	var zero O
	var output O
	var err error
	var panicked bool

	defer func() {
		if r := recover(); r != nil {
			panicked = true
			s.stats.AddPanic()
			panicErr := fmt.Errorf("Panicked: %v", r)
			err = panicErr
			if recErr := s.recover(msg, s.handler, panicErr, debug.Stack()); recErr != nil {
				s.shutdown(recErr)
			}
		}

		if msg.reply != nil {
			if panicked {
				msg.reply <- Response[O]{Value: zero, Err: err}
			} else {
				if err != nil {
					s.stats.AddFail()
				}
				msg.reply <- Response[O]{Value: output, Err: err}
			}
		}
	}()

	if !s.terminated.Load() {
		output, err = s.handler.Handle(msg)
	}
}

func (s *Server[T, M, O]) shutdown(reason error) {
	if s.terminated.CompareAndSwap(false, true) {
		s.terminatedCh <- struct{}{}
		close(s.terminatedCh)
		s.queue.Close()
		s.handler.Terminate(reason)
	}
}

func (s *Server[T, M, O]) processQueue(ctx context.Context) error {
	for {
		if s.terminated.Load() {
			return ErrServerTerminated
		}

		msg, ok := s.queue.Pop()
		if !ok {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			s.dispatch(msg)
		}
	}
	return nil
}

func (s *Server[T, M, O]) drainRemaining() {
	var zero O
	for {
		msg, ok := s.queue.Pop()
		if !ok {
			break
		}
		if msg.reply != nil {
			msg.reply <- Response[O]{Value: zero, Err: ErrServerTerminated}
		}
	}
}

// # Public API for Server

// Run sets up the server's main loop and defers cleanup functions. It must be run before sending anything
// to the server. It returns upon context cancelation or s.terminated being true or the queues.Notify()
// reporting the underlying channel to be closed. The main loop awaits new data and wakes up when
// the queue notifies it.
func (s *Server[T, M, O]) Run(ctx context.Context) {
	workerCount := int(s.workers)
	for range workerCount {
		s.wg.Add(1)
		go s.worker(ctx)
	}
	go func() {
		select {
		case <-ctx.Done():
			s.stop()
		case <- s.terminatedCh:
			return
		}
	}()
}

func (s *Server[T, M, O]) stop() {
	s.shutdown(nil)
	s.drainRemaining()
	s.wg.Wait()
}
