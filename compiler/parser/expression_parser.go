// Arquivo: parser/expression_parser.go
package parser

import (
	"fmt"
	"strconv"
	"taquion/compiler/ast"
	"taquion/compiler/token"
)

func (p *Parser) parseExpression(precedence int) ast.Expression {
	logger.Printf("        >> parseExpression (precedência: %d)", precedence)
	defer logger.Printf("        << parseExpression (precedência: %d)", precedence)

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	// Se o próximo token for um abre chaves, isso é um literal de struct
	if p.peekTokenIs(token.LBRACE) {
		typeName := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken() // consome o nome do tipo para que curToken seja '{'
		return p.parseCompositeLiteral(typeName)
	}

	// Caso contrário, é apenas um identificador normal
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("não foi possível analisar %q como inteiro", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{Token: p.curToken, Operator: p.curToken.Literal}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

// parseIfExpression analisa uma expressão 'if'.
// A sintaxe esperada é: if <condição> <consequência> [else <alternativa>]
func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	// Avança para o próximo token, que deve ser o início da condição.
	p.nextToken()

	// Analisa a expressão da condição.
	expression.Condition = p.parseExpression(LOWEST)

	// Após a condição, espera-se um abre chaves '{' para o bloco de consequência.
	if !p.expectPeek(token.LBRACE) {
		return nil // Retorna nulo se não encontrar o '{'
	}

	// Analisa o bloco de código da consequência.
	expression.Consequence = p.parseBlockStatement()

	// Verifica opcionalmente pela cláusula 'else'.
	// Se o próximo token for 'else', há um bloco de alternativa.
	if p.peekTokenIs(token.ELSE) {
		p.nextToken() // Consome o token 'else'

		// Espera-se um abre chaves '{' para o bloco de alternativa.
		if !p.expectPeek(token.LBRACE) {
			return nil // Retorna nulo se não encontrar o '{'
		}

		// Analisa o bloco de código da alternativa.
		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	lit.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	lit.Body = p.parseBlockStatement()
	return lit
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

func (p *Parser) parseAssignmentExpression(left ast.Expression) ast.Expression {
	expr := &ast.AssignmentExpression{Token: p.curToken, Left: left}
	p.nextToken()
	expr.Value = p.parseExpression(LOWEST)
	return expr
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) parseMemberExpression(left ast.Expression) ast.Expression {
	exp := &ast.MemberExpression{Token: p.curToken, Object: left}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	exp.Property = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
	return exp
}

// parseCompositeLiteral analisa um literal composto,
// aceitando nome:valor ou nome=valor, e vírgulas ou ponto‑e‑vírgula como separador.
func (p *Parser) parseCompositeLiteral(typeName *ast.Identifier) ast.Expression {
	lit := &ast.CompositeLiteral{
		Token:    p.curToken,
		TypeName: typeName,
		Fields:   []*ast.KeyValueExpr{},
	}

	// repete enquanto não fechar '}' ou chegar ao EOF
	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken() // avança para o IDENT do campo
		key := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		// aceita ':' ou '='
		if p.peekTokenIs(token.COLON) || p.peekTokenIs(token.ASSIGN) {
			p.nextToken() // consome ':' ou '='
		} else {
			p.errors = append(p.errors, fmt.Sprintf("esperava ':' ou '=', mas obteve %q", p.peekToken.Literal))
			return nil
		}

		p.nextToken() // avança para o valor
		value := p.parseExpression(LOWEST)

		lit.Fields = append(lit.Fields, &ast.KeyValueExpr{
			Key:   key,
			Value: value,
		})

		// aceita ',' ou ';' como separador opcional
		if p.peekTokenIs(token.COMMA) || p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return lit
}
