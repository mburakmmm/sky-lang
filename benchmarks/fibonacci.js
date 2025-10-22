#!/usr/bin/env node

function fibonacci(n) {
    if (n <= 1) {
        return n;
    }
    return fibonacci(n - 1) + fibonacci(n - 2);
}

function main() {
    const start = Date.now();
    const result = fibonacci(35);
    const end = Date.now();
    const duration = end - start;
    
    console.log(`JavaScript Fibonacci(35) = ${result}`);
    console.log(`Duration: ${duration} ms`);
}

main();
