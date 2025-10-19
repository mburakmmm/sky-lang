#!/bin/bash

echo "======================================"
echo "FIBONACCI BENCHMARK COMPARISON"
echo "Computing fib(35) - Recursive"
echo "======================================"
echo ""

# Compile programs
echo "Compiling..."
gcc -O2 -o fibonacci_c fibonacci.c
go build -o fibonacci_go fibonacci.go
echo ""

# Run benchmarks
echo "1. C (gcc -O2):"
time ./fibonacci_c
echo ""

echo "2. Go (compiled):"
time ./fibonacci_go
echo ""

echo "3. Python 3:"
time python3 fibonacci.py
echo ""

echo "4. SKY (interpreter):"
time ../bin/sky run fibonacci.sky
echo ""

echo "======================================"
echo "BENCHMARK COMPLETE"
echo "======================================"

# Cleanup
rm -f fibonacci_c fibonacci_go

