package parser

import (
	"fmt"
	"taquion/compiler/token"
)

// --- Funções Auxiliares e de Erro ---

func (p *Parser) Errors() []string { return p.errors }

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
	p.logger.Printf("Avançando token: cur=%-10s ('%s') | peek=%-10s ('%s')\n", p.curToken.Type, p.curToken.Literal, p.peekToken.Type, p.peekToken.Literal)
}

func (p *Parser) curTokenIs(t token.TokenType) bool  { return p.curToken.Type == t }
func (p *Parser) peekTokenIs(t token.TokenType) bool { return p.peekToken.Type == t }

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("nenhuma função de parsing de prefixo encontrada para %s", t)
	p.errors = append(p.errors, msg)
	p.logTrace(msg)
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("esperava o próximo token ser %s, mas obteve %s (%q)",
		t, p.peekToken.Type, p.peekToken.Literal)
	p.errors = append(p.errors, msg)
	p.logTrace(msg)
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
		indent += "    " // 4 espaços para indentação
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
