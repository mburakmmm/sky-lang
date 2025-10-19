#!/usr/bin/env python3
# Fibonacci benchmark for Python

def fib(n):
    if n <= 1:
        return n
    return fib(n - 1) + fib(n - 2)

def main():
    n = 35
    print("Computing fib(35)...")
    result = fib(n)
    print(f"Result: {result}")

if __name__ == "__main__":
    main()

