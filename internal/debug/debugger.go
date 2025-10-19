package debug

// Debugger - Debugging support
// TODO: Implement LLDB/GDB bridge
// Bu dosya gelecekte debugging desteği sağlayacak

// Breakpoint bir breakpoint'i temsil eder
type Breakpoint struct {
	File string
	Line int
}

// Debugger debugger interface
type Debugger struct {
	breakpoints []Breakpoint
	running     bool
}

// NewDebugger yeni bir debugger oluşturur
func NewDebugger() *Debugger {
	return &Debugger{
		breakpoints: []Breakpoint{},
		running:     false,
	}
}

// AddBreakpoint breakpoint ekler
func (d *Debugger) AddBreakpoint(file string, line int) {
	d.breakpoints = append(d.breakpoints, Breakpoint{
		File: file,
		Line: line,
	})
}

// Run programı debug modunda çalıştırır
func (d *Debugger) Run(program string) error {
	// TODO: Implement LLDB/GDB integration
	d.running = true
	return nil
}

// Step bir adım ilerler
func (d *Debugger) Step() error {
	// TODO: Implement step
	return nil
}

// Continue çalışmaya devam eder
func (d *Debugger) Continue() error {
	// TODO: Implement continue
	return nil
}

// Stop durur
func (d *Debugger) Stop() {
	d.running = false
}

