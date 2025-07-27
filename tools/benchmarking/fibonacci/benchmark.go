package main

import "fmt"

func fib(n int) int {
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func main() {
	const N = 42
	result := fib(N)
	fmt.Printf("Go      | Resultado: %d\n", result)
}
