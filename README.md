# Taquion - Uma Linguagem de Programação Compilada

![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![LLVM Version](https://img.shields.io/badge/llvm-20-blueviolet)
![Language](https://img.shields.io/badge/language-Go-blue)

Taquion é uma linguagem de programação imperativa, compilada e estaticamente tipada, projetada para ser simples, moderna e eficiente. O compilador, `TaquionC`, é escrito em Go e utiliza o LLVM como backend para gerar código de máquina otimizado.

## ✨ Funcionalidades

A linguagem Taquion atualmente suporta um conjunto robusto de funcionalidades essenciais:

* **Variáveis e Constantes:** Declaração com `let` e `const`.
* **Tipos Primitivos:** Inteiros, Booleanos e Strings.
* **Arrays:** Declaração de arrays de tamanho fixo, com acesso e atribuição por índice.
* **Operadores Aritméticos:** `+`, `-`, `*`, `/`, `%` com suporte a precedência de operadores.
* **Operadores Lógicos e de Comparação:** `!`, `==`, `!=`, `<`, `>`.
* **Estruturas de Controle:** Condicionais `if/else` e loops `while`.
* **Controle de Fluxo em Loops:** Suporte a `break` e `continue`.
* **Funções:** Declaração, chamada e suporte a recursão.
* **Concatenação de Strings:** Usando o operador `+`.
* **Escopo:** Regras de escopo léxico, incluindo sombreamento de variáveis (*scope shadowing*).

## 🚀 Instalação e Compilação

Para compilar e executar programas em Taquion, você precisará do `go`, `clang` e das bibliotecas do `LLVM` (versão 20) instalados no seu sistema.

#### 1. Compile o Compilador `TaquionC`

Primeiro, compile o próprio `taquionc` a partir do código-fonte. Dentro do diretório `compiler/`, execute:

```sh
make build
# ou o comando Go diretamente:
# go build -o ../build/taquionc.exe ./cmd/taquionc
```

Isso irá gerar o executável `taquionc.exe` no diretório `build/`.

#### 2. Compile um Programa Taquion

O processo de compilação de um programa Taquion ocorre em dois estágios:

**a. TaquionC: `.taq` -> `.ll` (LLVM IR)**

Use o `taquionc` para converter seu código-fonte Taquion (`.taq`) em um arquivo de Representação Intermediária do LLVM (`.ll`).

```sh
./build/taquionc seu_programa.taq -o saida.ll
```

**b. Clang: `.ll` -> Executável**

Use o `clang` para compilar o arquivo LLVM IR em um executável nativo.

```sh
clang saida.ll -o seu_programa.exe
```

#### 3. Execute seu Programa

Agora você pode executar seu programa compilado!

```sh
./seu_programa.exe
```

## 📝 Exemplos de Sintaxe

#### Olá, Mundo!

```go
package main

func main() {
    print("Olá, Mundo!");
    return 0;
}
```

#### Funções, Loops e Arrays

Este exemplo demonstra funcionalidades mais avançadas, como a verificação de números primos e a manipulação de um array.

```go
package main

// Verifica se um número é primo
func isPrime(n) {
    if (n <= 1) { return false; }
    if (n <= 3) { return true; }
    if ((n % 2) == 0) { return false; }

    let i = 5;
    while (i * i <= n) {
        if ((n % i) == 0) {
            return false;
        }
        i = i + 2;
    }
    return true;
}

func main() {
    // Manipulação de Array
    let items = [10, 20, 30];
    print("Item inicial:");
    print(items[1]); // Imprime 20

    items[1] = 25;
    print("Novo item:");
    print(items[1]); // Imprime 25

    // Loop com chamada de função
    print("Encontrando um primo...");
    let num = 13;
    if (isPrime(num)) {
        print("É primo!");
    } else {
        print("Não é primo.");
    }
    
    return 0;
}
```

## 🤝 Contribuindo

Contribuições são bem-vindas! Se você encontrar um bug ou tiver uma ideia para uma nova funcionalidade, sinta-se à vontade para abrir uma *issue* ou enviar um *pull request*.

## 📄 Licença

Este projeto é distribuído sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.