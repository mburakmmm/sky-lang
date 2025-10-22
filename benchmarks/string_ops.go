package main

import (
	"fmt"
	"strings"
	"time"
)

func stringOperations() {
	text := "Hello World from SKY Programming Language"
	iterations := 100000
	
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		upper := strings.ToUpper(text)
		lower := strings.ToLower(text)
		replaced := strings.ReplaceAll(text, "SKY", "GO")
		joined := strings.Join([]string{"a", "b", "c"}, "-")
		splitResult := strings.Split(text, " ")
		_ = upper
		_ = lower
		_ = replaced
		_ = joined
		_ = splitResult
	}
	
	duration := time.Since(start)
	
	fmt.Printf("Go String operations (%d iterations)\n", iterations)
	fmt.Printf("Duration: %v\n", duration)
}

func main() {
	stringOperations()
}
