package skylib

import (
	"fmt"
	"runtime"
)

// SkyError represents a runtime error with stacktrace
type SkyError struct {
	message    string
	stacktrace []string
	cause      error
}

// NewError creates a new error
func NewError(message string) *SkyError {
	return &SkyError{
		message:    message,
		stacktrace: captureStacktrace(),
		cause:      nil,
	}
}

// Error implements error interface
func (e *SkyError) Error() string {
	return e.message
}

// Stacktrace returns the stacktrace
func (e *SkyError) Stacktrace() []string {
	return e.stacktrace
}

// Cause returns the underlying error
func (e *SkyError) Cause() error {
	return e.cause
}

// captureStacktrace captures current stacktrace
func captureStacktrace() []string {
	const maxDepth = 32
	pcs := make([]uintptr, maxDepth)
	n := runtime.Callers(3, pcs) // Skip runtime.Callers, captureStacktrace, NewError

	frames := runtime.CallersFrames(pcs[:n])
	stacktrace := []string{}

	for {
		frame, more := frames.Next()
		stacktrace = append(stacktrace, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}

	return stacktrace
}

// Specific error types

// IOError represents an I/O error
type IOError struct {
	SkyError
}

func NewIOError(message string) *IOError {
	return &IOError{
		SkyError: SkyError{
			message:    "IOError: " + message,
			stacktrace: captureStacktrace(),
		},
	}
}

// ValueError represents a value error
type ValueError struct {
	SkyError
}

func NewValueError(message string) *ValueError {
	return &ValueError{
		SkyError: SkyError{
			message:    "ValueError: " + message,
			stacktrace: captureStacktrace(),
		},
	}
}

// KeyError represents a key not found error
type KeyError struct {
	SkyError
	Key string
}

func NewKeyError(key string) *KeyError {
	return &KeyError{
		SkyError: SkyError{
			message:    "KeyError: " + key,
			stacktrace: captureStacktrace(),
		},
		Key: key,
	}
}

// TypeError represents a type error
type TypeError struct {
	SkyError
	Expected string
	Got      string
}

func NewTypeError(expected, got string) *TypeError {
	return &TypeError{
		SkyError: SkyError{
			message:    fmt.Sprintf("TypeError: expected %s, got %s", expected, got),
			stacktrace: captureStacktrace(),
		},
		Expected: expected,
		Got:      got,
	}
}

// IndexError represents an index out of bounds error
type IndexError struct {
	SkyError
	Index int
	Size  int
}

func NewIndexError(index, size int) *IndexError {
	return &IndexError{
		SkyError: SkyError{
			message:    fmt.Sprintf("IndexError: index %d out of range [0:%d)", index, size),
			stacktrace: captureStacktrace(),
		},
		Index: index,
		Size:  size,
	}
}

// RuntimeError represents a generic runtime error
type RuntimeError struct {
	SkyError
}

func NewRuntimeError(message string) *RuntimeError {
	return &RuntimeError{
		SkyError: SkyError{
			message:    "RuntimeError: " + message,
			stacktrace: captureStacktrace(),
		},
	}
}

// Helper functions

// Raise raises an error
func ErrorRaise(err *SkyError) {
	panic(err)
}

// Try executes a function and catches errors
func ErrorTry(fn func() interface{}) (interface{}, *SkyError) {
	var result interface{}
	var err *SkyError

	func() {
		defer func() {
			if r := recover(); r != nil {
				if skyErr, ok := r.(*SkyError); ok {
					err = skyErr
				} else {
					runtimeErr := NewRuntimeError(fmt.Sprintf("%v", r))
					err = &runtimeErr.SkyError
				}
			}
		}()
		result = fn()
	}()

	return result, err
}
