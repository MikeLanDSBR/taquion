// O pacote codegen é responsável por traduzir a Árvore Sintática Abstrata (AST)
// para a Representação Intermediária (IR) do LLVM.
package codegen

import (
	"taquion/compiler/ast"

	"tinygo.org/x/go-llvm"
)

// CodeGenerator mantém o estado durante a geração de código.
type CodeGenerator struct {
	module      llvm.Module
	builder     llvm.Builder
	context     llvm.Context
	symbolTable map[string]llvm.Value // Tabela de símbolos para rastrear variáveis.
}

// NewCodeGenerator cria uma nova instância do gerador de código.
func NewCodeGenerator() *CodeGenerator {
	ctx := llvm.NewContext()
	return &CodeGenerator{
		context:     ctx,
		module:      ctx.NewModule("main_module"),
		builder:     ctx.NewBuilder(),
		symbolTable: make(map[string]llvm.Value),
	}
}

// Generate é o ponto de entrada que traduz a AST para um módulo LLVM.
func (c *CodeGenerator) Generate(program *ast.Program) llvm.Module {
	// Itera sobre as declarações de alto nível no programa.
	for _, stmt := range program.Statements {
		// Procura pela declaração da função 'main'.
		if funcDecl, ok := stmt.(*ast.FunctionDeclaration); ok && funcDecl.Name.Value == "main" {
			// Define o tipo da função main: int main()
			funcType := llvm.FunctionType(c.context.Int32Type(), []llvm.Type{}, false)
			mainFunc := llvm.AddFunction(c.module, "main", funcType)
			entryBlock := llvm.AddBasicBlock(mainFunc, "entry")
			c.builder.SetInsertPointAtEnd(entryBlock)

			// Gera o código para todas as declarações dentro do corpo da função main.
			for _, bodyStmt := range funcDecl.Body.Statements {
				c.genStatement(bodyStmt)
			}

			// CORREÇÃO: Garante que a função 'main' sempre tenha um terminador (return).
			// Usa a API compatível para verificar a última instrução.
			currentBlock := c.builder.GetInsertBlock()
			lastInst := currentBlock.LastInstruction()
			if lastInst.IsNil() || lastInst.IsAReturnInst().IsNil() {
				// Adiciona um `return 0` padrão se nenhum `return` explícito for encontrado.
				c.builder.CreateRet(llvm.ConstInt(c.context.Int32Type(), 0, false))
			}
			break // Sai do loop após encontrar e processar a função main.
		}
	}

	return c.module // Retorna o módulo LLVM completo.
}
