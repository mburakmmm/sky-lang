#!/bin/bash
# SKY E2E Test Script

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0

echo "╔══════════════════════════════════════════════════════════╗"
echo "║           SKY E2E Test Suite                              ║"
echo "╚══════════════════════════════════════════════════════════╝"
echo

# Build first
echo "Building SKY..."
make build > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Build successful${NC}"
echo

# Test directories
TEST_DIRS="examples/smoke examples/mvp examples/sema"

for dir in $TEST_DIRS; do
    if [ ! -d "$dir" ]; then
        continue
    fi
    
    echo "Testing $dir/"
    echo "─────────────────────────────────────────────────────────"
    
    for file in $dir/*.sky; do
        if [ ! -f "$file" ]; then
            continue
        fi
        
        filename=$(basename "$file")
        expected_file="${file%.sky}.expected"
        
        # Run the file
        output=$(./bin/sky run "$file" 2>&1)
        exit_code=$?
        
        # Check for expected output file
        if [ -f "$expected_file" ]; then
            expected=$(cat "$expected_file")
            
            if [ "$output" = "$expected" ]; then
                echo -e "  ${GREEN}✓${NC} $filename"
                ((PASSED++))
            else
                echo -e "  ${RED}✗${NC} $filename (output mismatch)"
                echo "    Expected: $expected"
                echo "    Got: $output"
                ((FAILED++))
            fi
        else
            # No expected file, just check if it runs
            if [ $exit_code -eq 0 ] || [ "$filename" = "const_error.sky" ] || [ "$filename" = "type_error.sky" ]; then
                echo -e "  ${GREEN}✓${NC} $filename"
                ((PASSED++))
            else
                echo -e "  ${RED}✗${NC} $filename (exit code: $exit_code)"
                ((FAILED++))
            fi
        fi
    done
    echo
done

# Summary
echo "═══════════════════════════════════════════════════════════"
echo "Summary:"
echo -e "  ${GREEN}Passed: $PASSED${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "  ${RED}Failed: $FAILED${NC}"
else
    echo -e "  ${YELLOW}Failed: $FAILED${NC}"
fi
echo "═══════════════════════════════════════════════════════════"

if [ $FAILED -gt 0 ]; then
    exit 1
fi

echo
echo -e "${GREEN}All tests passed!${NC} 🎉"

