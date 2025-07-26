package codegen

import (
	"fmt"
	"strconv"
	"taquion/compiler/ast"
	"taquion/compiler/token"

	"github.com/MikeLanDSBR/go-llvm"
)

func (c *CodeGenerator) genExpression(expr ast.Expression) llvm.Value {
	defer c.trace(fmt.Sprintf("genExpression (%T)", expr))()

	i8PtrType := llvm.PointerType(c.context.Int8Type(), 0)

	switch node := expr.(type) {
	case *ast.IntegerLiteral:
		val, _ := strconv.ParseInt(node.TokenLiteral(), 10, 64)
		return llvm.ConstInt(c.context.Int32Type(), uint64(val), false)

	case *ast.StringLiteral:
		c.logTrace(fmt.Sprintf("DEBUG: Gerando StringLiteral para: '%s'", node.Value))
		globalStringPtr := c.builder.CreateGlobalStringPtr(node.Value, "str_literal")
		c.logTrace(fmt.Sprintf("DEBUG: GlobalStringPtr criado. Valor: %v, Tipo: %v", globalStringPtr, c.GetValueTypeSafe(globalStringPtr)))
		return c.builder.CreatePointerCast(globalStringPtr, i8PtrType, "str_literal_to_i8ptr")

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
		c.logTrace(fmt.Sprintf("DEBUG: Símbolo '%s' encontrado. IsLiteral: %t, Ptr: %v, Value: %v, Typ: %v", node.Value, entry.IsLiteral, entry.Ptr, entry.Value, entry.Typ))

		if entry.IsLiteral {
			c.logTrace(fmt.Sprintf("DEBUG: Símbolo '%s' é um literal/função. Retornando valor: %v", node.Value, entry.Value))
			// Se for um literal que é um ponteiro (string alocada), precisa ser carregado
			if !entry.Ptr.IsNil() {
				return c.builder.CreateLoad(entry.Typ, entry.Ptr, node.Value)
			}
			return entry.Value
		}

		c.logTrace(fmt.Sprintf("DEBUG: Símbolo '%s' é uma variável. Carregando do ponteiro: %v", node.Value, entry.Ptr))
		loadedValue := c.builder.CreateLoad(entry.Typ, entry.Ptr, node.Value)
		return loadedValue

	case *ast.InfixExpression:
		left := c.genExpression(node.Left)
		right := c.genExpression(node.Right)

		isLeftString := c.GetValueTypeSafe(left).TypeKind() == llvm.PointerTypeKind
		isRightString := c.GetValueTypeSafe(right).TypeKind() == llvm.PointerTypeKind

		if node.Operator == token.PLUS && isLeftString && isRightString {
			c.logTrace(fmt.Sprintf("DEBUG: Entrando em genStringConcat para '%s' + '%s'", node.Left.String(), node.Right.String()))
			return c.genStringConcat(left, right)
		} else {
			c.logTrace(fmt.Sprintf("DEBUG: Entrando no switch de operadores aritméticos para '%s'", node.Operator))
			switch node.Operator {
			case token.PLUS:
				return c.builder.CreateAdd(left, right, "addtmp")
			case token.MINUS:
				return c.builder.CreateSub(left, right, "subtmp")
			case token.ASTERISK:
				return c.builder.CreateMul(left, right, "multmp")
			case token.SLASH:
				return c.builder.CreateSDiv(left, right, "divtmp")
			case token.MODULO:
				return c.builder.CreateSRem(left, right, "modtmp")
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
		if entry.IsLiteral {
			panic(fmt.Sprintf("atribuição a constante não é permitida: %s", ident.Value))
		}
		c.builder.CreateStore(val, entry.Ptr)
		return val

	case *ast.CallExpression:
		if node.Function.String() == "print" {
			return c.genPrintCall(node)
		}

		symbol, ok := c.getSymbol(node.Function.String())
		if !ok {
			panic(fmt.Sprintf("função não definida: %s", node.Function.String()))
		}

		function := symbol.Value
		functionType := symbol.Typ

		args := []llvm.Value{}
		for _, argExpr := range node.Arguments {
			args = append(args, c.genExpression(argExpr))
		}

		return c.builder.CreateCall(functionType, function, args, "calltmp")

	case *ast.IfExpression:
		return c.genIfExpression(node)

	case *ast.ArrayLiteral:
		panic("Expressão não suportada: *ast.ArrayLiteral")

	default:
		panic(fmt.Sprintf("Expressão não suportada: %T\n", node))
	}
}

func (c *CodeGenerator) genStringConcat(left, right llvm.Value) llvm.Value {
	c.logTrace("Gerando concatenação de strings")

	i8PtrType := llvm.PointerType(c.context.Int8Type(), 0)
	sizeType := c.context.Int64Type()

	strlenType := llvm.FunctionType(sizeType, []llvm.Type{i8PtrType}, false)
	mallocType := llvm.FunctionType(i8PtrType, []llvm.Type{sizeType}, false)
	strcpyType := llvm.FunctionType(i8PtrType, []llvm.Type{i8PtrType, i8PtrType}, false)
	strcatType := llvm.FunctionType(i8PtrType, []llvm.Type{i8PtrType, i8PtrType}, false)

	finalLeft := c.builder.CreatePointerCast(left, i8PtrType, "str_concat_left_cast")
	finalRight := c.builder.CreatePointerCast(right, i8PtrType, "str_concat_right_cast")
	c.logTrace(fmt.Sprintf("DEBUG: Argumentos de concatenação: finalLeft=%v, finalRight=%v", finalLeft, finalRight))

	len1 := c.builder.CreateCall(strlenType, c.strlenFunc, []llvm.Value{finalLeft}, "len1")
	len2 := c.builder.CreateCall(strlenType, c.strlenFunc, []llvm.Value{finalRight}, "len2")
	totalLen := c.builder.CreateAdd(len1, len2, "totalLen")
	bufferSize := c.builder.CreateAdd(totalLen, llvm.ConstInt(c.context.Int64Type(), 1, false), "bufferSize")
	newBuffer := c.builder.CreateCall(mallocType, c.mallocFunc, []llvm.Value{bufferSize}, "new_string")

	c.builder.CreateCall(strcpyType, c.strcpyFunc, []llvm.Value{newBuffer, finalLeft}, "")
	c.builder.CreateCall(strcatType, c.strcatFunc, []llvm.Value{newBuffer, finalRight}, "")

	c.logTrace(fmt.Sprintf("DEBUG: Concatenação completa. Retornando novo buffer: %v", newBuffer))

	return newBuffer
}

// genPrintCall foi refatorada para ser mais robusta
func (c *CodeGenerator) genPrintCall(call *ast.CallExpression) llvm.Value {
	arg := c.genExpression(call.Arguments[0])
	argType := c.GetValueTypeSafe(arg)
	var format llvm.Value
	finalArg := arg

	if arg.IsNil() {
		panic(fmt.Sprintf("valor nulo passado para a função print"))
	}

	if argType.IsNil() {
		panic(fmt.Sprintf("tipo nulo para o argumento da função print: %v", arg))
	}

	switch argType.TypeKind() {
	case llvm.IntegerTypeKind:
		format = c.builder.CreateGlobalStringPtr("%d\n", "fmt_int")
		if argType.IntTypeWidth() < 32 {
			finalArg = c.builder.CreateSExt(arg, c.context.Int32Type(), "printf_arg_promo")
		}
	case llvm.PointerTypeKind:
		format = c.builder.CreateGlobalStringPtr("%s\n", "fmt_str")
		finalArg = arg
	default:
		i8PtrType := llvm.PointerType(c.context.Int8Type(), 0)
		finalArg = c.builder.CreatePointerCast(arg, i8PtrType, "printf_arg_forced_cast")
		format = c.builder.CreateGlobalStringPtr("%s\n", "fmt_str")
	}

	printfFuncType := llvm.FunctionType(c.context.Int32Type(), []llvm.Type{llvm.PointerType(c.context.Int8Type(), 0)}, true)

	return c.builder.CreateCall(printfFuncType, c.printfFunc, []llvm.Value{format, finalArg}, "printf_call")
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
