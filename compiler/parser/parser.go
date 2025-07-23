// ARQUIVO COMPLETO E CORRIGIDO: compiler/parser/parser.go

package parser

import (
	"fmt"
	"taquion/compiler/ast"
	"taquion/compiler/lexer"
	"taquion/compiler/token"
)

type Parser struct {
	l         *lexer.Lexer
	errors    []string
	curToken  token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	// Lê dois tokens para que tanto curToken quanto peekToken sejam preenchidos
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// ParseProgram é o ponto de entrada e o loop principal
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// O loop continua até o final do arquivo
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		// ESSA LINHA É CRUCIAL:
		// Após uma declaração ser analisada (e parar no ';'),
		// nós avançamos para o token seguinte para começar a próxima.
		p.nextToken()
	}
	return program
}

// parseStatement decide qual tipo de declaração analisar
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return nil
	}
}

// parseReturnStatement analisa UMA ÚNICA declaração de retorno
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken} // O token atual é 'return'

	p.nextToken() // Avança para a expressão (ex: '5')

	// Este loop pula a expressão até encontrar o ponto e vírgula.
	// No futuro, aqui você analisará a expressão em vez de pular.
	for !p.curTokenIs(token.SEMICOLON) && !p.curTokenIs(token.EOF) {
		p.nextToken()
	}

	// A função termina aqui. O token atual é ';'.
	// NÃO chamamos nextToken() aqui. Deixamos essa tarefa para o ParseProgram.
	return stmt
}

// --- Funções Auxiliares ---

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
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("esperado que o próximo token fosse %s, mas recebido %s",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}
