package interpreter

import "fmt"

// CallFrame represents a single function call in our custom stack
type CallFrame struct {
	FuncName string
	Args     []Value
	Env      *Environment
	Func     func(*Environment) (Value, error)
}

// TrampolineStack manages function calls without using Go's call stack
type TrampolineStack struct {
	frames      []*CallFrame
	maxDepth    int
	resultCache map[string]Value // Memoization cache
}

// NewTrampolineStack creates a new trampoline stack
func NewTrampolineStack(maxDepth int) *TrampolineStack {
	return &TrampolineStack{
		frames:      make([]*CallFrame, 0, 1024),
		maxDepth:    maxDepth,
		resultCache: make(map[string]Value),
	}
}

// Push adds a new call frame
func (ts *TrampolineStack) Push(frame *CallFrame) error {
	if len(ts.frames) >= ts.maxDepth {
		return &RuntimeError{
			Message: fmt.Sprintf("stack overflow: maximum recursion depth (%d) exceeded", ts.maxDepth),
		}
	}
	ts.frames = append(ts.frames, frame)
	return nil
}

// Pop removes the top call frame
func (ts *TrampolineStack) Pop() *CallFrame {
	if len(ts.frames) == 0 {
		return nil
	}
	frame := ts.frames[len(ts.frames)-1]
	ts.frames = ts.frames[:len(ts.frames)-1]
	return frame
}

// Depth returns current stack depth
func (ts *TrampolineStack) Depth() int {
	return len(ts.frames)
}

// GetCached checks if result is cached
func (ts *TrampolineStack) GetCached(funcName string, args []Value) (Value, bool) {
	key := ts.makeCacheKey(funcName, args)
	val, ok := ts.resultCache[key]
	return val, ok
}

// SetCached stores result in cache
func (ts *TrampolineStack) SetCached(funcName string, args []Value, result Value) {
	key := ts.makeCacheKey(funcName, args)
	ts.resultCache[key] = result
}

// makeCacheKey creates a cache key from function name and arguments
func (ts *TrampolineStack) makeCacheKey(funcName string, args []Value) string {
	key := funcName + ":"
	for i, arg := range args {
		if i > 0 {
			key += ","
		}
		// Simple string representation for cacheable types
		switch v := arg.(type) {
		case *Integer:
			key += fmt.Sprintf("%d", v.Value)
		case *Float:
			key += fmt.Sprintf("%g", v.Value)
		case *Boolean:
			key += fmt.Sprintf("%t", v.Value)
		case *String:
			if len(v.Value) < 50 { // Only cache short strings
				key += v.Value
			} else {
				return "" // Don't cache large strings
			}
		default:
			return "" // Don't cache complex types
		}
	}
	return key
}

// ClearCache clears the memoization cache
func (ts *TrampolineStack) ClearCache() {
	ts.resultCache = make(map[string]Value)
}
