package codegen

import (
	"fmt"
	"taquion/compiler/ast"

	"github.com/MikeLanDSBR/go-llvm"
)

// genStatement é o ponto de entrada para a geração de código de uma declaração.
func (c *CodeGenerator) genStatement(stmt ast.Statement) {
	defer c.trace(fmt.Sprintf("genStatement (%T)", stmt))()

	switch node := stmt.(type) {
	case *ast.PackageStatement:
		c.genPackageStatement(node)
	case *ast.LetStatement:
		c.genLetStatement(node)
	case *ast.ConstStatement:
		c.genConstStatement(node)
	case *ast.ReturnStatement:
		c.genReturnStatement(node)
	case *ast.ExpressionStatement:
		c.genExpressionStatement(node)
	case *ast.BlockStatement:
		c.genBlockStatement(node)
	case *ast.FunctionDeclaration:
		c.genFunctionDeclaration(node)
	case *ast.WhileStatement:
		c.genWhileStatement(node)
	case *ast.BreakStatement:
		c.genBreakStatement(node)
	case *ast.ContinueStatement:
		c.genContinueStatement(node)
	default:
		panic(fmt.Sprintf("Declaração não suportada: %T\n", node))
	}
}

// genPackageStatement ignora a declaração de pacote.
func (c *CodeGenerator) genPackageStatement(node *ast.PackageStatement) {
	c.logTrace(fmt.Sprintf("Ignorando declaração de pacote: package %s", node.Name.Value))
}

// genLetStatement gera código para a declaração de variáveis `let`.
func (c *CodeGenerator) genLetStatement(node *ast.LetStatement) {
	c.logTrace(fmt.Sprintf("Gerando declaração 'let' para a variável '%s'", node.Name.Value))

	val := c.genExpression(node.Value)
	valType := c.GetValueTypeSafe(val)
	if valType.IsNil() {
		panic(fmt.Sprintf("tipo inválido para a variável 'let' %s", node.Name.Value))
	}

	ptr := c.builder.CreateAlloca(valType, node.Name.Value)
	c.builder.CreateStore(val, ptr)
	c.logTrace(fmt.Sprintf("DEBUG: Alocando ponteiro para a variável: %v", ptr))

	entry := SymbolEntry{Ptr: ptr, Typ: valType, IsLiteral: false}

	switch valueNode := node.Value.(type) {
	case *ast.ArrayLiteral:
		elemCount := len(valueNode.Elements)
		elemType := c.context.Int32Type()
		entry.ArrayType = llvm.ArrayType(elemType, elemCount)
	case *ast.Identifier:
		if symbol, ok := c.getSymbol(valueNode.Value); ok {
			if !symbol.ArrayType.IsNil() {
				entry.ArrayType = symbol.ArrayType
			}
		}
	}
	c.setSymbol(node.Name.Value, entry)
}

// genConstStatement gera código para a declaração de constantes.
func (c *CodeGenerator) genConstStatement(node *ast.ConstStatement) {
	c.logTrace(fmt.Sprintf("Gerando declaração 'const' para a constante '%s'", node.Name.Value))
	val := c.genExpression(node.Value)
	isConst := !val.IsAConstant().IsNil()
	typ := c.GetValueTypeSafe(val)

	if isConst {
		c.logTrace(fmt.Sprintf("DEBUG: Constante '%s' é um literal. Armazenando valor diretamente.", node.Name.Value))
		c.setSymbol(node.Name.Value, SymbolEntry{Value: val, Typ: typ, IsLiteral: true})
	} else {
		c.logTrace(fmt.Sprintf("DEBUG: Constante '%s' é um resultado de instrução, tratando como variável imutável.", node.Name.Value))
		ptr := c.builder.CreateAlloca(typ, node.Name.Value)
		c.builder.CreateStore(val, ptr)
		c.setSymbol(node.Name.Value, SymbolEntry{Ptr: ptr, Typ: typ, IsLiteral: true})
	}
}

// genReturnStatement gera código para a instrução `return`.
func (c *CodeGenerator) genReturnStatement(node *ast.ReturnStatement) {
	c.logTrace("Gerando declaração 'return'")
	val := c.genExpression(node.ReturnValue)
	c.builder.CreateRet(val)
}

// genExpressionStatement gera código para uma declaração de expressão.
func (c *CodeGenerator) genExpressionStatement(node *ast.ExpressionStatement) {
	c.logTrace("Gerando declaração de expressão")
	c.genExpression(node.Expression)
}

// genBlockStatement gera código para um bloco de declarações.
func (c *CodeGenerator) genBlockStatement(node *ast.BlockStatement) {
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
}

// genFunctionDeclaration gera código para a declaração de funções.
func (c *CodeGenerator) genFunctionDeclaration(node *ast.FunctionDeclaration) {
	var retType llvm.Type
	if node.Name.Value == "main" {
		retType = c.context.Int32Type()
	} else {
		retType = c.context.Int32Type()
	}
	c.currentFunctionReturnType = retType
	c.logTrace(fmt.Sprintf("DEBUG: Tipo de retorno da função '%s': %v", node.Name.Value, retType))

	// CORREÇÃO: Revertido para que parâmetros sejam i32 por padrão.
	paramTypes := make([]llvm.Type, len(node.Parameters))
	for i := range node.Parameters {
		paramTypes[i] = c.context.Int32Type()
	}

	funcType := llvm.FunctionType(retType, paramTypes, false)
	function := llvm.AddFunction(c.module, node.Name.Value, funcType)
	c.setSymbol(node.Name.Value, SymbolEntry{Value: function, Typ: funcType, IsLiteral: true})

	if node.Body != nil {
		entryBlock := c.context.AddBasicBlock(function, "entry")
		c.builder.SetInsertPointAtEnd(entryBlock)
		c.pushScope()

		for i, param := range node.Parameters {
			paramValue := function.Param(i)
			paramValue.SetName(param.Value)
			alloca := c.builder.CreateAlloca(paramValue.Type(), param.Value)
			c.builder.CreateStore(paramValue, alloca)
			c.setSymbol(param.Value, SymbolEntry{Ptr: alloca, Typ: paramValue.Type(), IsLiteral: false})
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

// ... (resto do arquivo `statement.go` sem alterações) ...
func (c *CodeGenerator) genWhileStatement(node *ast.WhileStatement) {
	function := c.builder.GetInsertBlock().Parent()
	condBlock := c.context.AddBasicBlock(function, "loop_cond")
	loopBlock := c.context.AddBasicBlock(function, "loop_body")
	endBlock := c.context.AddBasicBlock(function, "loop_end")

	prevLoopCond := c.loopCondBlock
	prevLoopEnd := c.loopEndBlock
	c.loopCondBlock = condBlock
	c.loopEndBlock = endBlock

	c.builder.CreateBr(condBlock)
	c.builder.SetInsertPointAtEnd(condBlock)

	cond := c.genExpression(node.Condition)
	condType := c.GetValueTypeSafe(cond)
	if condType.TypeKind() != llvm.IntegerTypeKind || condType.IntTypeWidth() != 1 {
		panic(fmt.Sprintf("expressão condicional inválida no while, esperava i1, recebeu %v", condType))
	}

	c.builder.CreateCondBr(cond, loopBlock, endBlock)
	c.builder.SetInsertPointAtEnd(loopBlock)

	c.genStatement(node.Body)
	if !isBlockTerminated(c.builder.GetInsertBlock()) {
		c.builder.CreateBr(condBlock)
	}

	c.builder.SetInsertPointAtEnd(endBlock)
	c.loopCondBlock = prevLoopCond
	c.loopEndBlock = prevLoopEnd
}

func (c *CodeGenerator) genBreakStatement(node *ast.BreakStatement) {
	if c.loopEndBlock.IsNil() {
		panic("'break' fora de um loop")
	}
	c.builder.CreateBr(c.loopEndBlock)
}

func (c *CodeGenerator) genContinueStatement(node *ast.ContinueStatement) {
	if c.loopCondBlock.IsNil() {
		panic("'continue' fora de um loop")
	}
	c.builder.CreateBr(c.loopCondBlock)
}
