package main

import (
	"fmt"
	"os"

	"taquion/compiler/ast"
	"taquion/compiler/codegen"
	"taquion/compiler/lexer"
	"taquion/compiler/parser"
	"taquion/compiler/token"

	"github.com/taquion-lang/go-llvm"
)

func main() {
	// Garante que todos os loggers globais sejam fechados na saída.
	defer token.CloseLogger()
	defer ast.CloseLogger()

	// --- Processamento de Argumentos Simples (Compatível com o Tester) ---
	if len(os.Args) < 2 {
		fmt.Println("Uso: taquionc <arquivo.taq>")
		os.Exit(1)
	}
	inputFilePath := os.Args[1]

	// O nome do arquivo de saída é fixo para ser compatível com o Makefile e o script de teste.
	// O script de teste lida com a renomeação e movimentação dos arquivos.
	outputFilename := "../build/output.ll"
	if len(os.Args) > 3 && os.Args[2] == "-o" {
		outputFilename = os.Args[3]
	}

	// --- Pipeline de Compilação ---
	sourceCode, err := os.ReadFile(inputFilePath)
	if err != nil {
		fmt.Printf("Erro ao ler o arquivo %s: %s\n", inputFilePath, err)
		os.Exit(1)
	}

	l := lexer.New(string(sourceCode))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Encontrados erros de parsing:")
		for _, msg := range p.Errors() {
			fmt.Println("\t" + msg)
		}
		os.Exit(1)
	}

	fmt.Println("--- AST Gerada ---")
	fmt.Println(program.String())
	fmt.Println("--------------------")

	generator := codegen.NewCodeGenerator()
	defer generator.Close() // Isso chama codegen.CloseLogger() internamente
	module := generator.Generate(program)

	// Verifica se o módulo LLVM é válido
	if err := llvm.VerifyModule(module, llvm.PrintMessageAction); err != nil {
		fmt.Printf("Erro na verificação do módulo LLVM: %s\n", err)
		// Opcional: Descomente para ver o IR mesmo com erro
		// fmt.Println("--- LLVM IR (Inválido) ---")
		// fmt.Println(module.String())
		os.Exit(1)
	}

	err = os.WriteFile(outputFilename, []byte(module.String()), 0644)
	if err != nil {
		fmt.Printf("Erro ao escrever o arquivo .ll: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Arquivo LLVM IR gerado com sucesso: %s\n", outputFilename)
}
