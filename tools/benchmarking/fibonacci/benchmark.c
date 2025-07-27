#include <stdio.h>

int fib(int n) {
    if (n < 2) {
        return n;
    }
    return fib(n - 1) + fib(n - 2);
}

int main() {
    int N = 42;
    int result = fib(N);
    printf("C       | Resultado: %d\n", result);
    return 0;
}