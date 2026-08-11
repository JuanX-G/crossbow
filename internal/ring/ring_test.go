/* Crossbow, a simple go library for actor-like worker-pools with inboxes supporting parallel and synchronous processing.
 * Copyright (C) 2026 Maciej "juan_em" Woźniak, full license found in the LICENSE file
 */
package ring

import (
	"testing"
)

func TestPushPop(t *testing.T) {
	r := NewRingBuffer[int](2, 0)

	r.Push(1)
	r.Push(2)

	a, _ := r.Pop()
	b, _ := r.Pop()

	if a != 1 || b != 2 {
		t.Fatalf("pushed `1, 2` into ring buffer, returned was: a = %d; b = %d", a, b)
	}
}

func TestResize(t *testing.T) {
	r := NewRingBuffer[int](2, 0)
	for i := range 1000 {
		r.Push(i)
	}

	for i := range 1000 {
		v, _ := r.Pop()

		if v != i {
			t.Fatalf("buffer resizing failed, expcted: %d; found: %d", i, v)
		}
	}

}

func TestCap(t *testing.T) {
	r := NewRingBuffer[int](2, 5)
	r.Push(42)
	r.Push(42)
	r.Push(42) // r.count: 3, len(buf): 4
	if len(r.buf) != 4 {
		t.Fatalf("pushed 3 items into buffer of initial size: 2 and max capacity: 5, expected buffer length: 4, found: %d", len(r.buf))
	} else if r.count != 3 {
		t.Fatalf("pushed 3 items into buffer of initial size: 2 and max capacity: 5, expected count: 3, found: %d", r.count)
	}
	r.Push(42)
	if ok := r.Push(42); !ok {
		t.Fatalf("pushing 5th item into buffer of initial size: 2 and max capacity: 5 failed. Buffer count: %d, buffer length: %d", r.count, len(r.buf))
	}
	if ok := r.Push(2137); ok {
		t.Fatalf("pushing 5th item into buffer of initial size: 2 and max capacity: 5 succeded. Buffer state: count: %d, buffer length: %d", r.count, len(r.buf))
	}
}
