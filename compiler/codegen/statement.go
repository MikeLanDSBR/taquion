package codegen

import (
	"fmt"
	"taquion/compiler/ast"
)

// genStatement gera código LLVM IR para uma declaração AST.
func (c *CodeGenerator) genStatement(stmt ast.Statement) {
	switch node := stmt.(type) {
	case *ast.LetStatement:
		val := c.genExpression(node.Value)
		ptr := c.builder.CreateAlloca(c.context.Int32Type(), node.Name.Value)
		c.builder.CreateStore(val, ptr)
		c.symbolTable[node.Name.Value] = ptr

	// --- NOVO CASE PARA REATRIBUIÇÃO ---
	case *ast.AssignmentStatement:
		// Gera o código para o novo valor
		val := c.genExpression(node.Value)
		// Procura a variável na tabela de símbolos
		ptr, ok := c.symbolTable[node.Name.Value]
		if !ok {
			panic(fmt.Sprintf("atribuição a variável não declarada: %s", node.Name.Value))
		}
		// Armazena (store) o novo valor no ponteiro existente
		c.builder.CreateStore(val, ptr)

	case *ast.ReturnStatement:
		val := c.genExpression(node.ReturnValue)
		c.builder.CreateRet(val)

	case *ast.ExpressionStatement:
		c.genExpression(node.Expression)

	default:
		fmt.Printf("Declaração não suportada: %T\n", node)
	}
}
