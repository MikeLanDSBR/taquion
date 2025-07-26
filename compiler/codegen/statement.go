package codegen

import (
	"fmt"
	"taquion/compiler/ast"

	"github.com/MikeLanDSBR/go-llvm"
)

func (c *CodeGenerator) genStatement(stmt ast.Statement) {
	defer c.trace(fmt.Sprintf("genStatement (%T)", stmt))()

	switch node := stmt.(type) {
	case *ast.PackageStatement:
		c.logTrace(fmt.Sprintf("Ignorando declaração de pacote: package %s", node.Name.Value))

	case *ast.LetStatement:
		c.logTrace(fmt.Sprintf("Gerando declaração 'let' para a variável '%s'", node.Name.Value))
		val := c.genExpression(node.Value)
		typ := c.GetValueTypeSafe(val)
		if typ.IsNil() {
			panic(fmt.Sprintf("tipo inválido para a variável 'let' %s", node.Name.Value))
		}
		ptr := c.builder.CreateAlloca(typ, node.Name.Value)
		c.builder.CreateStore(val, ptr)
		c.setSymbol(node.Name.Value, SymbolEntry{Ptr: ptr, Typ: typ, IsLiteral: false})

	case *ast.ConstStatement:
		c.logTrace(fmt.Sprintf("Gerando declaração 'const' para a constante '%s'", node.Name.Value))
		val := c.genExpression(node.Value)

		isAConstant := !val.IsAConstant().IsNil()

		if isAConstant {
			c.logTrace(fmt.Sprintf("DEBUG: Constante '%s' é um literal. Armazenando valor diretamente.", node.Name.Value))
			c.setSymbol(node.Name.Value, SymbolEntry{Value: val, Typ: c.GetValueTypeSafe(val), IsLiteral: true})
		} else {
			c.logTrace(fmt.Sprintf("DEBUG: Constante '%s' é um resultado de instrução, tratando como variável imutável.", node.Name.Value))
			typ := c.GetValueTypeSafe(val)
			if typ.IsNil() {
				panic(fmt.Sprintf("tipo inválido para a constante '%s'", node.Name.Value))
			}
			ptr := c.builder.CreateAlloca(typ, node.Name.Value)
			c.builder.CreateStore(val, ptr)
			c.setSymbol(node.Name.Value, SymbolEntry{Ptr: ptr, Typ: typ, IsLiteral: true})
		}

	case *ast.FunctionDeclaration:
		var retType llvm.Type
		if node.Name.Value == "main" {
			retType = c.context.Int32Type()
		} else {
			retType = c.context.Int32Type()
		}
		c.currentFunctionReturnType = retType

		paramTypes := make([]llvm.Type, len(node.Parameters))
		for i := range node.Parameters {
			paramTypes[i] = c.context.Int32Type()
		}

		funcType := llvm.FunctionType(retType, paramTypes, false)
		function := llvm.AddFunction(c.module, node.Name.Value, funcType)
		c.setSymbol(node.Name.Value, SymbolEntry{Value: function, Typ: funcType, IsLiteral: true})

		if node.Body != nil {
			entryBlock := c.context.AddBasicBlock(function, "entry")
			c.builder.SetInsertPointAtEnd(entryBlock)
			c.pushScope()

			for i, param := range node.Parameters {
				paramValue := function.Param(i)
				paramValue.SetName(param.Value)
				alloca := c.builder.CreateAlloca(paramValue.Type(), param.Value)
				c.builder.CreateStore(paramValue, alloca)
				c.setSymbol(param.Value, SymbolEntry{Ptr: alloca, Typ: paramValue.Type(), IsLiteral: false})
			}

			c.genStatement(node.Body)
			c.popScope()
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
		prevLoopCond := c.loopCondBlock
		prevLoopEnd := c.loopEndBlock
		c.loopCondBlock = condBlock
		c.loopEndBlock = endBlock
		c.builder.CreateBr(condBlock)
		c.builder.SetInsertPointAtEnd(condBlock)
		cond := c.genExpression(node.Condition)
		c.builder.CreateCondBr(cond, loopBlock, endBlock)
		c.builder.SetInsertPointAtEnd(loopBlock)
		c.genStatement(node.Body)
		if !isBlockTerminated(c.builder.GetInsertBlock()) {
			c.builder.CreateBr(condBlock)
		}
		c.builder.SetInsertPointAtEnd(endBlock)
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
