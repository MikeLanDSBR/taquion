package parser

import (
	"taquion/compiler/ast"
	"taquion/compiler/token"
)

// parseStatement agora decide qual tipo de declaração analisar.
func (p *Parser) parseStatement() ast.Statement {
	defer p.traceOut("parseStatement")
	p.traceIn("parseStatement")

	switch p.curToken.Type {
	// CORREÇÃO 1: Adicionada a regra para analisar 'package'.
	// Isso resolve o erro: "nenhuma função de parsing de prefixo encontrada para PACKAGE"
	case token.PACKAGE:
		return p.parsePackageStatement()
	case token.CONST:
		return p.parseConstStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.FUNCTION:
		return p.parseFunctionDeclaration()
	default:
		if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.ASSIGN) {
			return p.parseAssignmentStatement()
		}
		return p.parseExpressionStatement()
	}
}

// --- NOVAS FUNÇÕES DE PARSING ---

// parsePackageStatement analisa a declaração 'package <nome>;'
func (p *Parser) parsePackageStatement() *ast.PackageStatement {
	stmt := &ast.PackageStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Consome o ponto e vírgula opcional no final da linha.
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// --- FUNÇÕES MODIFICADAS ---

// parseFunctionDeclaration foi corrigida para não esperar um tipo de retorno.
func (p *Parser) parseFunctionDeclaration() ast.Statement {
	decl := &ast.FunctionDeclaration{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	decl.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	decl.Parameters = p.parseFunctionParameters()

	// CORREÇÃO 2: A lógica que procurava por um tipo de retorno foi removida.
	// Agora, o parser espera um '{' logo após os parênteses.
	// Isso resolve o erro: "esperava o próximo token ser {, mas obteve RETURN"
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	decl.Body = p.parseBlockStatement()

	return decl
}

// parseBlockStatement foi corrigido para não consumir um token a mais.
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken() // Pula o token '{'

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken() // Avança para o próximo token DENTRO do bloco.
	}

	// CORREÇÃO 3: O `p.nextToken()` extra no final do loop foi removido.
	// O loop agora termina corretamente quando encontra '}'.
	// Isso resolve o erro: "nenhuma função de parsing de prefixo encontrada para }"

	return block
}

// parseFunctionParameters foi simplificado para a nova sintaxe sem tipos.
func (p *Parser) parseFunctionParameters() []*ast.Parameter {
	params := []*ast.Parameter{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return params
	}

	p.nextToken()

	param := &ast.Parameter{
		Token: p.curToken,
		Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}
	params = append(params, param)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		param := &ast.Parameter{
			Token: p.curToken,
			Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
		}
		params = append(params, param)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return params
}

// (O restante das suas funções de parsing de statement, como parseConstStatement, etc., permanecem as mesmas)
func (p *Parser) parseConstStatement() *ast.ConstStatement {
	stmt := &ast.ConstStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseAssignmentStatement() *ast.AssignmentStatement {
	stmt := &ast.AssignmentStatement{
		Name: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}
	p.nextToken()
	stmt.Token = p.curToken
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}
