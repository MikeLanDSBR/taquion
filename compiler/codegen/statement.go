// Arquivo: codegen/statement.go
// Função: Geração de código para todos os tipos de declarações (statements).
package codegen

import (
	"fmt"
	"taquion/compiler/ast"

	"tinygo.org/x/go-llvm"
)

func (c *CodeGenerator) genStatement(stmt ast.Statement) {
	// MODIFICADO: Usando a nova função de trace para simplificar.
	defer c.trace(fmt.Sprintf("genStatement (%T)", stmt))()

	switch node := stmt.(type) {
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
		// O tipo aqui é o tipo da função (assinatura), não o tipo de retorno.
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

				// O tipo do valor do parâmetro é int32
				paramType := c.context.Int32Type()
				ptr := c.builder.CreateAlloca(paramType, param.Name.Value)
				c.builder.CreateStore(llvmParam, ptr)
				c.setSymbol(param.Name.Value, SymbolEntry{Ptr: ptr, Typ: paramType})
			}

			// Gera o código para o corpo da função
			c.genStatement(node.Body)

			// Garante um terminador se não houver um 'return' explícito
			if !isBlockTerminated(c.builder.GetInsertBlock()) {
				// Se for a função main, retorna 0 por padrão.
				// Para outras funções, isso significa que não há retorno explícito.
				// Em uma linguagem mais estrita, isso seria um erro se a função devesse retornar um valor.
				if node.Name.Value == "main" {
					c.builder.CreateRet(llvm.ConstInt(c.context.Int32Type(), 0, false))
				} else {
					// Para outras funções void (que não retornam valor), um `ret void` seria inserido.
					// Como estamos assumindo retornos int32, a falta de um return é um comportamento indefinido.
					// LLVM requer um terminador, então adicionamos um `unreachable` para indicar isso.
					// c.builder.CreateUnreachable() // Ou um ret padrão se a linguagem permitir.
				}
			}

			// Restaura o builder para onde estava antes desta função.
			// Verifica se prevBlock é válido e não é o mesmo que o bloco atual
			if !prevBlock.IsNil() && prevBlock.C != c.builder.GetInsertBlock().C {
				c.builder.SetInsertPointAtEnd(prevBlock)
			}
		}

	case *ast.LetStatement:
		c.logTrace(fmt.Sprintf("Gerando declaração 'let' para a variável '%s'", node.Name.Value))
		val := c.genExpression(node.Value)
		typ := val.Type()
		ptr := c.builder.CreateAlloca(typ, node.Name.Value)
		c.builder.CreateStore(val, ptr)
		// O tipo na tabela de símbolos é o tipo do valor que a variável armazena.
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
		c.logTrace("Gerando declaração de bloco")
		for _, s := range node.Statements {
			c.genStatement(s)
		}

	default:
		panic(fmt.Sprintf("Declaração não suportada: %T\n", node))
	}
}
