#!/bin/bash
# VM Ultimate Test Runner

echo "╔════════════════════════════════════════════╗"
echo "║   SKY VM ULTIMATE TEST SUITE               ║"
echo "╚════════════════════════════════════════════╝"

BIN=./bin/sky

# Test 1: Simple recursion
echo "1. Fibonacci (recursion)..."
$BIN run --vm examples/vm/fibonacci.sky > /tmp/vm_fib.txt 2>&1
if diff -q examples/vm/fibonacci.expected /tmp/vm_fib.txt > /dev/null; then
    echo "   ✅ PASS"
else
    echo "   ❌ FAIL"
    exit 1
fi

# Test 2: Ultimate comprehensive
echo "2. Ultimate comprehensive test..."
$BIN run --vm examples/vm/ultimate_test.sky > /tmp/vm_ultimate.txt 2>&1
if diff -q examples/vm/ultimate_test.expected /tmp/vm_ultimate.txt > /dev/null; then
    echo "   ✅ PASS"
else
    echo "   ❌ FAIL"
    diff examples/vm/ultimate_test.expected /tmp/vm_ultimate.txt
    exit 1
fi

echo ""
echo "╔════════════════════════════════════════════╗"
echo "║     ✅ ALL VM TESTS PASSED! ✅             ║"
echo "╚════════════════════════════════════════════╝"
