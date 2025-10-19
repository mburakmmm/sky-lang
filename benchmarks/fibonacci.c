#include <stdio.h>

int fib(int n) {
    if (n <= 1) {
        return n;
    }
    return fib(n - 1) + fib(n - 2);
}

int main() {
    int n = 35;
    printf("Computing fib(35)...\n");
    int result = fib(n);
    printf("Result: %d\n", result);
    return 0;
}

