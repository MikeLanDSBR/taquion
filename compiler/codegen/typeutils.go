package codegen

import (
	"fmt"
	"taquion/compiler/ast"

	"github.com/MikeLanDSBR/go-llvm"
)

// Retorna (ou panica) o llvm.Type previamente registrado para a struct.
func (c *CodeGenerator) getLLVMStructType(name string) llvm.Type {
	ty, ok := c.structTypes[name]
	if !ok || ty.IsNil() {
		panic(fmt.Sprintf("tipo de struct não encontrado: %s", name))
	}
	return ty
}

// Converte um tipo da AST para um tipo LLVM.
// (ajuste se no seu AST existir um nó específico para tipos; aqui usei ast.Expression
// porque foi isso que você mostrou que tem em StructField.Type).
func (c *CodeGenerator) lookupLLVMType(t ast.Expression) llvm.Type {
	switch tt := t.(type) {
	case *ast.Identifier:
		switch tt.Value {
		case "int", "int32":
			return c.context.Int32Type()
		case "int8":
			return c.context.Int8Type()
		case "bool":
			return c.context.Int1Type()
		case "string":
			return llvm.PointerType(c.context.Int8Type(), 0)
		default:
			// struct definida pelo usuário
			return c.getLLVMStructType(tt.Value)
		}
	default:
		panic(fmt.Sprintf("tipo não suportado: %T", t))
	}
}

// Garante que o llvm.StructType da struct já está criado e registrado.
// Chame isso no início de genTypeDeclaration.
func (c *CodeGenerator) ensureStructType(node *ast.TypeDeclaration) {
	name := node.Name.Value
	if ty, ok := c.structTypes[name]; ok && !ty.IsNil() {
		return
	}

	st := c.context.StructCreateNamed(name)
	c.structTypes[name] = st
	c.structFieldIndices[name] = make(map[string]int) // ▼▼▼ INICIALIZE O MAPA INTERNO ▼▼▼

	fieldLLVM := make([]llvm.Type, len(node.Fields))
	for i, f := range node.Fields {
		fieldLLVM[i] = c.lookupLLVMType(f.Type)
		c.structFieldIndices[name][f.Name.Value] = i // ▼▼▼ GUARDE O ÍNDICE DO CAMPO ▼▼▼
	}
	st.StructSetBody(fieldLLVM, false)
}
