#!/bin/bash

echo "======================================"
echo "DETAILED BENCHMARK - 5 RUNS EACH"
echo "======================================"
echo ""

# Compile
gcc -O2 -o fibonacci_c fibonacci.c
go build -o fibonacci_go fibonacci.go

echo "Running C (5 times)..."
for i in {1..5}; do
  echo -n "Run $i: "
  time ./fibonacci_c 2>&1 | grep real
done
echo ""

echo "Running Go (5 times)..."
for i in {1..5}; do
  echo -n "Run $i: "
  time ./fibonacci_go 2>&1 | grep real
done
echo ""

echo "Running Python (5 times)..."
for i in {1..5}; do
  echo -n "Run $i: "
  time python3 fibonacci.py 2>&1 | grep real
done
echo ""

echo "Running SKY (5 times)..."
for i in {1..5}; do
  echo -n "Run $i: "
  time ../bin/sky run fibonacci.sky 2>&1 | grep real
done
echo ""

# Cleanup
rm -f fibonacci_c fibonacci_go

