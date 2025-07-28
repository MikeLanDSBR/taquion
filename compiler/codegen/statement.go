package codegen

import (
	"fmt"
	"taquion/compiler/ast"

	"github.com/taquion-lang/go-llvm"
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
	case *ast.TypeDeclaration:
		c.genTypeDeclaration(node)
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

	if compLit, ok := node.Value.(*ast.CompositeLiteral); ok {
		entry.TypeName = compLit.TypeName.Value
	}

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

// Add parameter naming inside the genConstructor function

func (c *CodeGenerator) genConstructor(node *ast.TypeDeclaration) {
	structName := node.Name.Value
	constructorName := structName + ".constructor"

	structType := c.getLLVMStructType(structName)
	ptrType := llvm.PointerType(structType, 0)

	paramTypes := []llvm.Type{ptrType}
	for _, field := range node.Fields {
		paramTypes = append(paramTypes, c.lookupLLVMType(field.Type))
	}

	fnType := llvm.FunctionType(c.context.VoidType(), paramTypes, false)
	constructor := llvm.AddFunction(c.module, constructorName, fnType)

	// ▼▼▼ ADD THIS BLOCK ▼▼▼
	// Name the parameters so they can be looked up later.
	constructor.Param(0).SetName("self_ptr")
	for i, field := range node.Fields {
		constructor.Param(i + 1).SetName(field.Name.Value)
	}
	// ▲▲▲ END OF BLOCK ▲▲▲

	entry := llvm.AddBasicBlock(constructor, "entry")
	c.builder.SetInsertPointAtEnd(entry)

	structPtr := constructor.Param(0)

	for i, field := range node.Fields {
		ptr := c.builder.CreateStructGEP(structType, structPtr, i, field.Name.Value+"_ptr")
		c.builder.CreateStore(constructor.Param(i+1), ptr)
	}

	c.builder.CreateRetVoid()
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

// In codegen/statement.go

func (c *CodeGenerator) genTypeDeclaration(node *ast.TypeDeclaration) {
	// ▼▼▼ ADD THIS LINE AT THE TOP ▼▼▼
	// This creates an opaque struct type first, allowing methods to
	// safely refer to the struct's own type.
	c.ensureStructType(node)
	// ▲▲▲ END OF CHANGE ▲▲▲

	// Now that the struct name is registered, you can generate the
	// constructor and methods.
	if len(node.Fields) > 0 {
		c.genConstructor(node)
	}

	for _, method := range node.Methods {
		// ... your existing method generation logic ...
		// This logic will now be able to find the 'Pessoa' type for the 'self' parameter.
		methodName := fmt.Sprintf("%s.%s", node.Name.Value, method.Name.Value)

		params := make([]*ast.Identifier, 0)
		selfType := &ast.Identifier{Value: node.Name.Value}
		selfParam := &ast.Identifier{Token: method.Token, Value: "self", Type: selfType}
		params = append(params, selfParam)

		for _, field := range node.Fields {
			fieldType, ok := field.Type.(*ast.Identifier)
			if !ok {
				panic(fmt.Sprintf("tipo de campo não-identificador não suportado em métodos: %s", field.Name.Value))
			}
			fieldParam := &ast.Identifier{Token: field.Name.Token, Value: field.Name.Value, Type: fieldType}
			params = append(params, fieldParam)
		}

		params = append(params, method.Parameters...)

		fnDecl := &ast.FunctionDeclaration{
			Token:      method.Token,
			Name:       &ast.Identifier{Token: method.Token, Value: methodName},
			Parameters: params,
			Body:       method.Body,
		}

		c.genFunctionDeclaration(fnDecl)
	}
}

// rewriteMemberExprs anda no AST de um BlockStatement,
// trocando self.X → Identifier("X").
func rewriteMemberExprs(bs *ast.BlockStatement) *ast.BlockStatement {
	stmts := make([]ast.Statement, 0, len(bs.Statements))
	for _, st := range bs.Statements {
		stmts = append(stmts, rewriteStmt(st))
	}
	return &ast.BlockStatement{Token: bs.Token, Statements: stmts}
}

func rewriteStmt(stmt ast.Statement) ast.Statement {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		return &ast.ExpressionStatement{
			Token:      s.Token,
			Expression: rewriteExpr(s.Expression),
		}
	case *ast.ReturnStatement:
		return &ast.ReturnStatement{
			Token:       s.Token,
			ReturnValue: rewriteExpr(s.ReturnValue),
		}
	case *ast.BlockStatement:
		return rewriteMemberExprs(s)
	// … você pode adicionar outros casos, se for suportar break/if/while/etc.
	default:
		return stmt
	}
}

func rewriteExpr(expr ast.Expression) ast.Expression {
	switch e := expr.(type) {
	case *ast.MemberExpression:
		// se for self.<campo>, troca
		if obj, ok := e.Object.(*ast.Identifier); ok && obj.Value == "self" {
			return &ast.Identifier{Token: e.Property.Token, Value: e.Property.Value}
		}
		// senão, recursão no objeto e propriedade
		return &ast.MemberExpression{
			Object:   rewriteExpr(e.Object),
			Property: e.Property,
		}
	case *ast.InfixExpression:
		return &ast.InfixExpression{
			Token:    e.Token,
			Left:     rewriteExpr(e.Left),
			Operator: e.Operator,
			Right:    rewriteExpr(e.Right),
		}
	// trate também PrefixExpression, CallExpression, etc, conforme sua linguagem
	default:
		return expr
	}
}

func (c *CodeGenerator) genMemberExpression(node *ast.MemberExpression) llvm.Value {
	c.logTrace(fmt.Sprintf(
		"DEBUG: Acessando campo '%s' do objeto '%s'",
		node.Property.Value,
		node.Object.String(),
	))

	// 1. Obtém o ponteiro para o objeto struct (ex: 'self')
	objectIdent, ok := node.Object.(*ast.Identifier)
	if !ok {
		// Por enquanto, só suportamos acesso a membros de identificadores diretos.
		panic("acesso a membro em um não-identificador ainda não é suportado")
	}
	entry, ok := c.getSymbol(objectIdent.Value)
	if !ok {
		panic(fmt.Sprintf("objeto desconhecido: %s", objectIdent.Value))
	}

	// O 'Ptr' na tabela de símbolos é o ponteiro para a nossa struct alocada.
	objectPtr := entry.Ptr

	// 2. Get the struct type and find the field's numerical index.
	structType := objectPtr.Type().ElementType()
	structName := entry.TypeName // Use the TypeName from the symbol table

	if structName == "" {
		panic(fmt.Sprintf("não foi possível determinar o nome do tipo para o objeto '%s'", objectIdent.Value))
	}

	fieldIndex, ok := c.structFieldIndices[structName][node.Property.Value]
	if !ok {
		panic(fmt.Sprintf("campo '%s' não encontrado no tipo '%s'", node.Property.Value, structName))
	}

	// 3. Usa CreateStructGEP para obter um ponteiro para o campo específico.
	elementPtr := c.builder.CreateStructGEP(structType, objectPtr, fieldIndex, node.Property.Value+"_ptr")

	// 4. Carrega o valor do endereço de memória do campo.
	fieldType := structType.StructElementTypes()[fieldIndex]
	return c.builder.CreateLoad(fieldType, elementPtr, node.Property.Value+"_val")
}

// genCompositeLiteral gera o valor de um literal composto.
// Ex: Pessoa { nome: "Carlos", idade: 30 }
// Replace the entire genCompositeLiteral function with this new version

// genCompositeLiteral gera o valor de um literal composto, lidando com qualquer ordem de campos.
func (c *CodeGenerator) genCompositeLiteral(lit *ast.CompositeLiteral) llvm.Value {
	typeName := lit.TypeName.Value
	structType := c.getLLVMStructType(typeName)
	ctorName := fmt.Sprintf("%s.constructor", typeName)

	fn := c.module.NamedFunction(ctorName)
	if fn.IsNil() {
		panic(fmt.Sprintf("construtor não encontrado: %s", ctorName))
	}
	fnParams := fn.Params()

	// 1. Create a map of the provided literal fields for easy, order-independent lookup.
	literalFields := make(map[string]ast.Expression)
	for _, kv := range lit.Fields {
		// kv is *ast.KeyValueExpr, and kv.Key is already *ast.Identifier.
		// No type assertion is needed.
		literalFields[kv.Key.Value] = kv.Value
	}

	// 2. Build arguments in the correct order, as defined by the constructor's parameters.
	alloca := c.builder.CreateAlloca(structType, "tmp_struct")
	args := []llvm.Value{alloca}

	// The first param is the pointer to the struct itself, so we iterate from the second param.
	for i := 1; i < len(fnParams); i++ {
		param := fnParams[i]
		fieldName := param.Name()

		valueExpr, ok := literalFields[fieldName]
		if !ok {
			panic(fmt.Sprintf("campo obrigatório '%s' ausente no literal do tipo '%s'", fieldName, typeName))
		}

		value := c.genExpression(valueExpr)

		// Perform type casting/truncation, just like in the previous fix.
		expectedType := param.Type()
		actualType := value.Type()
		if actualType != expectedType {
			if actualType.TypeKind() == llvm.IntegerTypeKind && expectedType.TypeKind() == llvm.IntegerTypeKind {
				value = c.builder.CreateTrunc(value, expectedType, fieldName+"_trunc")
			} else {
				panic(fmt.Sprintf(
					"tipo incompatível para o campo '%s': esperado %s, recebido %s",
					fieldName, expectedType.String(), actualType.String(),
				))
			}
		}
		args = append(args, value)
		delete(literalFields, fieldName) // Remove field to detect extras later.
	}

	// 3. Ensure no extra, unknown fields were provided in the literal.
	if len(literalFields) > 0 {
		for extraField := range literalFields {
			panic(fmt.Sprintf("campo desconhecido '%s' no literal para o tipo '%s'", extraField, typeName))
		}
	}

	// 4. Call the constructor and load the result.
	fnTy := fn.Type().ElementType()
	c.builder.CreateCall(fnTy, fn, args, "")

	return c.builder.CreateLoad(structType, alloca, "result_struct")
}
