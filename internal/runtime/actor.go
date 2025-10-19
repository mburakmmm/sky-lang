package runtime

import (
	"sync"
)

// Actor represents an actor with a mailbox
type Actor struct {
	id      string
	mailbox *Channel
	handler func(interface{}) interface{}
	running bool
	mu      sync.Mutex
}

// NewActor creates a new actor
func NewActor(id string, mailboxSize int, handler func(interface{}) interface{}) *Actor {
	return &Actor{
		id:      id,
		mailbox: NewChannel(mailboxSize),
		handler: handler,
		running: false,
	}
}

// Start starts the actor
func (a *Actor) Start() {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return
	}
	a.running = true
	a.mu.Unlock()
	
	go a.run()
}

// Send sends a message to the actor
func (a *Actor) Send(msg interface{}) error {
	return a.mailbox.Send(msg)
}

// Stop stops the actor
func (a *Actor) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if !a.running {
		return
	}
	
	a.running = false
	a.mailbox.Close()
}

// run is the actor's message loop
func (a *Actor) run() {
	for {
		msg, ok, err := a.mailbox.Receive()
		if err != nil || !ok {
			break
		}
		
		if a.handler != nil {
			a.handler(msg)
		}
	}
}

// ID returns the actor's ID
func (a *Actor) ID() string {
	return a.id
}

