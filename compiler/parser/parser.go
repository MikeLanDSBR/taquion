// O pacote parser é responsável por receber uma sequência de tokens do lexer
// e construir uma Árvore Sintática Abstrata (AST) que representa a estrutura
// do código-fonte.
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

// Define a precedência dos operadores para o analisador Pratt.
const (
	_ int = iota
	LOWEST
	// Futuramente: EQUALS, LESSGREATER, SUM, PRODUCT, PREFIX, CALL
)

// Declara os tipos para as funções de parsing de prefixo e infixo.
type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Parser contém o estado necessário para analisar o código Taquion.
type Parser struct {
	l         *lexer.Lexer
	errors    []string
	curToken  token.Token
	peekToken token.Token

	// Mapas para associar tipos de token às suas funções de parsing.
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn

	// Campos para logging.
	logger           *log.Logger
	LogFile          *os.File // Exportado para ser fechado pelo main.
	indentationLevel int
}

// New cria um novo Parser e inicializa o sistema de logging.
func New(l *lexer.Lexer) *Parser {
	// Abre (ou cria/trunca) o arquivo de log do parser.
	file, err := os.OpenFile("parser.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
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

	// Registra as funções de parsing de prefixo.
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)

	p.nextToken()
	p.nextToken()
	return p
}

// ParseProgram é o ponto de entrada que analisa todo o programa.
func (p *Parser) ParseProgram() *ast.Program {
	defer p.traceOut("ParseProgram")
	p.traceIn("ParseProgram")

	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		p.logTrace(fmt.Sprintf("Analisando token: %s ('%s')", p.curToken.Type, p.curToken.Literal))
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

// parseStatement decide qual tipo de declaração analisar com base no token atual.
func (p *Parser) parseStatement() ast.Statement {
	defer p.traceOut("parseStatement")
	p.traceIn("parseStatement")

	switch p.curToken.Type {
	case token.FUNCTION: // CORREÇÃO: Adicionado case para 'func'
		p.logTrace("Encontrado token FUNCTION, chamando parseFunctionDeclaration")
		return p.parseFunctionDeclaration()
	case token.RETURN:
		p.logTrace("Encontrado token RETURN, chamando parseReturnStatement")
		return p.parseReturnStatement()
	default:
		// Se não for uma palavra-chave de declaração, pode ser uma expressão.
		p.logTrace("Nenhuma declaração conhecida, tentando analisar como ExpressionStatement.")
		return p.parseExpressionStatement()
	}
}

// parseFunctionDeclaration analisa uma declaração de função: `func <nome>() <tipo> { ... }`
// NOTA: Esta é uma versão MUITO simplificada para o seu caso atual.
func (p *Parser) parseFunctionDeclaration() *ast.FunctionDeclaration {
	defer p.traceOut("parseFunctionDeclaration")
	p.traceIn("parseFunctionDeclaration")

	// Cria o nó da declaração de função com o token `func`.
	decl := &ast.FunctionDeclaration{Token: p.curToken}

	// Espera e consome o nome da função (ex: 'main').
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	decl.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Espera e consome o parêntese de abertura '('.
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	// Futuramente, aqui viria a análise de parâmetros.
	// Por agora, apenas esperamos o parêntese de fechamento ')'.
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// Espera e consome o tipo de retorno (ex: 'int').
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	decl.ReturnType = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Espera e consome a chave de abertura '{'.
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Analisa o corpo da função.
	decl.Body = p.parseBlockStatement()

	return decl
}

// parseBlockStatement analisa um bloco de código contido entre `{` e `}`.
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	defer p.traceOut("parseBlockStatement")
	p.traceIn("parseBlockStatement")

	block := &ast.BlockStatement{Token: p.curToken} // O token '{'.
	block.Statements = []ast.Statement{}

	p.nextToken() // Pula o '{'.

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

	// Consome o ponto e vírgula opcional no final da linha.
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

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
	p.logger.Printf("Avançando token: cur=%-10s ('%s') | peek=%-10s ('%s')\n", p.curToken.Type, p.curToken.Literal, p.peekToken.Type, p.peekToken.Literal)
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	// Adicionaremos tratamento de erro aqui no futuro.
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
