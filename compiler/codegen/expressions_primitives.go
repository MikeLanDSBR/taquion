package codegen

import (
	"fmt"
	"strconv"
	"taquion/compiler/ast"

	"github.com/MikeLanDSBR/go-llvm"
)

// genIntegerLiteral gera um literal inteiro.
func (c *CodeGenerator) genIntegerLiteral(node *ast.IntegerLiteral) llvm.Value {
	val, _ := strconv.ParseInt(node.TokenLiteral(), 10, 64)
	c.logTrace(fmt.Sprintf("DEBUG: Gerando literal inteiro: %d", val))
	return llvm.ConstInt(c.context.Int32Type(), uint64(val), false)
}

// genStringLiteral gera um literal de string.
func (c *CodeGenerator) genStringLiteral(node *ast.StringLiteral) llvm.Value {
	i8PtrType := llvm.PointerType(c.context.Int8Type(), 0)
	c.logTrace(fmt.Sprintf("DEBUG: Gerando StringLiteral para: '%s'", node.Value))
	globalStringPtr := c.builder.CreateGlobalStringPtr(node.Value, "str_literal")
	c.logTrace(fmt.Sprintf("DEBUG: GlobalStringPtr criado. Valor: %v, Tipo: %v", globalStringPtr, c.GetValueTypeSafe(globalStringPtr)))
	return c.builder.CreatePointerCast(globalStringPtr, i8PtrType, "str_literal_to_i8ptr")
}

// genBooleanLiteral gera um literal booleano.
func (c *CodeGenerator) genBooleanLiteral(node *ast.BooleanLiteral) llvm.Value {
	c.logTrace(fmt.Sprintf("DEBUG: Gerando literal booleano: %v", node.Value))
	if node.Value {
		return llvm.ConstInt(c.context.Int1Type(), 1, false)
	}
	return llvm.ConstInt(c.context.Int1Type(), 0, false)
}

// genIdentifier gera o código para um identificador (carrega o valor da memória).
func (c *CodeGenerator) genIdentifier(node *ast.Identifier) llvm.Value {
	entry, ok := c.getSymbol(node.Value)
	if !ok {
		panic(fmt.Sprintf("variável não definida: %s", node.Value))
	}
	c.logTrace(fmt.Sprintf("DEBUG: Símbolo '%s' encontrado. IsLiteral: %t, Ptr: %v, Value: %v, Typ: %v", node.Value, entry.IsLiteral, entry.Ptr, entry.Value, entry.Typ))

	if entry.IsLiteral {
		c.logTrace(fmt.Sprintf("DEBUG: Símbolo '%s' é um literal/função. Retornando valor: %v", node.Value, entry.Value))
		if !entry.Ptr.IsNil() {
			c.logTrace(fmt.Sprintf("DEBUG: Símbolo literal é um ponteiro. Carregando valor. Tipo do ponteiro: %v, Tipo do valor: %v", c.GetValueTypeSafe(entry.Ptr), entry.Typ))
			return c.builder.CreateLoad(entry.Typ, entry.Ptr, node.Value)
		}
		return entry.Value
	}

	c.logTrace(fmt.Sprintf("DEBUG: Símbolo '%s' é uma variável. Carregando do ponteiro: %v. Tipo do valor: %v", node.Value, entry.Ptr, entry.Typ))
	loadedValue := c.builder.CreateLoad(entry.Typ, entry.Ptr, node.Value)
	return loadedValue
}
