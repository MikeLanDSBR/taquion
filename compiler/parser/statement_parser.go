// Arquivo: parser/statement_parser.go
package parser

import (
	"taquion/compiler/ast"
	"taquion/compiler/token"
)

func (p *Parser) parseStatement() ast.Statement {
	logger.Println("    >> parseStatement")
	defer logger.Println("    << parseStatement")

	switch p.curToken.Type {
	case token.CONST:
		return p.parseConstStatement()
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.PACKAGE:
		return p.parsePackageStatement()
	case token.FUNCTION:
		return p.parseFunctionDeclaration()
	case token.TYPE:
		return p.parseTypeDeclaration()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.BREAK:
		return p.parseBreakStatement()
	case token.CONTINUE:
		return p.parseContinueStatement()
	default:
		return p.parseExpressionStatement()
	}
}

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

// parseBlockStatement analisa um bloco de código delimitado por chaves '{}'.
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	// Avança o token para o interior do bloco.
	p.nextToken()

	// Continua analisando statements até encontrar um fecha chaves '}' ou o fim do arquivo.
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
			continue
		}
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parsePackageStatement() *ast.PackageStatement {
	stmt := &ast.PackageStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseFunctionDeclaration() *ast.FunctionDeclaration {
	decl := &ast.FunctionDeclaration{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	decl.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	decl.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	decl.Body = p.parseBlockStatement()
	return decl
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	stmt := &ast.BreakStatement{Token: p.curToken}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	stmt := &ast.ContinueStatement{Token: p.curToken}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseTypeDeclaration() *ast.TypeDeclaration {
	// Cria o nó da declaração de tipo. Token atual é 'type'.
	stmt := &ast.TypeDeclaration{Token: p.curToken}

	// Espera o nome do tipo (ex: Pessoa)
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Espera o abre chaves '{' que inicia o corpo do tipo
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Fields = []*ast.StructField{}
	stmt.Methods = []*ast.FunctionLiteral{}

	// Loop para analisar o corpo do tipo (campos e métodos)
	// O loop continua enquanto não encontrarmos a chave de fechamento '}'
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {

		// CASO 1: É um método
		if p.curTokenIs(token.FUNCTION) {
			method := &ast.FunctionLiteral{Token: p.curToken}

			if !p.expectPeek(token.IDENT) { // Nome do método
				return nil
			}
			method.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

			if !p.expectPeek(token.LPAREN) { // Parâmetros
				return nil
			}
			method.Parameters = p.parseFunctionParameters()

			// Suporte para tipo de retorno opcional (ex: func saudacao() string {...})
			// Se não for uma chave, deve ser o tipo de retorno.
			if !p.peekTokenIs(token.LBRACE) {
				p.nextToken() // Consome o token do tipo (ex: 'string')
				// Você pode querer criar um nó AST para o tipo de retorno aqui
			}

			if !p.expectPeek(token.LBRACE) { // Corpo do método
				return nil
			}
			method.Body = p.parseBlockStatement() // Esta função termina com curToken em '}'
			stmt.Methods = append(stmt.Methods, method)

			// CASO 2: É um campo
		} else if p.curTokenIs(token.IDENT) && p.peekTokenIs(token.COLON) {
			field := &ast.StructField{}
			field.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

			p.nextToken() // Consome o nome do campo
			p.nextToken() // Consome o ':'

			field.Type = p.parseExpression(LOWEST) // Analisa a expressão do tipo
			stmt.Fields = append(stmt.Fields, field)

			// CASO 3: Token inesperado
		} else {
			// Pula tokens inesperados para evitar loop infinito
			p.nextToken()
			continue
		}

		// Após analisar um campo ou método, avança para o próximo token
		// para iniciar a próxima iteração do loop.
		p.nextToken()
	}

	// Garante que a declaração de tipo foi fechada corretamente com '}'
	if !p.curTokenIs(token.RBRACE) {
		p.errors = append(p.errors, "esperava '}' para fechar a declaração de tipo")
		return nil
	}

	return stmt
}
