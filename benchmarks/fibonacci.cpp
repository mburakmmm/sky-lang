#include <iostream>
#include <chrono>

int fibonacci(int n) {
    if (n <= 1) {
        return n;
    }
    return fibonacci(n - 1) + fibonacci(n - 2);
}

int main() {
    auto start = std::chrono::high_resolution_clock::now();
    int result = fibonacci(35);
    auto end = std::chrono::high_resolution_clock::now();
    
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end - start);
    
    std::cout << "C++ Fibonacci(35) = " << result << std::endl;
    std::cout << "Duration: " << duration.count() << " ms" << std::endl;
    
    return 0;
}
