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

	i8PtrType := llvm.PointerType(c.context.Int8Type(), 0) // Define i8PtrType uma vez

	switch node := expr.(type) {
	case *ast.IntegerLiteral:
		val, _ := strconv.ParseInt(node.TokenLiteral(), 10, 64)
		// Simplificado para i32 por enquanto para evitar problemas de promoção
		return llvm.ConstInt(c.context.Int32Type(), uint64(val), false)

	case *ast.StringLiteral:
		// Cria a string global. O resultado é um ponteiro para o array ([N x i8]*).
		globalStringPtr := c.builder.CreateGlobalStringPtr(node.Value, "str_literal")
		// Converte o ponteiro para o array ([N x i8]*) para um i8*
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

		// Se o símbolo é uma função, retorna o valor da função diretamente.
		if entry.Typ.TypeKind() == llvm.FunctionTypeKind {
			c.logTrace(fmt.Sprintf("DEBUG: Identifier '%s' is Function. Type: %s", node.Value, entry.Typ.String()))
			return entry.Ptr
		}

		// Carrega o valor da variável.
		loadedValue := c.builder.CreateLoad(entry.Typ, entry.Ptr, node.Value)

		// Verifica se o tipo carregado é um ponteiro.
		if loadedValue.Type().TypeKind() == llvm.PointerTypeKind {
			// Caso 1: Ponteiro para um array de chars (ex: string literal global armazenada em const)
			if loadedValue.Type().ElementType().TypeKind() == llvm.ArrayTypeKind &&
				loadedValue.Type().ElementType().ElementType().TypeKind() == llvm.IntegerTypeKind &&
				loadedValue.Type().ElementType().ElementType().IntTypeWidth() == 8 {

				zero := llvm.ConstInt(c.context.Int32Type(), 0, false)
				// GEP para obter um ponteiro para o primeiro char do array (i8*)
				gepPtr := c.builder.CreateInBoundsGEP(loadedValue.Type().ElementType(), loadedValue, []llvm.Value{zero, zero}, "str_ptr_from_array")
				return c.builder.CreatePointerCast(gepPtr, i8PtrType, "gep_to_i8ptr") // Garante que o GEP é i8*
			} else if loadedValue.Type() == i8PtrType {
				// Caso 2: Já é um i8*, retorna diretamente.
				return loadedValue
			}
			// Se for um ponteiro para outro tipo (que não é string), retorna o valor carregado como está.
			// Isso é importante para ponteiros para inteiros, etc.
			// Não tentamos um cast para i8* aqui, pois isso pode ser incorreto para outros tipos de ponteiro.
			return loadedValue
		}
		// Para tipos não-ponteiro (int, bool, etc.), retorna o valor carregado como está.
		return loadedValue

	case *ast.InfixExpression:
		left := c.genExpression(node.Left)
		right := c.genExpression(node.Right)

		// Adicionado logging para depuração
		// Verifica se left e right são válidos e se Type() retorna um tipo válido antes de chamar Type().String()
		leftTypeStr := "UNKNOWN_OR_INVALID_TYPE"
		leftTypeKindStr := "UNKNOWN_KIND"
		if !left.IsNil() && !left.Type().IsNil() {
			leftTypeStr = left.Type().String()
			// Adiciona uma verificação para TypeKind para evitar pânico em tipos desconhecidos
			if left.Type().TypeKind() >= llvm.VoidTypeKind && left.Type().TypeKind() <= llvm.MetadataTypeKind {
				leftTypeKindStr = left.Type().TypeKind().String()
			} else {
				leftTypeKindStr = fmt.Sprintf("UNRECOGNIZED_TYPE_KIND(%d)", left.Type().TypeKind())
			}
		} else if !left.IsNil() && left.Type().IsNil() { // Caso onde Type() é nulo, mas Value não é
			leftTypeStr = fmt.Sprintf("NIL_TYPE_FOR_VALUE(%v)", left)
		}

		rightTypeStr := "UNKNOWN_OR_INVALID_TYPE"
		rightTypeKindStr := "UNKNOWN_KIND"
		if !right.IsNil() && !right.Type().IsNil() {
			rightTypeStr = right.Type().String()
			if right.Type().TypeKind() >= llvm.VoidTypeKind && right.Type().TypeKind() <= llvm.MetadataTypeKind {
				rightTypeKindStr = right.Type().TypeKind().String()
			} else {
				rightTypeKindStr = fmt.Sprintf("UNRECOGNIZED_TYPE_KIND(%d)", right.Type().TypeKind())
			}
		} else if !right.IsNil() && right.Type().IsNil() { // Caso onde Type() é nulo, mas Value não é
			rightTypeStr = fmt.Sprintf("NIL_TYPE_FOR_VALUE(%v)", right)
		}
		c.logTrace(fmt.Sprintf("DEBUG: Infix '%s'. Left Type: %s (Kind: %s), Right Type: %s (Kind: %s)", node.Operator, leftTypeStr, leftTypeKindStr, rightTypeStr, rightTypeKindStr))

		// A verificação de string agora é mais simples, pois genExpression já deve retornar i8*
		isLeftString := left.Type() == i8PtrType
		isRightString := right.Type() == i8PtrType

		c.logTrace(fmt.Sprintf("DEBUG: isLeftString: %t, isRightString: %t", isLeftString, isRightString))

		if node.Operator == token.PLUS && isLeftString && isRightString {
			c.logTrace(fmt.Sprintf("DEBUG: Entrando em genStringConcat para '%s' + '%s'", node.Left.String(), node.Right.String()))
			return c.genStringConcat(left, right)
		} else { // Adicionado o bloco 'else' para garantir exclusividade
			c.logTrace(fmt.Sprintf("DEBUG: Entrando no switch de operadores aritméticos para '%s'", node.Operator))
			switch node.Operator {
			case token.PLUS:
				c.logTrace(fmt.Sprintf("DEBUG: Gerando ADD para tipos: %s e %s", left.Type().String(), right.Type().String()))
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

	// As funções de string C (strlen, strcpy, strcat) esperam i8*
	// Certifique-se de que 'left' e 'right' são i8*
	i8PtrType := llvm.PointerType(c.context.Int8Type(), 0)

	// Garante que 'left' e 'right' são i8* usando CreatePointerCast.
	// Isso é necessário porque o tipo pode ser um ponteiro para um array ([N x i8]*),
	// e precisamos de um ponteiro para o primeiro elemento (i8*).
	// CreatePointerCast pode converter entre tipos de ponteiro.
	finalLeft := c.builder.CreatePointerCast(left, i8PtrType, "str_concat_left_cast")
	finalRight := c.builder.CreatePointerCast(right, i8PtrType, "str_concat_right_cast")

	len1 := c.builder.CreateCall(c.strlenFunc.Type(), c.strlenFunc, []llvm.Value{finalLeft}, "len1")
	len2 := c.builder.CreateCall(c.strlenFunc.Type(), c.strlenFunc, []llvm.Value{finalRight}, "len2")
	totalLen := c.builder.CreateAdd(len1, len2, "totalLen")
	bufferSize := c.builder.CreateAdd(totalLen, llvm.ConstInt(c.context.Int64Type(), 1, false), "bufferSize")
	newBuffer := c.builder.CreateCall(c.mallocFunc.Type(), c.mallocFunc, []llvm.Value{bufferSize}, "new_string")

	// O newBuffer retornado por malloc já é i8*, então não precisa de cast adicional
	// a menos que mallocType seja diferente de i8PtrType (o que não é o caso aqui).

	c.builder.CreateCall(c.strcpyFunc.Type(), c.strcpyFunc, []llvm.Value{newBuffer, finalLeft}, "")
	c.builder.CreateCall(c.strcatFunc.Type(), c.strcatFunc, []llvm.Value{newBuffer, finalRight}, "")
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
		// Garante que a string passada para printf seja i8*
		i8PtrType := llvm.PointerType(c.context.Int8Type(), 0)
		if arg.Type() != i8PtrType {
			finalArg = c.builder.CreatePointerCast(arg, i8PtrType, "printf_arg_str_cast")
		}
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
