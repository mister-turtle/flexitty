package tty

import (
	"errors"
	"sync"
)

var (
	ErrEmpty = errors.New("Circle buffer was empty")
)

// CircleBuffer struct to track the buffer, head and tail pointers, and provide a mutex for safe concurrent access
type CircleBuffer struct {
	buf     []byte
	size    int
	head    int
	tail    int
	written bool
	mu      sync.Mutex
}

// Write implements a naive write, iterating over the incoming bytes and writing them to the buffer.
// It returns the number of bytes written.
// The current implementation will continiously write round the buffer on over-size writes, by design.
func (c *CircleBuffer) Write(data []byte) (int, error) {

	c.written = true
	for _, b := range data {
		c.tail = (c.tail + 1) % c.size

		if c.tail == c.head {
			c.head = (c.head + 1) % c.size
		}
		c.buf[c.tail] = b

	}
	return len(data), nil
}

// Read will read the full data from the head to tail of the buffer, returning the length of data read.
// Note this will not increment the head. This design allows the buffer to be used as a "historical log" as such.
func (c *CircleBuffer) Read(p []byte) (int, error) {

	if c.head == c.tail && !c.written {
		return 0, ErrEmpty
	}

	readPos := c.head
	read := 0
	readTemp := make([]byte, c.size)
	for {
		readTemp[read] = c.buf[readPos]
		read++
		if readPos == c.tail {
			break
		}
		readPos = (readPos + 1) % (c.size)
	}

	copy(p, readTemp)

	return read, nil
}

// New takes a size of the buffer to create and returns a pointer to the new Circle Buffer.
func NewCircleBuffer(size int) *CircleBuffer {
	return &CircleBuffer{
		buf:  make([]byte, size),
		size: size,
		tail: -1,
	}
}
