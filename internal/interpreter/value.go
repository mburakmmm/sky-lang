package interpreter

import "fmt"

// ValueKind değer tiplerini belirtir
type ValueKind int

const (
	IntValue ValueKind = iota
	FloatValue
	StringValue
	BoolValue
	NilValue
	ListValue
	DictValue
	FunctionValue
	PromiseValue
	ClassValue
	InstanceValue
)

// Value runtime değerlerini temsil eder
type Value interface {
	Kind() ValueKind
	String() string
	IsTruthy() bool
}

// Integer değer
type Integer struct {
	Value int64
}

func (i *Integer) Kind() ValueKind { return IntValue }
func (i *Integer) String() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) IsTruthy() bool  { return i.Value != 0 }

// Float değer
type Float struct {
	Value float64
}

func (f *Float) Kind() ValueKind { return FloatValue }
func (f *Float) String() string  { return fmt.Sprintf("%f", f.Value) }
func (f *Float) IsTruthy() bool  { return f.Value != 0.0 }

// String değer
type String struct {
	Value string
}

func (s *String) Kind() ValueKind { return StringValue }
func (s *String) String() string  { return s.Value }
func (s *String) IsTruthy() bool  { return s.Value != "" }

// Boolean değer
type Boolean struct {
	Value bool
}

func (b *Boolean) Kind() ValueKind { return BoolValue }
func (b *Boolean) String() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) IsTruthy() bool  { return b.Value }

// Nil değer
type Nil struct{}

func (n *Nil) Kind() ValueKind { return NilValue }
func (n *Nil) String() string  { return "nil" }
func (n *Nil) IsTruthy() bool  { return false }

// List değer
type List struct {
	Elements []Value
}

func (l *List) Kind() ValueKind { return ListValue }
func (l *List) String() string {
	result := "["
	for i, elem := range l.Elements {
		if i > 0 {
			result += ", "
		}
		result += elem.String()
	}
	result += "]"
	return result
}
func (l *List) IsTruthy() bool { return len(l.Elements) > 0 }

// Dict değer
type Dict struct {
	Pairs map[string]Value
}

func (d *Dict) Kind() ValueKind { return DictValue }
func (d *Dict) String() string {
	result := "{"
	first := true
	for key, value := range d.Pairs {
		if !first {
			result += ", "
		}
		result += fmt.Sprintf("%s: %s", key, value.String())
		first = false
	}
	result += "}"
	return result
}
func (d *Dict) IsTruthy() bool { return len(d.Pairs) > 0 }

// Function değer
type Function struct {
	Name       string
	Parameters []string
	Body       func(*Environment) (Value, error)
	Env        *Environment
	Async      bool // async function flag
}

func (f *Function) Kind() ValueKind { return FunctionValue }
func (f *Function) String() string {
	if f.Async {
		return fmt.Sprintf("<async function %s>", f.Name)
	}
	return fmt.Sprintf("<function %s>", f.Name)
}
func (f *Function) IsTruthy() bool { return true }

// RuntimeError runtime hatalarını temsil eder
type RuntimeError struct {
	Message string
}

func (e *RuntimeError) Error() string {
	return e.Message
}

// BreakSignal signals a break statement execution
type BreakSignal struct{}

func (b *BreakSignal) Error() string {
	return "break"
}

// ContinueSignal signals a continue statement execution
type ContinueSignal struct{}

func (c *ContinueSignal) Error() string {
	return "continue"
}

// Environment değişken ortamını temsil eder
type Environment struct {
	store  map[string]Value
	parent *Environment
}

// NewEnvironment yeni bir environment oluşturur
func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		store:  make(map[string]Value),
		parent: parent,
	}
}

// Get değişken değerini alır
func (e *Environment) Get(name string) (Value, bool) {
	if val, ok := e.store[name]; ok {
		return val, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

// Set değişken değerini ayarlar
func (e *Environment) Set(name string, value Value) {
	e.store[name] = value
}

// Update var olan değişkeni günceller
func (e *Environment) Update(name string, value Value) error {
	if _, ok := e.store[name]; ok {
		e.store[name] = value
		return nil
	}
	if e.parent != nil {
		return e.parent.Update(name, value)
	}
	return &RuntimeError{Message: fmt.Sprintf("undefined variable: %s", name)}
}

// Promise represents an async value that will be resolved later
type Promise struct {
	State    string // "pending", "fulfilled", "rejected"
	Value    Value  // resolved value
	Error    error  // rejection reason
	executor func() (Value, error)
}

func (p *Promise) Kind() ValueKind { return PromiseValue }
func (p *Promise) String() string {
	switch p.State {
	case "fulfilled":
		return fmt.Sprintf("<Promise resolved: %v>", p.Value)
	case "rejected":
		return fmt.Sprintf("<Promise rejected: %v>", p.Error)
	default:
		return "<Promise pending>"
	}
}
func (p *Promise) IsTruthy() bool { return p.State == "fulfilled" }

// NewPromise creates a new promise with an executor function
func NewPromise(executor func() (Value, error)) *Promise {
	p := &Promise{
		State:    "pending",
		executor: executor,
	}
	// Execute immediately for now (will be async with event loop)
	go func() {
		result, err := executor()
		if err != nil {
			p.State = "rejected"
			p.Error = err
		} else {
			p.State = "fulfilled"
			p.Value = result
		}
	}()
	return p
}

// Await blocks until the promise is resolved and returns the value
func (p *Promise) Await() (Value, error) {
	// Simple busy-wait (will be improved with event loop)
	for p.State == "pending" {
		// In production, this would yield to event loop
	}
	if p.State == "rejected" {
		return nil, p.Error
	}
	return p.Value, nil
}

// Class represents a class definition
type Class struct {
	Name       string
	SuperClass *Class               // parent class
	Methods    map[string]*Function // method name -> function
	Env        *Environment         // class environment
}

func (c *Class) Kind() ValueKind { return ClassValue }
func (c *Class) String() string  { return fmt.Sprintf("<class %s>", c.Name) }
func (c *Class) IsTruthy() bool  { return true }

// Instance represents an instance of a class
type Instance struct {
	Class  *Class
	Fields map[string]Value
}

func (i *Instance) Kind() ValueKind { return InstanceValue }
func (i *Instance) String() string {
	return fmt.Sprintf("<instance of %s>", i.Class.Name)
}
func (i *Instance) IsTruthy() bool { return true }

// Get retrieves a field or method from the instance
func (i *Instance) Get(name string) (Value, bool) {
	// Check instance fields first
	if val, ok := i.Fields[name]; ok {
		return val, true
	}

	// Check class methods
	if method, ok := i.Class.Methods[name]; ok {
		// Bind method to instance (self)
		boundMethod := &Function{
			Name:       method.Name,
			Parameters: method.Parameters,
			Async:      method.Async,
			Env:        method.Env,
			Body: func(callEnv *Environment) (Value, error) {
				// Set 'self' to this instance
				callEnv.Set("self", i)
				return method.Body(callEnv)
			},
		}
		return boundMethod, true
	}

	// Check superclass methods
	if i.Class.SuperClass != nil {
		if method, ok := i.Class.SuperClass.Methods[name]; ok {
			boundMethod := &Function{
				Name:       method.Name,
				Parameters: method.Parameters,
				Async:      method.Async,
				Env:        method.Env,
				Body: func(callEnv *Environment) (Value, error) {
					callEnv.Set("self", i)
					callEnv.Set("super", i.Class.SuperClass)
					return method.Body(callEnv)
				},
			}
			return boundMethod, true
		}
	}

	return nil, false
}

// Set sets a field on the instance
func (i *Instance) Set(name string, value Value) {
	i.Fields[name] = value
}
