package ringbuffer

import "sync"

type Buffer struct {
	data []byte
	head int
	tail int
	size int
	mu   sync.Mutex
}

func New(size int) *Buffer {
	return &Buffer{
		data: make([]byte, size),
		size: size,
		head: 0,
		tail: 0,
	}
}

func (rb *Buffer) getOccupied() int {
	if rb.head >= rb.tail {
		return rb.head - rb.tail
	}
	return (rb.size - rb.tail) + rb.head
}

func (rb *Buffer) Read(p []byte, c []byte) (ln int, buff []byte) {
	if rb.getOccupied() <= rb.size/4 {
		return 0, make([]byte, 0)
	}
	rb.mu.Lock()
	defer rb.mu.Unlock()
	n := 0
	for i := 0; i < len(p); i++ {
		if rb.tail == rb.head {
			break
		}

		c[i] = rb.data[rb.tail]

		rb.data[rb.tail] = 0

		rb.tail = (rb.tail + 1) % rb.size
		n++
	}

	return n, c
}

func (rb *Buffer) Write(data []byte) int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	n := 0
	for i := 0; i < len(data); i++ {
		if (rb.head+1)%rb.size == rb.tail {
			rb.tail = (rb.tail + 1) % rb.size
		}
		rb.data[rb.head] = data[i]
		n++
		rb.head = (rb.head + 1) % rb.size
	}
	return n
}
