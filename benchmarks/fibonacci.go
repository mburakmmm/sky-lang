package main

import "fmt"

func fib(n int) int {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func main() {
	n := 35
	fmt.Println("Computing fib(35)...")
	result := fib(n)
	fmt.Printf("Result: %d\n", result)
}

