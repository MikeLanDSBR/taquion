// O pacote main é o ponto de entrada para o executável do compilador Taquion.
// Ele orquestra as fases de lexing, parsing e code generation.
package main

import (
	"fmt"
	"os"

	"taquion/compiler/codegen"
	"taquion/compiler/lexer"
	"taquion/compiler/parser"
)

func main() {
	// 1. Valida se um nome de arquivo foi passado como argumento.
	if len(os.Args) < 2 {
		fmt.Println("Uso: taquionc <arquivo.taq>")
		return
	}
	filepath := os.Args[1]

	// 2. Lê o conteúdo do arquivo de código-fonte.
	sourceCode, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Printf("Erro ao ler o arquivo %s: %s\n", filepath, err)
		return
	}

	// 3. Executa o front-end (Lexer e Parser).
	l := lexer.New(string(sourceCode))
	defer l.LogFile.Close() // Garante que o log do lexer seja fechado no final.

	p := parser.New(l)
	defer p.LogFile.Close() // Garante que o log do parser seja fechado no final.

	program := p.ParseProgram()

	// Verifica se o parser encontrou erros de sintaxe.
	if len(p.Errors()) != 0 {
		fmt.Println("Encontrados erros de parsing:")
		for _, msg := range p.Errors() {
			fmt.Println("\t" + msg)
		}
		return
	}

	// 4. Executa o back-end (Gerador de Código LLVM).
	module := codegen.Generate(program)

	// 5. Salva o código LLVM IR gerado em um arquivo `output.ll`.
	outputFilename := "output.ll"
	err = os.WriteFile(outputFilename, []byte(module.String()), 0644)
	if err != nil {
		fmt.Printf("Erro ao escrever o arquivo .ll: %s\n", err)
		return
	}

	fmt.Printf("Arquivo LLVM IR gerado com sucesso: %s\n", outputFilename)
}
