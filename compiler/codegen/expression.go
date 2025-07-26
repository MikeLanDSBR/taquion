package codegen

import (
	"fmt"
	"math"
	"strconv"
	"taquion/compiler/ast"
	"taquion/compiler/token"

	"tinygo.org/x/go-llvm"
)

func (c *CodeGenerator) genExpression(expr ast.Expression) llvm.Value {
	defer c.trace(fmt.Sprintf("genExpression (%T)", expr))()

	switch node := expr.(type) {
	case *ast.BooleanLiteral:
		c.logTrace(fmt.Sprintf("Gerando literal booleano: %s", node.TokenLiteral()))
		if node.Value {
			return llvm.ConstInt(c.context.Int1Type(), 1, false)
		}
		return llvm.ConstInt(c.context.Int1Type(), 0, false)

	// --- FUNÇÃO MODIFICADA ---
	case *ast.CallExpression:
		c.logTrace(fmt.Sprintf("Gerando chamada para a função '%s'", node.Function.String()))

		calleeEntry, ok := c.getSymbol(node.Function.String())
		if !ok {
			panic(fmt.Sprintf("função não definida: %s", node.Function.String()))
		}
		callee := calleeEntry.Ptr
		// CORREÇÃO: Usar o tipo que já foi salvo na tabela de símbolos.
		// A chamada a callee.Type().ElementType() estava causando o travamento.
		funcType := calleeEntry.Typ

		args := []llvm.Value{}
		expectedParamTypes := funcType.ParamTypes()

		for i, arg := range node.Arguments {
			argVal := c.genExpression(arg)

			// Verifica se o tipo do argumento corresponde ao esperado pela função.
			if i < len(expectedParamTypes) {
				expectedType := expectedParamTypes[i]
				actualType := argVal.Type()

				// Se os tipos forem inteiros de tamanhos diferentes, promove o argumento.
				if actualType != expectedType && actualType.TypeKind() == llvm.IntegerTypeKind && expectedType.TypeKind() == llvm.IntegerTypeKind {
					c.logTrace(fmt.Sprintf("Convertendo argumento %d de %s para %s", i, actualType.String(), expectedType.String()))
					// Usa SExt para estender o valor, preservando o sinal (ex: -1_i8 para -1_i32).
					argVal = c.builder.CreateSExt(argVal, expectedType, fmt.Sprintf("argcast%d", i))
				}
			}
			args = append(args, argVal)
		}

		// A função CreateCall precisa do tipo da função.
		return c.builder.CreateCall(funcType, callee, args, "calltmp")

	case *ast.IntegerLiteral:
		c.logTrace(fmt.Sprintf("Gerando literal inteiro: %s", node.TokenLiteral()))
		val, err := strconv.ParseInt(node.TokenLiteral(), 10, 64)
		if err != nil {
			panic(fmt.Sprintf("não foi possível converter o literal inteiro: %s", node.TokenLiteral()))
		}

		var intType llvm.Type
		if val >= math.MinInt8 && val <= math.MaxInt8 {
			intType = c.context.Int8Type()
			c.logTrace(fmt.Sprintf("Valor %d cabe em um int8. Usando i8.", val))
		} else if val >= math.MinInt16 && val <= math.MaxInt16 {
			intType = c.context.Int16Type()
			c.logTrace(fmt.Sprintf("Valor %d cabe em um int16. Usando i16.", val))
		} else if val >= math.MinInt32 && val <= math.MaxInt32 {
			intType = c.context.Int32Type()
			c.logTrace(fmt.Sprintf("Valor %d cabe em um int32. Usando i32.", val))
		} else {
			intType = c.context.Int64Type()
			c.logTrace(fmt.Sprintf("Valor %d é grande. Usando i64.", val))
		}
		return llvm.ConstInt(intType, uint64(val), false)

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
			return c.builder.CreateICmp(llvm.IntSGT, left, right, "gttmp")
		case token.LT:
			return c.builder.CreateICmp(llvm.IntSLT, left, right, "lttmp")
		case token.EQ:
			return c.builder.CreateICmp(llvm.IntEQ, left, right, "eqtmp")
		case token.NOT_EQ:
			return c.builder.CreateICmp(llvm.IntNE, left, right, "neqtmp")
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
	defer c.trace("genIfExpression")()

	cond := c.genExpression(ie.Condition)
	function := c.builder.GetInsertBlock().Parent()

	thenBlock := c.context.AddBasicBlock(function, "then")
	elseBlock := c.context.AddBasicBlock(function, "else")
	mergeBlock := c.context.AddBasicBlock(function, "merge")

	c.builder.CreateCondBr(cond, thenBlock, elseBlock)

	c.builder.SetInsertPointAtEnd(thenBlock)
	c.genStatement(ie.Consequence)
	if !isBlockTerminated(c.builder.GetInsertBlock()) {
		c.builder.CreateBr(mergeBlock)
	}

	c.builder.SetInsertPointAtEnd(elseBlock)
	if ie.Alternative != nil {
		c.genStatement(ie.Alternative)
	}
	if !isBlockTerminated(c.builder.GetInsertBlock()) {
		c.builder.CreateBr(mergeBlock)
	}

	c.builder.SetInsertPointAtEnd(mergeBlock)
	return llvm.Value{}
}
