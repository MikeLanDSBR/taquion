package codegen

import (
	"fmt"
	"taquion/compiler/ast"

	"github.com/taquion-lang/go-llvm"
)

// genArrayLiteral gera código para um literal de array.
func (c *CodeGenerator) genArrayLiteral(node *ast.ArrayLiteral) llvm.Value {
	c.logTrace("Gerando ArrayLiteral")

	elemCount := len(node.Elements)
	elemType := c.context.Int32Type()
	arrayType := llvm.ArrayType(elemType, elemCount)

	arrayPtr := c.builder.CreateAlloca(arrayType, "array_tmp")
	c.logTrace(fmt.Sprintf("DEBUG: Alocando array na stack. Tipo: %v", arrayPtr.Type()))

	for i, elemExpr := range node.Elements {
		c.logTrace(fmt.Sprintf("DEBUG: Gerando elemento de array no índice %d", i))
		elemValue := c.genExpression(elemExpr)

		indices := []llvm.Value{
			llvm.ConstInt(c.context.Int32Type(), 0, false),
			llvm.ConstInt(c.context.Int32Type(), uint64(i), false),
		}
		elemPtr := c.builder.CreateInBoundsGEP(arrayType, arrayPtr, indices, fmt.Sprintf("array_elem_%d_ptr", i))
		c.logTrace(fmt.Sprintf("DEBUG: GEP para elemento %d. Endereço: %v", i, elemPtr))

		c.builder.CreateStore(elemValue, elemPtr)
	}
	c.logTrace(fmt.Sprintf("DEBUG: Finalizando ArrayLiteral. Retornando ponteiro para o array: %v", arrayPtr))
	return arrayPtr
}

// genIndexExpression gera código para o acesso a um elemento de array.
func (c *CodeGenerator) genIndexExpression(node *ast.IndexExpression) llvm.Value {
	c.logTrace("Gerando IndexExpression")

	arrayIdent, ok := node.Left.(*ast.Identifier)
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
	c.logTrace(fmt.Sprintf("DEBUG: Ponteiro para o array (valor da variável): %v", arrayPtr))

	indexValue := c.genExpression(node.Index)
	c.logTrace(fmt.Sprintf("DEBUG: Valor do índice: %v", indexValue))

	indices := []llvm.Value{
		llvm.ConstInt(c.context.Int32Type(), 0, false),
		indexValue,
	}

	// CORREÇÃO: Usa o tipo de array explicitamente armazenado na tabela de símbolos.
	elementPtr := c.builder.CreateInBoundsGEP(arrayEntry.ArrayType, arrayPtr, indices, "element_ptr")
	c.logTrace(fmt.Sprintf("DEBUG: GEP para elemento do array. Endereço: %v", elementPtr))

	return c.builder.CreateLoad(arrayEntry.ArrayType.ElementType(), elementPtr, "array_element_val")
}
