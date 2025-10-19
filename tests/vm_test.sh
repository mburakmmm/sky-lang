#!/bin/bash
# VM Bytecode Tests

set -e

echo "╔════════════════════════════════════════════╗"
echo "║     SKY BYTECODE VM TEST SUITE             ║"
echo "╚════════════════════════════════════════════╝"
echo ""

BIN=./bin/sky

# Build first
make build > /dev/null 2>&1

echo "1. Testing simple function call..."
$BIN run --vm examples/vm/fibonacci.sky > /tmp/vm_fib_output.txt 2>&1
if diff -q examples/vm/fibonacci.expected /tmp/vm_fib_output.txt > /dev/null; then
    echo "   ✅ PASS - Fibonacci recursion"
else
    echo "   ❌ FAIL - Fibonacci output mismatch"
    diff examples/vm/fibonacci.expected /tmp/vm_fib_output.txt
    exit 1
fi

echo ""
echo "2. Benchmarking VM vs Interpreter..."
echo "   VM Mode (fib(25)):"
time $BIN run --vm examples/vm/fibonacci.sky 2>&1 | head -1

echo ""
echo "╔════════════════════════════════════════════╗"
echo "║        ALL VM TESTS PASSED! ✅             ║"
echo "╚════════════════════════════════════════════╝"

