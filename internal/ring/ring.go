// Copyright 2026 Maciej "juan_em (JuanX-G)" Woźniak
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Or use the bundled copy in the LICENSE file.

package ring

type RingBuffer[T any] struct {
	cap   int
	buf   []T
	head  int
	tail  int
	count int
}

// Allocates a new RingBuffer. Defaults
// - If max capacity is less than 2 it defaults to 2.
// - If initialCapacity is less than 2 it defaults to 2.
// - If initialCapacity is above maxCapacity is it set to maxCapacity.
func NewRingBuffer[T any](initialCapacity int, maxCapacity int) *RingBuffer[T] {
	if initialCapacity < 2 {
		initialCapacity = 2
	}
	return &RingBuffer[T]{
		buf: make([]T, initialCapacity),
		cap: maxCapacity,
	}
}

// Pushes an element to the end of the buffer. Returns false if the push
// fails. A push can fail most notably because of the buffer hitting its cap.
// Push automatically doubles the buffer size if the cap allows so, if the new
// buffer size would overrun the cap, but there is still space below the cap,
// the buffer is grown to fill all that space.
func (r *RingBuffer[T]) Push(item T) bool {
	if r.count == len(r.buf) {
		newSize := len(r.buf) * 2
		if newSize > r.cap && r.cap != 0 && r.cap != len(r.buf) {
			deltaSize := r.cap - len(r.buf)
			r.resize(len(r.buf) + deltaSize)
		} else if newSize > r.cap && r.cap != 0 {
			return false
		} else {
			r.resize(newSize)
		}
	}

	r.buf[r.tail] = item
	r.tail = (r.tail + 1) % len(r.buf)
	r.count++
	return true
}

// Removes the next element of the buffer. Returns false if the element
// is missing, true if it's found. If flase is returned, the element will the
// zero value of T.
// Automatically halves the buffer size if the element count falls below
// 25% of the buffer size. As calculated by the following code:
// len(r.buf)/4.
func (r *RingBuffer[T]) Pop() (T, bool) {
	var zero T
	if r.count == 0 {
		return zero, false
	}

	item := r.buf[r.head]
	r.buf[r.head] = zero
	r.head = (r.head + 1) % len(r.buf)
	r.count--

	if r.count > 0 && r.count <= len(r.buf)/4 && len(r.buf) > 16 {
		r.resize(len(r.buf) / 2)
	}

	return item, true
}

// resize unwraps the buffer and copies it into a buffer of size newCapacity.
// It then swaps internal buffer pointer, sets head and tail.
func (r *RingBuffer[T]) resize(newCapacity int) {
	newBuf := make([]T, newCapacity)

	if r.count > 0 {
		if r.head < r.tail {
			copy(newBuf, r.buf[r.head:r.tail])
		} else {
			headLen := len(r.buf) - r.head
			copy(newBuf, r.buf[r.head:])
			copy(newBuf[headLen:], r.buf[:r.tail])
		}
	}

	r.buf = newBuf
	r.head = 0
	r.tail = r.count
}

// Returns the current count of elements in the buffer.
func (r *RingBuffer[T]) Len() int {
	return r.count
}
