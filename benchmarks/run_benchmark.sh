#!/bin/bash

# SKY Language Benchmark Suite
# Compares SKY with Go, Python, JavaScript, and C++

set -e

echo "🚀 SKY Language Benchmark Suite"
echo "================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to run benchmark and capture results
run_benchmark() {
    local name="$1"
    local command="$2"
    local file="$3"
    
    echo -e "${BLUE}Testing $name...${NC}"
    
    if [ ! -f "$file" ]; then
        echo -e "${RED}❌ File not found: $file${NC}"
        return 1
    fi
    
    # Run the command and capture output
    local output
    local exit_code
    
    if output=$($command 2>&1); then
        exit_code=0
    else
        exit_code=$?
    fi
    
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✅ $name completed successfully${NC}"
        echo "$output"
    else
        echo -e "${RED}❌ $name failed (exit code: $exit_code)${NC}"
        echo "$output"
    fi
    
    echo ""
}

# Function to measure memory usage
measure_memory() {
    local command="$1"
    local file="$2"
    
    if command -v valgrind >/dev/null 2>&1; then
        echo "Memory usage (valgrind):"
        valgrind --tool=massif --pages-as-heap=yes $command "$file" 2>&1 | grep -E "(peak|total)" || true
    elif command -v /usr/bin/time >/dev/null 2>&1; then
        echo "Memory usage (time):"
        /usr/bin/time -v $command "$file" 2>&1 | grep -E "(Maximum resident|Average resident)" || true
    fi
}

# Build SKY if needed
echo "Building SKY language..."
cd "$(dirname "$0")/.."
make build >/dev/null 2>&1 || echo "SKY already built"

# Build C++ if needed
if [ -f "benchmarks/fibonacci.cpp" ]; then
    echo "Building C++ benchmark..."
    g++ -O2 -o benchmarks/fibonacci_cpp benchmarks/fibonacci.cpp 2>/dev/null || echo "C++ build failed"
fi

cd benchmarks

echo "Starting benchmarks..."
echo ""

# Fibonacci benchmarks
echo -e "${YELLOW}=== FIBONACCI BENCHMARKS (n=35) ===${NC}"
echo ""

run_benchmark "SKY" "sky run fibonacci_35.sky" "fibonacci_35.sky"
run_benchmark "Go" "go run fibonacci.go" "fibonacci.go"
run_benchmark "Python" "python3 fibonacci.py" "fibonacci.py"
run_benchmark "JavaScript" "node fibonacci.js" "fibonacci.js"
run_benchmark "C++" "./fibonacci_cpp" "fibonacci_cpp"

echo -e "${YELLOW}=== PRIME NUMBER BENCHMARKS (up to 10,000) ===${NC}"
echo ""

# Prime benchmarks
run_benchmark "SKY" "sky run prime.sky" "prime.sky"
run_benchmark "Go" "go run prime.go" "prime.go"
run_benchmark "Python" "python3 prime.py" "prime.py"
run_benchmark "JavaScript" "node prime.js" "prime.js"
run_benchmark "C++" "./prime_cpp" "prime_cpp"

echo -e "${YELLOW}=== STRING OPERATIONS BENCHMARKS (100,000 iterations) ===${NC}"
echo ""

# String operation benchmarks
run_benchmark "SKY" "sky run string_ops.sky" "string_ops.sky"
run_benchmark "Go" "go run string_ops.go" "string_ops.go"
run_benchmark "Python" "python3 string_ops.py" "string_ops.py"
run_benchmark "JavaScript" "node string_ops.js" "string_ops.js"
run_benchmark "C++" "./string_ops_cpp" "string_ops_cpp"

echo -e "${YELLOW}=== MEMORY USAGE ANALYSIS ===${NC}"
echo ""

# Memory usage analysis
echo "SKY Memory Usage:"
measure_memory "sky run" "fibonacci.sky"

echo ""
echo "Go Memory Usage:"
measure_memory "go run" "fibonacci.go"

echo ""
echo "Python Memory Usage:"
measure_memory "python3" "fibonacci.py"

echo ""
echo -e "${GREEN}🎉 Benchmark suite completed!${NC}"
echo ""
echo "Note: SKY is currently interpreted (JIT), so it's expected to be slower than compiled languages."
echo "Future AOT compilation will significantly improve performance."