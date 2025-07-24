// Arquivo: codegen/expression.go
package codegen

import (
	"fmt"
	"strconv"
	"taquion/compiler/ast"
	"taquion/compiler/token"

	"tinygo.org/x/go-llvm"
)

// genExpression gera código LLVM IR para uma expressão AST e retorna o llvm.Value resultante.
func (c *CodeGenerator) genExpression(expr ast.Expression) llvm.Value {
	switch node := expr.(type) {
	case *ast.IntegerLiteral:
		val, _ := strconv.ParseInt(node.TokenLiteral(), 10, 64)
		return llvm.ConstInt(c.context.Int32Type(), uint64(val), false)

	case *ast.Identifier:
		if ptr, ok := c.symbolTable[node.Value]; ok {
			return c.builder.CreateLoad(c.context.Int32Type(), ptr, node.Value+"_val")
		}
		panic(fmt.Sprintf("variável não definida: %s", node.Value))

	// O case para AssignmentExpression foi removido daqui.

	case *ast.InfixExpression:
		left := c.genExpression(node.Left)
		right := c.genExpression(node.Right)
		switch node.Operator {
		case token.PLUS:
			return c.builder.CreateAdd(left, right, "addtmp")
		case token.MINUS:
			return c.builder.CreateSub(left, right, "subtmp")
		case token.ASTERISK:
			return c.builder.CreateMul(left, right, "multmp")
		case token.SLASH:
			return c.builder.CreateSDiv(left, right, "divtmp")
		case token.GT:
			return c.builder.CreateICmp(llvm.IntSGT, left, right, "gttmp")
		case token.LT:
			return c.builder.CreateICmp(llvm.IntSLT, left, right, "lttmp")
		case token.EQ:
			return c.builder.CreateICmp(llvm.IntEQ, left, right, "eqtmp")
		case token.NOT_EQ:
			return c.builder.CreateICmp(llvm.IntNE, left, right, "neqtmp")
		default:
			panic(fmt.Sprintf("operador infix não suportado: %s", node.Operator))
		}

	case *ast.IfExpression:
		return c.genIfExpression(node)

	default:
		fmt.Printf("Expressão não suportada: %T\n", node)
		return llvm.Value{}
	}
}

// genIfExpression permanece o mesmo
func (c *CodeGenerator) genIfExpression(ie *ast.IfExpression) llvm.Value {
	condVal := c.genExpression(ie.Condition)
	function := c.builder.GetInsertBlock().Parent()
	thenBlock := llvm.AddBasicBlock(function, "then")
	elseBlock := llvm.AddBasicBlock(function, "else")
	mergeBlock := llvm.AddBasicBlock(function, "merge")
	c.builder.CreateCondBr(condVal, thenBlock, elseBlock)
	c.builder.SetInsertPointAtEnd(thenBlock)
	for _, stmt := range ie.Consequence.Statements {
		c.genStatement(stmt)
	}
	if c.builder.GetInsertBlock().LastInstruction().IsNil() || c.builder.GetInsertBlock().LastInstruction().IsAReturnInst().IsNil() {
		c.builder.CreateBr(mergeBlock)
	}
	c.builder.SetInsertPointAtEnd(elseBlock)
	if ie.Alternative != nil {
		for _, stmt := range ie.Alternative.Statements {
			c.genStatement(stmt)
		}
	}
	if c.builder.GetInsertBlock().LastInstruction().IsNil() || c.builder.GetInsertBlock().LastInstruction().IsAReturnInst().IsNil() {
		c.builder.CreateBr(mergeBlock)
	}
	c.builder.SetInsertPointAtEnd(mergeBlock)
	return llvm.Value{}
}
