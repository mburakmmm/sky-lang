package loopgo

import "fmt"

func RunBenchmark() {
	sum := 0
	for i := 0; i < 1000000; i++ {
		sum += i
	}
	fmt.Printf("Go Loop sum: %d\n", sum)
}
