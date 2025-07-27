function fib(n) {
    if (n < 2) {
        return n;
    }
    return fib(n - 1) + fib(n - 2);
}

function main() {
    const N = 42;
    const result = fib(N);
    console.log(`JS      | Resultado: ${result}`);
}

main();