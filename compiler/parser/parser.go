// Arquivo: parser/parser.go
package parser

import (
	"log"
	"os"
	"sync"
	"taquion/compiler/ast"
	"taquion/compiler/lexer"
	"taquion/compiler/token"
)

var (
	logger   *log.Logger
	logFile  *os.File
	initOnce sync.Once
)

func initLogger() {
	initOnce.Do(func() {
		if err := os.MkdirAll("log", 0755); err != nil {
			log.Fatalf("Erro ao criar diretório de log: %v", err)
		}
		var err error
		logFile, err = os.OpenFile("log/parser.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("Erro ao abrir o arquivo de log parser.log: %v", err)
		}
		logger = log.New(logFile, "PARSER: ", log.LstdFlags)
		logger.Println("Iniciando nova sessão de parsing.")
	})
}

func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// Níveis de Precedência dos Operadores
const (
	_ int = iota
	LOWEST
	ASSIGN      // =
	EQUALS      // ==
	LESSGREATER // > ou <
	SUM         // +
	PRODUCT     // * ou / ou %  <-- Adicionei aqui
	PREFIX      // -X ou !X
	CALL        // minhaFuncao(X)
	INDEX       // array[index]
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.MODULO:   PRODUCT, // Adicionei a precedência para o novo token
	token.LPAREN:   CALL,
	token.ASSIGN:   ASSIGN,
	token.LBRACKET: INDEX,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	initLogger()
	p := &Parser{l: l, errors: []string{}}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.MODULO, p.parseInfixExpression) // Adicionei o registro para o novo token
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) ParseProgram() *ast.Program {
	logger.Println(">> ParseProgram")
	defer logger.Println("<< ParseProgram")

	program := &ast.Program{Statements: []ast.Statement{}}
	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}
