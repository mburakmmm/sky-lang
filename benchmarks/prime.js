#!/usr/bin/env node

function isPrime(n) {
    if (n < 2) {
        return false;
    }
    
    for (let i = 2; i * i <= n; i++) {
        if (n % i === 0) {
            return false;
        }
    }
    
    return true;
}

function countPrimes(limit) {
    let count = 0;
    for (let i = 2; i < limit; i++) {
        if (isPrime(i)) {
            count++;
        }
    }
    return count;
}

function main() {
    const start = Date.now();
    const result = countPrimes(10000);
    const end = Date.now();
    const duration = end - start;
    
    console.log(`JavaScript Primes up to 10000: ${result}`);
    console.log(`Duration: ${duration} ms`);
}

main();
