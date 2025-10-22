package main

import (
	"fmt"
	"time"
)

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
	start := time.Now()
	result := fibonacci(35)
	duration := time.Since(start)

	fmt.Printf("Go Fibonacci(35) = %d\n", result)
	fmt.Printf("Duration: %v\n", duration)
}
