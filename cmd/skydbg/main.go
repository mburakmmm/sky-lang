package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	arg := os.Args[1]

	switch arg {
	case "version", "--version", "-v":
		fmt.Printf("skydbg version %s\n", version)
	case "help", "--help", "-h":
		printHelp()
	default:
		debugCommand(os.Args[1:])
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: skydbg <file> [options]

Options:
  --help, -h        Show this help
  --version, -v     Show version

Commands within debugger:
  break <line>      Set breakpoint
  run               Start execution
  step              Step one line
  continue          Continue execution
  print <var>       Print variable
  quit              Exit debugger
`)
}

func printHelp() {
	fmt.Print(`skydbg - SKY Debugger

A debugging adapter for SKY programs, providing integration
with LLDB/GDB for debugging compiled SKY code.

USAGE:
  skydbg <file> [options]

OPTIONS:
  --version, -v       Show version information
  --help, -h          Show this help message

DEBUGGER COMMANDS:
  break <line>        Set breakpoint at line
  break <func>        Set breakpoint at function
  run                 Start program execution
  step                Step to next line
  next                Step over function calls
  continue            Continue execution until breakpoint
  print <var>         Print variable value
  backtrace           Show call stack
  locals              Show local variables
  quit                Exit debugger

EXAMPLES:
  skydbg myprogram.sky           # Start debugging
  (skydbg) break 10              # Set breakpoint at line 10
  (skydbg) run                   # Start execution
  (skydbg) print x               # Print variable x

For more information, visit:
https://github.com/mburakmmm/sky-lang/docs/debug
`)
}

func debugCommand(args []string) {
	filename := args[0]
	fmt.Printf("Debugging: %s\n", filename)
	fmt.Println("Note: Debugger not yet implemented (Sprint 6)")
}

