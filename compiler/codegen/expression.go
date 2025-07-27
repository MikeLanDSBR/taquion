package codegen

import (
	"fmt"
	"taquion/compiler/ast"

	"github.com/MikeLanDSBR/go-llvm"
)

// genExpression é o ponto de entrada para a geração de código de uma expressão.
func (c *CodeGenerator) genExpression(expr ast.Expression) llvm.Value {
	defer c.trace(fmt.Sprintf("genExpression (%T)", expr))()

	switch node := expr.(type) {
	case *ast.IntegerLiteral:
		return c.genIntegerLiteral(node)
	case *ast.StringLiteral:
		return c.genStringLiteral(node)
	case *ast.BooleanLiteral:
		return c.genBooleanLiteral(node)
	case *ast.Identifier:
		return c.genIdentifier(node)
	case *ast.InfixExpression:
		return c.genInfixExpression(node)
	case *ast.AssignmentExpression:
		return c.genAssignmentExpression(node)
	case *ast.CallExpression:
		return c.genCallExpression(node)
	case *ast.IfExpression:
		return c.genIfExpression(node)
	case *ast.ArrayLiteral:
		return c.genArrayLiteral(node)
	case *ast.IndexExpression:
		return c.genIndexExpression(node)
	case *ast.PrefixExpression:
		return c.genPrefixExpression(node)
	case *ast.FunctionLiteral:
		return c.genFunctionLiteral(node)
	default:
		panic(fmt.Sprintf("Expressão não suportada: %T\n", node))
	}
}
