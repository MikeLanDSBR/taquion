package codegen

import (
	"fmt"
	"strconv"
	"taquion/compiler/ast"
	"taquion/compiler/token"

	"tinygo.org/x/go-llvm"
)

func (c *CodeGenerator) genExpression(expr ast.Expression) llvm.Value {
	defer c.trace(fmt.Sprintf("genExpression (%T)", expr))()

	switch node := expr.(type) {
	case *ast.IntegerLiteral:
		val, _ := strconv.ParseInt(node.TokenLiteral(), 10, 64)
		// Simplificado para i32 por enquanto para evitar problemas de promoção
		return llvm.ConstInt(c.context.Int32Type(), uint64(val), false)

	case *ast.StringLiteral:
		return c.builder.CreateGlobalStringPtr(node.Value, "str_literal")

	case *ast.BooleanLiteral:
		if node.Value {
			return llvm.ConstInt(c.context.Int1Type(), 1, false)
		}
		return llvm.ConstInt(c.context.Int1Type(), 0, false)

	case *ast.Identifier:
		entry, ok := c.getSymbol(node.Value)
		if !ok {
			panic(fmt.Sprintf("variável não definida: %s", node.Value))
		}
		return c.builder.CreateLoad(entry.Typ, entry.Ptr, node.Value)

	case *ast.InfixExpression:
		left := c.genExpression(node.Left)
		right := c.genExpression(node.Right)

		isLeftString := left.Type().TypeKind() == llvm.PointerTypeKind
		isRightString := right.Type().TypeKind() == llvm.PointerTypeKind

		if node.Operator == token.PLUS && isLeftString && isRightString {
			return c.genStringConcat(left, right)
		}

		switch node.Operator {
		case token.PLUS:
			return c.builder.CreateAdd(left, right, "addtmp")
		case token.MINUS:
			return c.builder.CreateSub(left, right, "subtmp")
		case token.ASTERISK:
			return c.builder.CreateMul(left, right, "multmp")
		case token.SLASH:
			return c.builder.CreateSDiv(left, right, "divtmp")
		case token.MODULO: // NOVO: Adicionado para o operador de módulo
			return c.builder.CreateSRem(left, right, "modtmp") // SRam para signed remainder (resto da divisão)
		case token.EQ:
			return c.builder.CreateICmp(llvm.IntEQ, left, right, "eqtmp")
		case token.NOT_EQ:
			return c.builder.CreateICmp(llvm.IntNE, left, right, "neqtmp")
		case token.LT:
			return c.builder.CreateICmp(llvm.IntSLT, left, right, "lttmp")
		case token.GT:
			return c.builder.CreateICmp(llvm.IntSGT, left, right, "gttmp")
		default:
			panic(fmt.Sprintf("operador infix não suportado: %s", node.Operator))
		}

	case *ast.AssignmentExpression:
		val := c.genExpression(node.Value)
		ident, ok := node.Left.(*ast.Identifier)
		if !ok {
			panic("o lado esquerdo de uma atribuição deve ser um identificador")
		}
		entry, ok := c.getSymbol(ident.Value)
		if !ok {
			panic(fmt.Sprintf("atribuição a variável não declarada: %s", ident.Value))
		}
		c.builder.CreateStore(val, entry.Ptr)
		return val

	case *ast.CallExpression:
		// Lógica para funções embutidas
		if node.Function.String() == "print" {
			return c.genPrintCall(node)
		}

		// Lógica para funções definidas pelo usuário
		symbol, ok := c.getSymbol(node.Function.String())
		if !ok {
			panic(fmt.Sprintf("função não definida: %s", node.Function.String()))
		}

		function := symbol.Ptr
		functionType := symbol.Typ

		args := []llvm.Value{}
		for _, argExpr := range node.Arguments {
			args = append(args, c.genExpression(argExpr))
		}

		return c.builder.CreateCall(functionType, function, args, "calltmp")

	case *ast.IfExpression:
		return c.genIfExpression(node)

	default:
		panic(fmt.Sprintf("Expressão não suportada: %T\n", node))
	}
}

func (c *CodeGenerator) genStringConcat(left, right llvm.Value) llvm.Value {
	c.logTrace("Gerando concatenação de strings")
	len1 := c.builder.CreateCall(c.strlenFunc.Type().ElementType(), c.strlenFunc, []llvm.Value{left}, "len1")
	len2 := c.builder.CreateCall(c.strlenFunc.Type().ElementType(), c.strlenFunc, []llvm.Value{right}, "len2")
	totalLen := c.builder.CreateAdd(len1, len2, "totalLen")
	bufferSize := c.builder.CreateAdd(totalLen, llvm.ConstInt(c.context.Int64Type(), 1, false), "bufferSize")
	newBuffer := c.builder.CreateCall(c.mallocFunc.Type().ElementType(), c.mallocFunc, []llvm.Value{bufferSize}, "new_string")
	c.builder.CreateCall(c.strcpyFunc.Type().ElementType(), c.strcpyFunc, []llvm.Value{newBuffer, left}, "")
	c.builder.CreateCall(c.strcatFunc.Type().ElementType(), c.strcatFunc, []llvm.Value{newBuffer, right}, "")
	return newBuffer
}

func (c *CodeGenerator) genPrintCall(call *ast.CallExpression) llvm.Value {
	arg := c.genExpression(call.Arguments[0])
	argType := arg.Type()
	var format llvm.Value
	finalArg := arg
	if argType.TypeKind() == llvm.IntegerTypeKind {
		format = c.builder.CreateGlobalStringPtr("%d\n", "fmt_int")
		if argType.IntTypeWidth() < 32 {
			finalArg = c.builder.CreateSExt(arg, c.context.Int32Type(), "printf_arg_promo")
		}
	} else if argType.TypeKind() == llvm.PointerTypeKind {
		format = c.builder.CreateGlobalStringPtr("%s\n", "fmt_str")
	} else {
		panic(fmt.Sprintf("tipo não suportado para a função print: %s", argType.String()))
	}
	return c.builder.CreateCall(c.printfFuncType, c.printfFunc, []llvm.Value{format, finalArg}, "printf_call")
}

func (c *CodeGenerator) genIfExpression(ie *ast.IfExpression) llvm.Value {
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
