// compiler/lexer/lexer.go
package lexer

import (
	"log"
	"os"
	"taquion/compiler/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte

	// Campos para logging
	logger  *log.Logger
	LogFile *os.File
}

func New(input string) *Lexer {
	file, err := os.OpenFile("log/lexer.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo de log do lexer: %v", err)
	}

	l := &Lexer{
		input:   input,
		logger:  log.New(file, "LEXER:  ", log.LstdFlags),
		LogFile: file,
	}
	l.logger.Println("Iniciando nova sessão de lexing.")
	l.readChar()
	return l
}

// NextToken analisa a entrada e retorna o próximo token.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	// --- NOVO CASE PARA STRINGS ---
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	// ------------------------------
	case '/':
		if l.peekChar() == '/' {
			l.logger.Println("Comentário '//' encontrado, pulando linha.")
			l.skipComment()
			return l.NextToken()
		}
		tok = newToken(token.SLASH, l.ch)
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '>':
		tok = newToken(token.GT, l.ch)
	case '<':
		tok = newToken(token.LT, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			l.logToken(tok)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			l.logToken(tok)
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	l.logToken(tok)
	return tok
}
