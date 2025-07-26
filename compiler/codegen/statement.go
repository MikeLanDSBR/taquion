package codegen

import (
	"fmt"
	"taquion/compiler/ast"

	"tinygo.org/x/go-llvm"
)

func (c *CodeGenerator) genStatement(stmt ast.Statement) {
	defer c.trace(fmt.Sprintf("genStatement (%T)", stmt))()

	switch node := stmt.(type) {
	case *ast.PackageStatement:
		c.logTrace(fmt.Sprintf("Ignorando declaração de pacote: package %s", node.Name.Value))

	case *ast.LetStatement:
		c.logTrace(fmt.Sprintf("Gerando declaração 'let' para a variável '%s'", node.Name.Value))
		val := c.genExpression(node.Value)
		typ := val.Type()
		ptr := c.builder.CreateAlloca(typ, node.Name.Value)
		c.builder.CreateStore(val, ptr)
		c.setSymbol(node.Name.Value, SymbolEntry{Ptr: ptr, Typ: typ})

	case *ast.ConstStatement:
		c.logTrace(fmt.Sprintf("Gerando declaração 'const' para a constante '%s'", node.Name.Value))
		val := c.genExpression(node.Value)
		typ := val.Type()
		ptr := c.builder.CreateAlloca(typ, node.Name.Value)
		c.builder.CreateStore(val, ptr)
		c.setSymbol(node.Name.Value, SymbolEntry{Ptr: ptr, Typ: typ})

	case *ast.FunctionDeclaration:
		var retType llvm.Type
		if node.Name.Value == "main" {
			retType = c.context.Int32Type()
		} else {
			retType = c.context.Int32Type() // Padrão
		}
		c.currentFunctionReturnType = retType

		// Constrói os tipos dos parâmetros
		paramTypes := make([]llvm.Type, len(node.Parameters))
		for i := range node.Parameters {
			// Por enquanto, todos os parâmetros são assumidos como i32
			paramTypes[i] = c.context.Int32Type()
		}

		funcType := llvm.FunctionType(retType, paramTypes, false)
		function := llvm.AddFunction(c.module, node.Name.Value, funcType)
		c.setSymbol(node.Name.Value, SymbolEntry{Ptr: function, Typ: funcType})

		if node.Body != nil {
			entryBlock := c.context.AddBasicBlock(function, "entry")
			c.builder.SetInsertPointAtEnd(entryBlock)
			c.pushScope() // Entra no novo escopo da função

			// === CORREÇÃO APLICADA AQUI ===
			// Adicionar os parâmetros da função à tabela de símbolos
			for i, param := range node.Parameters {
				paramValue := function.Param(i)
				paramValue.SetName(param.Value)

				// Alocar espaço na stack para o parâmetro e armazenar o valor
				alloca := c.builder.CreateAlloca(paramValue.Type(), param.Value)
				c.builder.CreateStore(paramValue, alloca)
				c.setSymbol(param.Value, SymbolEntry{Ptr: alloca, Typ: paramValue.Type()})
			}
			// ==============================

			c.genStatement(node.Body)
			c.popScope() // Sai do escopo
			if !isBlockTerminated(c.builder.GetInsertBlock()) {
				c.builder.CreateRet(llvm.ConstInt(retType, 0, false))
			}
		}

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
			if isBlockTerminated(c.builder.GetInsertBlock()) {
				c.logTrace("Bloco já terminado, pulando o resto das declarações.")
				break
			}
			c.genStatement(s)
		}

	case *ast.WhileStatement:
		function := c.builder.GetInsertBlock().Parent()

		condBlock := c.context.AddBasicBlock(function, "loop_cond")
		loopBlock := c.context.AddBasicBlock(function, "loop_body")
		endBlock := c.context.AddBasicBlock(function, "loop_end")

		// Salva o contexto do loop anterior e define o novo
		prevLoopCond := c.loopCondBlock
		prevLoopEnd := c.loopEndBlock
		c.loopCondBlock = condBlock
		c.loopEndBlock = endBlock

		c.builder.CreateBr(condBlock) // Pula para o bloco de condição

		// Bloco de condição
		c.builder.SetInsertPointAtEnd(condBlock)
		cond := c.genExpression(node.Condition)
		c.builder.CreateCondBr(cond, loopBlock, endBlock)

		// Bloco do corpo do loop
		c.builder.SetInsertPointAtEnd(loopBlock)
		c.genStatement(node.Body)
		if !isBlockTerminated(c.builder.GetInsertBlock()) {
			c.builder.CreateBr(condBlock) // Volta para a condição
		}

		// Bloco de saída
		c.builder.SetInsertPointAtEnd(endBlock)

		// Restaura o contexto do loop anterior
		c.loopCondBlock = prevLoopCond
		c.loopEndBlock = prevLoopEnd

	case *ast.BreakStatement:
		if c.loopEndBlock.IsNil() {
			panic("'break' fora de um loop")
		}
		c.builder.CreateBr(c.loopEndBlock)

	case *ast.ContinueStatement:
		if c.loopCondBlock.IsNil() {
			panic("'continue' fora de um loop")
		}
		c.builder.CreateBr(c.loopCondBlock)

	default:
		panic(fmt.Sprintf("Declaração não suportada: %T\n", node))
	}
}
