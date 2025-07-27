fn fib(n: i32) -> i32 {
    if n < 2 {
        return n;
    }
    fib(n - 1) + fib(n - 2)
}

fn main() {
    const N: i32 = 42;
    let result = fib(N);
    println!("Rust    | Resultado: {}", result);
}