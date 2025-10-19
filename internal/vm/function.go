package vm

// CompiledFunction represents a compiled function with its bytecode
type CompiledFunction struct {
	Name         string
	Arity        int  // number of parameters
	Async        bool // async function flag
	Instructions []Instruction
	Constants    []interface{}
	LocalCount   int // number of local variables
}

// NewCompiledFunction creates a new compiled function
func NewCompiledFunction(name string, arity int) *CompiledFunction {
	return &CompiledFunction{
		Name:         name,
		Arity:        arity,
		Instructions: make([]Instruction, 0),
		Constants:    make([]interface{}, 0),
		LocalCount:   0,
	}
}
