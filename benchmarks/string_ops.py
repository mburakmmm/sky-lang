#!/usr/bin/env python3
import time

def string_operations():
    text = "Hello World from SKY Programming Language"
    iterations = 100000
    
    start = time.time()
    
    for i in range(iterations):
        upper = text.upper()
        lower = text.lower()
        replaced = text.replace("SKY", "GO")
        joined = "-".join(["a", "b", "c"])
        split_result = text.split(" ")
    
    end = time.time()
    duration = (end - start) * 1000  # Convert to milliseconds
    
    print(f"Python String operations ({iterations} iterations)")
    print(f"Duration: {duration:.2f} ms")

def main():
    string_operations()

if __name__ == "__main__":
    main()
