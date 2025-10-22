#!/usr/bin/env python3
import time

def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n - 1) + fibonacci(n - 2)

def main():
    start = time.time()
    result = fibonacci(35)
    end = time.time()
    duration = (end - start) * 1000  # Convert to milliseconds
    
    print(f"Python Fibonacci(35) = {result}")
    print(f"Duration: {duration:.2f} ms")

if __name__ == "__main__":
    main()