package runtime

import (
	"sync"
)

// SelectCase represents a case in a select statement
type SelectCase struct {
	Channel   *Channel
	IsSend    bool
	SendValue interface{}
	Handler   func(interface{}) error
}

// Select implements channel multiplexing
type Select struct {
	cases  []*SelectCase
	result chan int // Index of the case that was selected
	mu     sync.Mutex
}

// NewSelect creates a new select multiplexer
func NewSelect(cases []*SelectCase) *Select {
	return &Select{
		cases:  cases,
		result: make(chan int, 1),
	}
}

// Execute runs the select statement
func (s *Select) Execute() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Try all cases in order (non-blocking)
	for i, c := range s.cases {
		if c.IsSend {
			// Try to send
			select {
			case <-c.Channel.sendQ[0]:
				if err := c.Channel.Send(c.SendValue); err == nil {
					return i, nil
				}
			default:
				continue
			}
		} else {
			// Try to receive
			value, ok, err := c.Channel.Receive()
			if ok && err == nil {
				if c.Handler != nil {
					if err := c.Handler(value); err != nil {
						return i, err
					}
				}
				return i, nil
			}
		}
	}

	// If no case is ready, block until one becomes ready
	// This is a simplified implementation
	// Real implementation would use goroutines and select

	return -1, nil
}

// AddCase adds a case to the select
func (s *Select) AddCase(c *SelectCase) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cases = append(s.cases, c)
}
