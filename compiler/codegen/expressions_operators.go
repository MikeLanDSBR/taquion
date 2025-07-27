package codegen

import (
	"fmt"
	"taquion/compiler/ast"
	"taquion/compiler/token"

	"github.com/MikeLanDSBR/go-llvm"
)

// genInfixExpression gera o código para uma expressão infixa.
func (c *CodeGenerator) genInfixExpression(node *ast.InfixExpression) llvm.Value {
	c.logTrace(fmt.Sprintf("DEBUG: Gerando expressão infix: %s", node.Operator))
	left := c.genExpression(node.Left)
	right := c.genExpression(node.Right)
	c.logTrace(fmt.Sprintf("DEBUG: Operandos da expressão infix: left=%v, right=%v", left, right))

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
}

// genAssignmentExpression gera código para uma atribuição.
func (c *CodeGenerator) genAssignmentExpression(node *ast.AssignmentExpression) llvm.Value {
	c.logTrace("DEBUG: Gerando expressão de atribuição")
	val := c.genExpression(node.Value)

	if ident, ok := node.Left.(*ast.Identifier); ok {
		c.logTrace(fmt.Sprintf("DEBUG: Atribuindo a um identificador: %s", ident.Value))
		entry, ok := c.getSymbol(ident.Value)
		if !ok {
			panic(fmt.Sprintf("atribuição a variável não declarada: %s", ident.Value))
		}
		if entry.IsLiteral {
			panic(fmt.Sprintf("atribuição a constante não é permitida: %s", ident.Value))
		}
		c.builder.CreateStore(val, entry.Ptr)
		return val
	}

	if indexExpr, ok := node.Left.(*ast.IndexExpression); ok {
		c.logTrace("DEBUG: Atribuindo a um elemento de array")
		arrayIdent, ok := indexExpr.Left.(*ast.Identifier)
		if !ok {
			panic("o lado esquerdo de uma expressão de índice deve ser um identificador")
		}
		arrayEntry, ok := c.getSymbol(arrayIdent.Value)
		if !ok {
			panic(fmt.Sprintf("array não declarado: %s", arrayIdent.Value))
		}

		if arrayEntry.ArrayType.IsNil() {
			panic(fmt.Sprintf("a variável '%s' não é um array indexável", arrayIdent.Value))
		}

		arrayPtr := c.genExpression(arrayIdent)
		indexValue := c.genExpression(indexExpr.Index)

		indices := []llvm.Value{
			llvm.ConstInt(c.context.Int32Type(), 0, false),
			indexValue,
		}

		// CORREÇÃO: Usa o tipo de array explicitamente armazenado na tabela de símbolos.
		elementPtr := c.builder.CreateInBoundsGEP(arrayEntry.ArrayType, arrayPtr, indices, "array_element_ptr")
		c.builder.CreateStore(val, elementPtr)
		return val
	}

	panic("o lado esquerdo de uma atribuição deve ser um identificador ou um índice de array")
}

// genCallExpression gera código para uma chamada de função.
func (c *CodeGenerator) genCallExpression(node *ast.CallExpression) llvm.Value {
	c.logTrace(fmt.Sprintf("DEBUG: Gerando chamada de função: %s", node.Function.String()))
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
}

// genIfExpression gera código para uma expressão `if`.
func (c *CodeGenerator) genIfExpression(node *ast.IfExpression) llvm.Value {
	c.logTrace("DEBUG: Gerando expressão 'if'")
	cond := c.genExpression(node.Condition)
	function := c.builder.GetInsertBlock().Parent()
	thenBlock := c.context.AddBasicBlock(function, "then")
	elseBlock := c.context.AddBasicBlock(function, "else")
	mergeBlock := c.context.AddBasicBlock(function, "merge")

	c.builder.CreateCondBr(cond, thenBlock, elseBlock)

	c.builder.SetInsertPointAtEnd(thenBlock)
	c.genStatement(node.Consequence)
	if !isBlockTerminated(c.builder.GetInsertBlock()) {
		c.builder.CreateBr(mergeBlock)
	}

	c.builder.SetInsertPointAtEnd(elseBlock)
	if node.Alternative != nil {
		c.genStatement(node.Alternative)
	}
	if !isBlockTerminated(c.builder.GetInsertBlock()) {
		c.builder.CreateBr(mergeBlock)
	}

	c.builder.SetInsertPointAtEnd(mergeBlock)
	return llvm.Value{}
}

// genPrefixExpression gera código para uma expressão de prefixo.
func (c *CodeGenerator) genPrefixExpression(node *ast.PrefixExpression) llvm.Value {
	c.logTrace(fmt.Sprintf("DEBUG: Gerando expressão prefixo: %s", node.Operator))
	right := c.genExpression(node.Right)
	switch node.Operator {
	case "-":
		return c.builder.CreateNeg(right, "negtmp")
	case "!":
		return c.builder.CreateNot(right, "nottmp")
	default:
		panic(fmt.Sprintf("operador prefixo não suportado: %s", node.Operator))
	}
}

// genFunctionLiteral gera código para uma função anônima (literal).
func (c *CodeGenerator) genFunctionLiteral(node *ast.FunctionLiteral) llvm.Value {
	// Implementação pendente
	return llvm.Value{}
}

// genPrintCall gera código para a chamada da função `print`.
func (c *CodeGenerator) genPrintCall(call *ast.CallExpression) llvm.Value {
	c.logTrace("DEBUG: Gerando chamada para a função 'print'")
	arg := c.genExpression(call.Arguments[0])
	argType := c.GetValueTypeSafe(arg)
	var format llvm.Value
	finalArg := arg

	if arg.IsNil() {
		panic(fmt.Sprintf("argumento nulo para a função print: %v", call.Arguments[0]))
	}

	if argType.IsNil() {
		panic(fmt.Sprintf("tipo nulo para o argumento da função print: %v", arg))
	}

	switch argType.TypeKind() {
	case llvm.IntegerTypeKind:
		c.logTrace("DEBUG: Argumento de impressão é um inteiro.")
		format = c.builder.CreateGlobalStringPtr("%d\n", "fmt_int")
		if argType.IntTypeWidth() < 32 {
			finalArg = c.builder.CreateSExt(arg, c.context.Int32Type(), "printf_arg_promo")
		}
	case llvm.PointerTypeKind:
		c.logTrace("DEBUG: Argumento de impressão é um ponteiro (string ou array).")
		format = c.builder.CreateGlobalStringPtr("%s\n", "fmt_str")
		finalArg = arg
	default:
		c.logTrace("DEBUG: Argumento de impressão é um tipo desconhecido, tentando conversão para string.")
		i8PtrType := llvm.PointerType(c.context.Int8Type(), 0)
		finalArg = c.builder.CreatePointerCast(arg, i8PtrType, "printf_arg_forced_cast")
		format = c.builder.CreateGlobalStringPtr("%s\n", "fmt_str")
	}

	printfFuncType := llvm.FunctionType(c.context.Int32Type(), []llvm.Type{llvm.PointerType(c.context.Int8Type(), 0)}, true)

	return c.builder.CreateCall(printfFuncType, c.printfFunc, []llvm.Value{format, finalArg}, "printf_call")
}

// genStringConcat gera código para a concatenação de strings.
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
