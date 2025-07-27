import sys

# Aumenta o limite de recursão do Python, se necessário
sys.setrecursionlimit(2000)

def fib(n):
    if n < 2:
        return n
    return fib(n - 1) + fib(n - 2)

def main():
    N = 42
    result = fib(N)
    print(f"Python  | Resultado: {result}")

if __name__ == "__main__":
    main()