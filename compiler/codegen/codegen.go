// O pacote codegen é responsável por traduzir a Árvore Sintática Abstrata (AST)
// para a Representação Intermediária (IR) do LLVM.
package codegen

import (
	"fmt"
	"strconv"
	"taquion/compiler/ast"

	"tinygo.org/x/go-llvm" // A biblioteca correta e mantida
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
	// A API do go-llvm requer um "Contexto" para gerenciar a memória e os tipos.
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
	for _, stmt := range program.Statements {
		if funcDecl, ok := stmt.(*ast.FunctionDeclaration); ok && funcDecl.Name.Value == "main" {
			// Tipos são obtidos através do contexto, ex: c.context.Int32Type().
			funcType := llvm.FunctionType(c.context.Int32Type(), []llvm.Type{}, false)
			mainFunc := llvm.AddFunction(c.module, "main", funcType)
			entryBlock := llvm.AddBasicBlock(mainFunc, "entry")
			c.builder.SetInsertPointAtEnd(entryBlock)

			// Gera o código para o corpo da função.
			for _, bodyStmt := range funcDecl.Body.Statements {
				c.genStatement(bodyStmt)
			}
			break
		}
	}

	// CORRIGIDO: A forma correta e final de verificar se um bloco não tem terminador,
	// usando os métodos que você encontrou.
	// 1. Pega o bloco de inserção atual do builder.
	currentBlock := c.builder.GetInsertBlock()
	// 2. Pega a última instrução do bloco.
	lastInst := currentBlock.LastInstruction()
	// 3. Verifica se a última instrução não existe (.IsNil()) OU se ela não é um terminador.
	//    Um terminador é uma instrução como Ret, Br (Branch), Switch, etc.
	if lastInst.IsNil() || lastInst.IsAReturnInst().IsNil() {
		// Adicionamos um `return 0` padrão se nenhum `return` explícito for encontrado.
		c.builder.CreateRet(llvm.ConstInt(c.context.Int32Type(), 0, false))
	}

	return c.module
}

// genStatement gera código para uma única declaração.
func (c *CodeGenerator) genStatement(stmt ast.Statement) {
	switch node := stmt.(type) {
	case *ast.LetStatement:
		val := c.genExpression(node.Value)
		ptr := c.builder.CreateAlloca(c.context.Int32Type(), node.Name.Value)
		c.builder.CreateStore(val, ptr)
		c.symbolTable[node.Name.Value] = ptr
	case *ast.ReturnStatement:
		val := c.genExpression(node.ReturnValue)
		c.builder.CreateRet(val)
	}
}

// genExpression gera código para uma única expressão.
func (c *CodeGenerator) genExpression(expr ast.Expression) llvm.Value {
	switch node := expr.(type) {
	case *ast.IntegerLiteral:
		val, _ := strconv.ParseInt(node.TokenLiteral(), 10, 64)
		return llvm.ConstInt(c.context.Int32Type(), uint64(val), false)
	case *ast.Identifier:
		if ptr, ok := c.symbolTable[node.Value]; ok {
			return c.builder.CreateLoad(c.context.Int32Type(), ptr, node.Value+"_val")
		}
		panic(fmt.Sprintf("variável não definida: %s", node.Value))
	default:
		return llvm.Value{}
	}
}
