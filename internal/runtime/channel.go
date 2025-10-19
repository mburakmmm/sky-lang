package runtime

import (
	"fmt"
	"sync"
)

// Channel represents a Go-style channel
type Channel struct {
	buffer   []interface{}
	capacity int
	closed   bool
	mu       sync.Mutex
	sendQ    []chan struct{}
	recvQ    []chan interface{}
}

// NewChannel creates a new channel
func NewChannel(capacity int) *Channel {
	return &Channel{
		buffer:   make([]interface{}, 0, capacity),
		capacity: capacity,
		closed:   false,
		sendQ:    []chan struct{}{},
		recvQ:    []chan interface{}{},
	}
}

// Send sends a value to the channel
func (c *Channel) Send(value interface{}) error {
	c.mu.Lock()

	if c.closed {
		c.mu.Unlock()
		return fmt.Errorf("send on closed channel")
	}

	// If there's a waiting receiver, send directly
	if len(c.recvQ) > 0 {
		recvCh := c.recvQ[0]
		c.recvQ = c.recvQ[1:]
		c.mu.Unlock()
		recvCh <- value
		return nil
	}

	// If buffer has space, add to buffer
	if len(c.buffer) < c.capacity {
		c.buffer = append(c.buffer, value)
		c.mu.Unlock()
		return nil
	}

	// Block until receiver or close
	sendCh := make(chan struct{})
	c.sendQ = append(c.sendQ, sendCh)
	c.mu.Unlock()

	<-sendCh

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return fmt.Errorf("send on closed channel")
	}
	c.buffer = append(c.buffer, value)
	c.mu.Unlock()

	return nil
}

// Receive receives a value from the channel
func (c *Channel) Receive() (interface{}, bool, error) {
	c.mu.Lock()

	// If buffer has data, return it
	if len(c.buffer) > 0 {
		value := c.buffer[0]
		c.buffer = c.buffer[1:]

		// Wake up a sender if any
		if len(c.sendQ) > 0 {
			sendCh := c.sendQ[0]
			c.sendQ = c.sendQ[1:]
			c.mu.Unlock()
			close(sendCh)
			return value, true, nil
		}

		c.mu.Unlock()
		return value, true, nil
	}

	// If closed and empty, return
	if c.closed {
		c.mu.Unlock()
		return nil, false, nil
	}

	// Block until sender or close
	recvCh := make(chan interface{})
	c.recvQ = append(c.recvQ, recvCh)
	c.mu.Unlock()

	value, ok := <-recvCh
	return value, ok, nil
}

// Close closes the channel
func (c *Channel) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	c.closed = true

	// Wake up all waiting receivers
	for _, recvCh := range c.recvQ {
		close(recvCh)
	}
	c.recvQ = nil
}

// Len returns the number of elements in the buffer
func (c *Channel) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.buffer)
}

// Cap returns the capacity of the channel
func (c *Channel) Cap() int {
	return c.capacity
}
