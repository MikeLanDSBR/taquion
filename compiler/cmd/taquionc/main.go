package main

import (
	"fmt"
	"os"

	"taquion/compiler/codegen"
	"taquion/compiler/lexer"
	"taquion/compiler/parser"
)

func main() {
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
	defer l.LogFile.Close()

	p := parser.New(l)
	defer p.LogFile.Close()

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		fmt.Println("Encontrados erros de parsing:")
		for _, msg := range p.Errors() {
			fmt.Println("\t" + msg)
		}
		return
	}

	// CORRIGIDO: Cria uma instância do gerador e chama o método Generate.
	generator := codegen.NewCodeGenerator()
	module := generator.Generate(program)

	outputFilename := "output.ll"
	err = os.WriteFile(outputFilename, []byte(module.String()), 0644)
	if err != nil {
		fmt.Printf("Erro ao escrever o arquivo .ll: %s\n", err)
		return
	}

	fmt.Printf("Arquivo LLVM IR gerado com sucesso: %s\n", outputFilename)
}
