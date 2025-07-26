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

	case *ast.ConstStatement:
		c.logTrace(fmt.Sprintf("Gerando declaração 'const' para a constante '%s'", node.Name.Value))
		val := c.genExpression(node.Value)
		typ := val.Type()
		ptr := c.builder.CreateAlloca(typ, node.Name.Value)
		c.builder.CreateStore(val, ptr)
		c.setSymbol(node.Name.Value, SymbolEntry{Ptr: ptr, Typ: typ})

	case *ast.FunctionDeclaration:
		defer func() { c.currentFunctionReturnType = llvm.Type{} }()

		var retType llvm.Type
		inferredRetType := c.inferFunctionReturnType(node.Body)

		// --- CORREÇÃO PRINCIPAL ---
		// Força a função 'main' a sempre retornar i32, conforme a convenção do sistema operacional.
		if node.Name.Value == "main" {
			retType = c.context.Int32Type()
			c.logTrace("Função 'main' detectada. Forçando tipo de retorno para i32.")
		} else if !inferredRetType.IsNil() {
			retType = inferredRetType
		} else {
			retType = c.context.Int32Type() // Padrão para outras funções
		}

		c.currentFunctionReturnType = retType

		paramTypes := []llvm.Type{}
		for range node.Parameters {
			paramTypes = append(paramTypes, c.context.Int32Type())
		}

		funcType := llvm.FunctionType(retType, paramTypes, false)
		function := llvm.AddFunction(c.module, node.Name.Value, funcType)

		c.setSymbol(node.Name.Value, SymbolEntry{Ptr: function, Typ: funcType})

		if node.Body != nil {
			entryBlock := c.context.AddBasicBlock(function, "entry")
			prevBlock := c.builder.GetInsertBlock()
			c.builder.SetInsertPointAtEnd(entryBlock)

			c.pushScope()
			defer c.popScope()

			for i, param := range node.Parameters {
				llvmParam := function.Param(i)
				llvmParam.SetName(param.Name.Value)
				paramType := c.context.Int32Type()
				ptr := c.builder.CreateAlloca(paramType, param.Name.Value)
				c.builder.CreateStore(llvmParam, ptr)
				c.setSymbol(param.Name.Value, SymbolEntry{Ptr: ptr, Typ: paramType})
			}

			c.genStatement(node.Body)

			if !isBlockTerminated(c.builder.GetInsertBlock()) {
				if node.Name.Value == "main" && retType.TypeKind() == llvm.IntegerTypeKind {
					c.builder.CreateRet(llvm.ConstInt(retType, 0, false))
				}
			}

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
		valType := val.Type()

		expectedRetType := c.currentFunctionReturnType

		if !expectedRetType.IsNil() && valType.TypeKind() == llvm.IntegerTypeKind && valType != expectedRetType {
			c.logTrace(fmt.Sprintf("Convertendo tipo de retorno de %s para %s", valType.String(), expectedRetType.String()))
			val = c.builder.CreateZExt(val, expectedRetType, "retcast")
		}

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

	default:
		panic(fmt.Sprintf("Declaração não suportada: %T\n", node))
	}
}

func (c *CodeGenerator) inferFunctionReturnType(body *ast.BlockStatement) llvm.Type {
	if body == nil {
		return llvm.Type{}
	}

	for i := len(body.Statements) - 1; i >= 0; i-- {
		if retStmt, ok := body.Statements[i].(*ast.ReturnStatement); ok {
			switch retValNode := retStmt.ReturnValue.(type) {
			case *ast.IntegerLiteral, *ast.StringLiteral, *ast.BooleanLiteral:
				val := c.genExpression(retValNode)
				return val.Type()
			default:
				c.logTrace(fmt.Sprintf("[infer] Não é possível inferir o tipo de retorno para %T. Usando tipo padrão.", retValNode))
				return llvm.Type{}
			}
		}
	}
	return llvm.Type{}
}
