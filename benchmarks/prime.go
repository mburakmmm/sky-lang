package main

import (
	"fmt"
	"time"
)

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	
	return true
}

func countPrimes(limit int) int {
	count := 0
	for i := 2; i < limit; i++ {
		if isPrime(i) {
			count++
		}
	}
	return count
}

func main() {
	start := time.Now()
	result := countPrimes(10000)
	duration := time.Since(start)
	
	fmt.Printf("Go Primes up to 10000: %d\n", result)
	fmt.Printf("Duration: %v\n", duration)
}
