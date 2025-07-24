// Arquivo: codegen/expression.go
// Função: Geração de código para todos os tipos de expressões (expressions).
package codegen

import (
	"fmt"
	"strconv"
	"taquion/compiler/ast"
	"taquion/compiler/token"

	"tinygo.org/x/go-llvm"
)

func (c *CodeGenerator) genExpression(expr ast.Expression) llvm.Value {
	defer c.traceOut(fmt.Sprintf("genExpression (%T)", expr))
	c.traceIn(fmt.Sprintf("genExpression (%T)", expr))

	switch node := expr.(type) {
	// --- NOVO CASE PARA CHAMADA DE FUNÇÃO ---
	case *ast.CallExpression:
		c.logTrace(fmt.Sprintf("Gerando chamada para a função '%s'", node.Function.String()))

		// Procura a função na tabela de símbolos.
		calleeEntry, ok := c.getSymbol(node.Function.String())
		if !ok {
			panic(fmt.Sprintf("função não definida: %s", node.Function.String()))
		}
		callee := calleeEntry.Ptr
		calleeType := calleeEntry.Typ

		// Gera o código para cada argumento da chamada.
		args := []llvm.Value{}
		for _, arg := range node.Arguments {
			args = append(args, c.genExpression(arg))
		}

		// Cria a instrução 'call'.
		return c.builder.CreateCall(calleeType, callee, args, "calltmp")

	case *ast.IntegerLiteral:
		c.logTrace(fmt.Sprintf("Gerando literal inteiro: %s", node.TokenLiteral()))
		val, _ := strconv.ParseInt(node.TokenLiteral(), 10, 64)
		return llvm.ConstInt(c.context.Int32Type(), uint64(val), false)

	case *ast.StringLiteral:
		c.logTrace(fmt.Sprintf("Gerando literal de string: %q", node.Value))
		return c.builder.CreateGlobalStringPtr(node.Value, "str_literal")

	case *ast.Identifier:
		c.logTrace(fmt.Sprintf("Carregando valor da variável '%s'", node.Value))
		entry, ok := c.getSymbol(node.Value)
		if !ok {
			panic(fmt.Sprintf("variável não definida: %s", node.Value))
		}
		return c.builder.CreateLoad(entry.Typ, entry.Ptr, node.Value)

	case *ast.InfixExpression:
		c.logTrace(fmt.Sprintf("Gerando expressão infixa com operador '%s'", node.Operator))
		left := c.genExpression(node.Left)
		right := c.genExpression(node.Right)
		switch node.Operator {
		case token.PLUS:
			return c.builder.CreateAdd(left, right, "addtmp")
		case token.MINUS:
			return c.builder.CreateSub(left, right, "subtmp")
		case token.ASTERISK:
			return c.builder.CreateMul(left, right, "multmp")
		case token.SLASH:
			return c.builder.CreateSDiv(left, right, "divtmp")
		case token.GT:
			cmp := c.builder.CreateICmp(llvm.IntSGT, left, right, "gttmp")
			return c.builder.CreateZExt(cmp, c.context.Int32Type(), "booltmp")
		case token.LT:
			cmp := c.builder.CreateICmp(llvm.IntSLT, left, right, "lttmp")
			return c.builder.CreateZExt(cmp, c.context.Int32Type(), "booltmp")
		case token.EQ:
			cmp := c.builder.CreateICmp(llvm.IntEQ, left, right, "eqtmp")
			return c.builder.CreateZExt(cmp, c.context.Int32Type(), "booltmp")
		case token.NOT_EQ:
			cmp := c.builder.CreateICmp(llvm.IntNE, left, right, "neqtmp")
			return c.builder.CreateZExt(cmp, c.context.Int32Type(), "booltmp")
		default:
			panic(fmt.Sprintf("operador infix não suportado: %s", node.Operator))
		}

	case *ast.IfExpression:
		return c.genIfExpression(node)

	default:
		panic(fmt.Sprintf("Expressão não suportada: %T\n", node))
	}
}

func (c *CodeGenerator) genIfExpression(ie *ast.IfExpression) llvm.Value {
	defer c.traceOut("genIfExpression")
	c.traceIn("genIfExpression")

	cond := c.genExpression(ie.Condition)
	condVal := c.builder.CreateICmp(llvm.IntNE, cond, llvm.ConstInt(cond.Type(), 0, false), "ifcond")

	function := c.builder.GetInsertBlock().Parent()

	thenBlock := c.context.AddBasicBlock(function, "then")
	elseBlock := c.context.AddBasicBlock(function, "else")
	mergeBlock := c.context.AddBasicBlock(function, "merge")

	c.builder.CreateCondBr(condVal, thenBlock, elseBlock)

	// --- Bloco 'then' ---
	c.builder.SetInsertPointAtEnd(thenBlock)
	c.genStatement(ie.Consequence)
	if !isBlockTerminated(c.builder.GetInsertBlock()) {
		c.builder.CreateBr(mergeBlock)
	}

	// --- Bloco 'else' ---
	c.builder.SetInsertPointAtEnd(elseBlock)
	if ie.Alternative != nil {
		c.genStatement(ie.Alternative)
	}
	if !isBlockTerminated(c.builder.GetInsertBlock()) {
		c.builder.CreateBr(mergeBlock)
	}

	// --- Bloco 'merge' ---
	c.builder.SetInsertPointAtEnd(mergeBlock)
	return llvm.Value{}
}
