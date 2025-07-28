package codegen

import (
	"fmt"
	"taquion/compiler/ast"

	"github.com/taquion-lang/go-llvm"
)

// genFunctionDeclaration gera código para a declaração de funções.
func (c *CodeGenerator) genFunctionDeclaration(node *ast.FunctionDeclaration) {
	var retType llvm.Type
	if node.Name.Value == "main" {
		retType = c.context.Int32Type()
	} else {
		// For now, let's assume other functions also return i32.
		// This might need to be determined from the AST later.
		retType = c.context.Int32Type()
	}
	c.currentFunctionReturnType = retType

	// ▼▼▼ START OF FIX ▼▼▼
	// Determine parameter types from the AST, not by hardcoding them.
	paramTypes := make([]llvm.Type, len(node.Parameters))
	for i, p := range node.Parameters {
		// This relies on the AST having the correct type information for each parameter.
		// Your ast.Identifier has a 'Type' field which should be an ast.Identifier itself (e.g., Value: "string").
		if p.Type == nil {
			panic(fmt.Sprintf("o parâmetro '%s' na função '%s' não possui um tipo definido na AST", p.Value, node.Name.Value))
		}
		// Use the existing type lookup utility.
		paramTypes[i] = c.lookupLLVMType(p.Type)
	}
	// ▲▲▲ END OF FIX ▲▲▲

	funcType := llvm.FunctionType(retType, paramTypes, false)
	// The rest of the function remains the same...
	function := llvm.AddFunction(c.module, node.Name.Value, funcType)
	c.setSymbol(node.Name.Value, SymbolEntry{Value: function, Typ: funcType, IsLiteral: true})

	if node.Body != nil {
		entryBlock := c.context.AddBasicBlock(function, "entry")
		c.builder.SetInsertPointAtEnd(entryBlock)
		c.pushScope()

		for i, param := range node.Parameters {
			paramValue := function.Param(i)
			paramValue.SetName(param.Value)
			// The alloca should also use the correct type.
			alloca := c.builder.CreateAlloca(paramTypes[i], param.Value)
			c.builder.CreateStore(paramValue, alloca)
			c.setSymbol(param.Value, SymbolEntry{Ptr: alloca, Typ: paramTypes[i], IsLiteral: false})
			c.setSymbol(param.Value, SymbolEntry{Ptr: alloca, Typ: paramTypes[i], TypeName: param.Type.Value, IsLiteral: false})
		}

		c.genStatement(node.Body)
		c.popScope()

		if !isBlockTerminated(c.builder.GetInsertBlock()) {
			if retType.TypeKind() == llvm.IntegerTypeKind {
				c.builder.CreateRet(llvm.ConstInt(retType, 0, false))
			}
		}
	}
}
