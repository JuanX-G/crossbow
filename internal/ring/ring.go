package ring

type RingBuffer[T any] struct {
	cap   int
	buf   []T
	head  int
	tail  int
	count int
}

func NewRingBuffer[T any](initialCapacity int, maxCapacity int) *RingBuffer[T] {
	if initialCapacity < 2 {
		initialCapacity = 2
	}
	return &RingBuffer[T]{
		buf: make([]T, initialCapacity),
		cap: maxCapacity,
	}
}

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

func (r *RingBuffer[T]) Len() int {
	return r.count
}
