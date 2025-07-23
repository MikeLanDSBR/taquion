// O pacote parser constrói a Árvore Sintática Abstrata (AST).
package parser

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"taquion/compiler/ast"
	"taquion/compiler/lexer"
	"taquion/compiler/token"
)

const (
	_ int = iota
	LOWEST
	// Futuramente: EQUALS, LESSGREATER, SUM, PRODUCT, etc.
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l         *lexer.Lexer
	errors    []string
	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn

	logger           *log.Logger
	LogFile          *os.File
	indentationLevel int
}

func New(l *lexer.Lexer) *Parser {
	file, err := os.OpenFile("log/parser.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo de log do parser: %v", err)
	}

	p := &Parser{
		l:                l,
		errors:           []string{},
		logger:           log.New(file, "PARSER: ", log.LstdFlags),
		LogFile:          file,
		indentationLevel: 0,
	}
	p.logger.Println("Iniciando nova sessão de parsing.")

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	// Registra as funções que sabem como analisar o início de uma expressão.
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)

	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) ParseProgram() *ast.Program {
	defer p.traceOut("ParseProgram")
	p.traceIn("ParseProgram")

	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	defer p.traceOut("parseStatement")
	p.traceIn("parseStatement")

	switch p.curToken.Type {
	case token.LET: // CORREÇÃO: Adicionado case para 'let'
		p.logTrace("Encontrado token LET, chamando parseLetStatement")
		return p.parseLetStatement()
	case token.FUNCTION:
		p.logTrace("Encontrado token FUNCTION, chamando parseFunctionDeclaration")
		return p.parseFunctionDeclaration()
	case token.RETURN:
		p.logTrace("Encontrado token RETURN, chamando parseReturnStatement")
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseLetStatement analisa: `let <ident> = <expressão>;`
func (p *Parser) parseLetStatement() *ast.LetStatement {
	defer p.traceOut("parseLetStatement")
	p.traceIn("parseLetStatement")

	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken() // Pula o '='

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseFunctionDeclaration() *ast.FunctionDeclaration {
	defer p.traceOut("parseFunctionDeclaration")
	p.traceIn("parseFunctionDeclaration")

	decl := &ast.FunctionDeclaration{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	decl.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	// TODO: Parse parameters
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	if !p.expectPeek(token.IDENT) {
		return nil
	} // Return type
	decl.ReturnType = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	decl.Body = p.parseBlockStatement()

	return decl
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	defer p.traceOut("parseBlockStatement")
	p.traceIn("parseBlockStatement")

	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}
	p.nextToken() // Pula o '{'

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	defer p.traceOut("parseReturnStatement")
	p.traceIn("parseReturnStatement")

	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	defer p.traceOut("parseExpressionStatement")
	p.traceIn("parseExpressionStatement")

	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	defer p.traceOut("parseExpression")
	p.traceIn(fmt.Sprintf("parseExpression (precedência: %d)", precedence))

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.logTrace(fmt.Sprintf("Nenhuma função de parsing de prefixo para o token %s", p.curToken.Type))
		return nil
	}
	leftExp := prefix()

	return leftExp
}

// parseIdentifier analisa um nome de variável como uma expressão.
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	defer p.traceOut("parseIntegerLiteral")
	p.traceIn("parseIntegerLiteral")

	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.logTrace(fmt.Sprintf("ERRO: não foi possível converter '%s' para inteiro", p.curToken.Literal))
		return nil
	}
	p.logTrace(fmt.Sprintf("Literal de inteiro analisado, valor: %d", value))

	lit.Value = value
	return lit
}

// --- Funções Auxiliares ---

func (p *Parser) Errors() []string { return p.errors }

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
	p.logger.Printf("Avançando token: cur=%-10s ('%s') | peek=%-10s ('%s')\n", p.curToken.Type, p.curToken.Literal, p.peekToken.Type, p.peekToken.Literal)
}

func (p *Parser) curTokenIs(t token.TokenType) bool { return p.curToken.Type == t }

func (p *Parser) peekTokenIs(t token.TokenType) bool { return p.peekToken.Type == t }

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	return false
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// --- Funções de Logging ---

func (p *Parser) logTrace(msg string) {
	indent := ""
	for i := 0; i < p.indentationLevel; i++ {
		indent += "  "
	}
	p.logger.Printf("%s%s\n", indent, msg)
}

func (p *Parser) traceIn(funcName string) {
	p.logTrace(">> " + funcName)
	p.indentationLevel++
}

func (p *Parser) traceOut(funcName string) {
	p.indentationLevel--
	p.logTrace("<< " + funcName)
}
