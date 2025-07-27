package codegen

import (
	"github.com/MikeLanDSBR/go-llvm"
)

// declareCFunctions declara as funções C externas necessárias para o compilador.
func (c *CodeGenerator) declareCFunctions() {
	i8PtrType := llvm.PointerType(c.context.Int8Type(), 0)
	sizeType := c.context.Int64Type()

	c.printfFuncType = llvm.FunctionType(c.context.Int32Type(), []llvm.Type{i8PtrType}, true)
	c.printfFunc = llvm.AddFunction(c.module, "printf", c.printfFuncType)

	mallocType := llvm.FunctionType(i8PtrType, []llvm.Type{sizeType}, false)
	c.mallocFunc = llvm.AddFunction(c.module, "malloc", mallocType)

	strlenType := llvm.FunctionType(sizeType, []llvm.Type{i8PtrType}, false)
	c.strlenFunc = llvm.AddFunction(c.module, "strlen", strlenType)

	strcpyType := llvm.FunctionType(i8PtrType, []llvm.Type{i8PtrType, i8PtrType}, false)
	c.strcpyFunc = llvm.AddFunction(c.module, "strcpy", strcpyType)
	c.strcatFunc = llvm.AddFunction(c.module, "strcat", strcpyType)
}
