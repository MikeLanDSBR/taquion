// parser/statement_parser.go
package parser

import (
	"taquion/compiler/ast"
	"taquion/compiler/token"
)

// parseStatement analisa uma única declaração.
func (p *Parser) parseStatement() ast.Statement {
	defer p.traceOut("parseStatement")
	p.traceIn("parseStatement")

	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.FUNCTION:
		return p.parseFunctionDeclaration()
	default:
		// --- LÓGICA DE DECISÃO PARA REATRIBUIÇÃO ---
		// Se encontrarmos um identificador e o próximo token for '=',
		// então é uma declaração de reatribuição.
		if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.ASSIGN) {
			return p.parseAssignmentStatement()
		}
		// Caso contrário, é uma declaração de expressão normal.
		return p.parseExpressionStatement()
	}
}

// --- NOVA FUNÇÃO PARA PARSE DE REATRIBUIÇÃO ---
func (p *Parser) parseAssignmentStatement() *ast.AssignmentStatement {
	defer p.traceOut("parseAssignmentStatement")
	p.traceIn("parseAssignmentStatement")

	stmt := &ast.AssignmentStatement{
		Name: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}

	p.nextToken() // Avança do identificador para o token '='
	stmt.Token = p.curToken

	p.nextToken() // Avança do '=' para o início da expressão
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// Funções parseLetStatement, parseReturnStatement, etc. (sem alterações)
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}
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

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

// --- FUNÇÃO MODIFICADA ---
func (p *Parser) parseFunctionDeclaration() ast.Statement {
	decl := &ast.FunctionDeclaration{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	decl.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// Chama a nova função auxiliar para analisar os parâmetros.
	decl.Parameters = p.parseFunctionParameters()

	// Analisa o tipo de retorno.
	if !p.peekTokenIs(token.RBRACE) { // Se não for uma função sem retorno
		p.nextToken()
		decl.ReturnType = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	decl.Body = p.parseBlockStatement()

	return decl
}

// --- NOVA FUNÇÃO AUXILIAR ---
func (p *Parser) parseFunctionParameters() []*ast.Parameter {
	params := []*ast.Parameter{}

	// Se o próximo token for ')', não há parâmetros.
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return params
	}

	p.nextToken() // Avança para o nome do primeiro parâmetro.

	param := &ast.Parameter{Token: p.curToken}
	param.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken() // Avança para o tipo do parâmetro.
	param.Type = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	params = append(params, param)

	// Continua analisando outros parâmetros enquanto encontrar vírgulas.
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // Pula a vírgula
		p.nextToken() // Avança para o nome do próximo parâmetro

		param := &ast.Parameter{Token: p.curToken}
		param.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		p.nextToken() // Avança para o tipo do parâmetro.
		param.Type = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		params = append(params, param)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return params
}
