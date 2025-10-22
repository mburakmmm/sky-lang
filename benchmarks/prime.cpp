#include <iostream>
#include <chrono>

bool isPrime(int n) {
    if (n < 2) {
        return false;
    }
    
    for (int i = 2; i * i <= n; i++) {
        if (n % i == 0) {
            return false;
        }
    }
    
    return true;
}

int countPrimes(int limit) {
    int count = 0;
    for (int i = 2; i < limit; i++) {
        if (isPrime(i)) {
            count++;
        }
    }
    return count;
}

int main() {
    auto start = std::chrono::high_resolution_clock::now();
    int result = countPrimes(10000);
    auto end = std::chrono::high_resolution_clock::now();
    
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end - start);
    
    std::cout << "C++ Primes up to 10000: " << result << std::endl;
    std::cout << "Duration: " << duration.count() << " ms" << std::endl;
    
    return 0;
}
