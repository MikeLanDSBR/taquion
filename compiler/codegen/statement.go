// Arquivo: codegen/statement.go
// Função: Geração de código para todos os tipos de declarações (statements).
package codegen

import (
	"fmt"
	"taquion/compiler/ast"

	"tinygo.org/x/go-llvm"
)

func (c *CodeGenerator) genStatement(stmt ast.Statement) {
	defer c.traceOut(fmt.Sprintf("genStatement (%T)", stmt))
	c.traceIn(fmt.Sprintf("genStatement (%T)", stmt))

	switch node := stmt.(type) {
	// --- NOVO CASE PARA DECLARAÇÃO DE FUNÇÃO ---
	case *ast.FunctionDeclaration:
		// Por enquanto, só lidamos com parâmetros int
		paramTypes := []llvm.Type{}
		for range node.Parameters {
			paramTypes = append(paramTypes, c.context.Int32Type())
		}

		// Por enquanto, só lidamos com retorno int
		retType := c.context.Int32Type()

		funcType := llvm.FunctionType(retType, paramTypes, false)
		function := llvm.AddFunction(c.module, node.Name.Value, funcType)

		// Salva a função na tabela de símbolos para que possa ser chamada depois.
		c.setSymbol(node.Name.Value, SymbolEntry{Ptr: function, Typ: funcType})

		// Somente gera o corpo se a função tiver um.
		if node.Body != nil {
			entryBlock := c.context.AddBasicBlock(function, "entry")
			// Salva o bloco atual para restaurar depois
			prevBlock := c.builder.GetInsertBlock()
			c.builder.SetInsertPointAtEnd(entryBlock)

			// Cria um novo escopo para o corpo da função
			c.pushScope()
			defer c.popScope()

			// Aloca espaço para os parâmetros e os copia para a tabela de símbolos
			for i, param := range node.Parameters {
				llvmParam := function.Param(i)
				llvmParam.SetName(param.Name.Value)

				ptr := c.builder.CreateAlloca(c.context.Int32Type(), param.Name.Value)
				c.builder.CreateStore(llvmParam, ptr)
				c.setSymbol(param.Name.Value, SymbolEntry{Ptr: ptr, Typ: c.context.Int32Type()})
			}

			// Gera o código para o corpo da função
			c.genStatement(node.Body)

			// Garante um terminador se não houver um 'return' explícito
			if !isBlockTerminated(c.builder.GetInsertBlock()) {
				// Se for a função main, retorna 0 por padrão. Para outras, pode ser um erro ou undefined behavior.
				if node.Name.Value == "main" {
					c.builder.CreateRet(llvm.ConstInt(c.context.Int32Type(), 0, false))
				}
			}
			// Restaura o builder para onde estava antes desta função.
			if !prevBlock.IsNil() {
				c.builder.SetInsertPointAtEnd(prevBlock)
			}
		}

	case *ast.LetStatement:
		c.logTrace(fmt.Sprintf("Gerando declaração 'let' para a variável '%s'", node.Name.Value))
		val := c.genExpression(node.Value)
		typ := val.Type()
		ptr := c.builder.CreateAlloca(typ, node.Name.Value)
		c.builder.CreateStore(val, ptr)
		c.setSymbol(node.Name.Value, SymbolEntry{Ptr: ptr, Typ: typ})

	case *ast.AssignmentStatement:
		c.logTrace(fmt.Sprintf("Gerando reatribuição para a variável '%s'", node.Name.Value))
		val := c.genExpression(node.Value)
		entry, ok := c.getSymbol(node.Name.Value)
		if !ok {
			panic(fmt.Sprintf("atribuição a variável não declarada: %s", node.Name.Value))
		}
		c.builder.CreateStore(val, entry.Ptr)

	case *ast.ReturnStatement:
		c.logTrace("Gerando declaração 'return'")
		val := c.genExpression(node.ReturnValue)
		c.builder.CreateRet(val)

	case *ast.ExpressionStatement:
		c.logTrace("Gerando declaração de expressão")
		c.genExpression(node.Expression)

	case *ast.BlockStatement:
		c.pushScope()
		defer c.popScope()
		c.logTrace("Gerando declaração de bloco (novo escopo)")
		for _, s := range node.Statements {
			c.genStatement(s)
		}

	default:
		panic(fmt.Sprintf("Declaração não suportada: %T\n", node))
	}
}
