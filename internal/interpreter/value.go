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
}

func (f *Function) Kind() ValueKind { return FunctionValue }
func (f *Function) String() string  { return fmt.Sprintf("<function %s>", f.Name) }
func (f *Function) IsTruthy() bool  { return true }

// RuntimeError runtime hatalarını temsil eder
type RuntimeError struct {
	Message string
}

func (e *RuntimeError) Error() string {
	return e.Message
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
