//go:build !llvm

package main

import "fmt"

func runWithJIT(filename string) error {
	return fmt.Errorf("JIT mode not available: build with -tags llvm")
}
