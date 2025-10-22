#!/usr/bin/env python3
import time

def is_prime(n):
    if n < 2:
        return False
    
    i = 2
    while i * i <= n:
        if n % i == 0:
            return False
        i += 1
    
    return True

def count_primes(limit):
    count = 0
    for i in range(2, limit):
        if is_prime(i):
            count += 1
    return count

def main():
    start = time.time()
    result = count_primes(10000)
    end = time.time()
    duration = (end - start) * 1000  # Convert to milliseconds
    
    print(f"Python Primes up to 10000: {result}")
    print(f"Duration: {duration:.2f} ms")

if __name__ == "__main__":
    main()
