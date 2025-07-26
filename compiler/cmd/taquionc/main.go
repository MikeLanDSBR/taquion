package main

import (
	"fmt"
	"os"

	"taquion/compiler/ast" // Importado para usar o logger da AST
	"taquion/compiler/codegen"
	"taquion/compiler/lexer"
	"taquion/compiler/parser"
	"taquion/compiler/token" // Importado para usar o logger do Token
)

func main() {
	// --- CONFIGURAÇÃO DOS LOGGERS ---
	// Garante que todos os loggers sejam fechados na saída.
	defer token.CloseLogger()
	defer ast.CloseLogger()
	// lexer e parser ainda usam o padrão antigo, então mantemos seus defers.
	// codegen.CloseLogger() é chamado por generator.Close()

	if len(os.Args) < 2 {
		fmt.Println("Uso: taquionc <arquivo.taq>")
		return
	}
	filepath := os.Args[1]

	sourceCode, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Printf("Erro ao ler o arquivo %s: %s\n", filepath, err)
		return
	}

	l := lexer.New(string(sourceCode))
	defer l.LogFile.Close() // Mantido pois lexer usa logger por instância

	p := parser.New(l)
	defer p.LogFile.Close() // Mantido pois parser usa logger por instância

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		fmt.Println("Encontrados erros de parsing:")
		for _, msg := range p.Errors() {
			fmt.Println("\t" + msg)
		}
		return
	}

	// --- ACIONA O LOG DA AST ---
	// Chamar program.String() é o que gera o log em `log/ast.log`
	// e também imprime a AST na tela para depuração.
	fmt.Println("--- AST Gerada ---")
	fmt.Println(program.String())
	fmt.Println("--------------------")

	generator := codegen.NewCodeGenerator()
	defer generator.Close() // Isso chama codegen.CloseLogger() internamente
	module := generator.Generate(program)

	outputFilename := "../build/output.ll"
	err = os.WriteFile(outputFilename, []byte(module.String()), 0644)
	if err != nil {
		fmt.Printf("Erro ao escrever o arquivo .ll: %s\n", err)
		return
	}

	fmt.Printf("Arquivo LLVM IR gerado com sucesso: %s\n", outputFilename)
}
