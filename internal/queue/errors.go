package queue

import "errors"

var (
	ErrClosed  = errors.New("queue is closed")
	ErrFull    = errors.New("queue is full")
	ErrDropped = errors.New("message dropped")
)
