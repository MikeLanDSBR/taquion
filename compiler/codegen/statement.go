// Arquivo: codegen/statement.go
package codegen

import (
	"fmt"
	"taquion/compiler/ast"
)

// genStatement gera código LLVM IR para uma declaração AST.
func (c *CodeGenerator) genStatement(stmt ast.Statement) {
	switch node := stmt.(type) {
	case *ast.LetStatement:
		// Gera o código para a expressão (ex: 10 ou "Hello, World!").
		val := c.genExpression(node.Value)

		// --- MUDANÇA PRINCIPAL AQUI ---
		// Em vez de sempre alocar para um inteiro (i32), agora alocamos
		// espaço baseado no tipo do valor que recebemos.
		// Se val for um i32, aloca espaço para i32.
		// Se val for um i8* (ponteiro de string), aloca espaço para um i8*.
		ptr := c.builder.CreateAlloca(val.Type(), node.Name.Value)

		c.builder.CreateStore(val, ptr)
		c.symbolTable[node.Name.Value] = ptr

	case *ast.AssignmentStatement:
		// Gera o código para o novo valor.
		val := c.genExpression(node.Value)
		// Procura a variável na tabela de símbolos.
		ptr, ok := c.symbolTable[node.Name.Value]
		if !ok {
			panic(fmt.Sprintf("atribuição a variável não declarada: %s", node.Name.Value))
		}
		// Armazena (store) o novo valor no ponteiro existente.
		// NOTA: Esta parte também precisará de uma atualização de tipo mais tarde,
		// mas para 'let' a mudança principal já foi feita acima.
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
