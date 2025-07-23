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

	// --- CAMPOS NOVOS PARA LOGGING ---
	logger  *log.Logger
	LogFile *os.File
}

func New(input string) *Lexer {
	// --- SETUP DO LOGGING ---
	file, err := os.OpenFile("lexer.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
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

func (l *Lexer) logToken(tok token.Token) {
	l.logger.Printf("Token gerado -> Tipo: %-10s | Literal: '%s'\n", tok.Type, tok.Literal)
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case '/':
		if l.peekChar() == '/' {
			l.logger.Println("Comentário '//' encontrado, pulando linha.")
			l.skipComment()
			return l.NextToken()
		}
		tok = newToken(token.SLASH, l.ch)
	case '=':
		tok = newToken(token.ASSIGN, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	// ... (outros casos simples)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
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
			l.logToken(tok) // Log antes de retornar
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			l.logToken(tok) // Log antes de retornar
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	l.logToken(tok) // Log antes de retornar
	return tok
}

// ... (resto do lexer.go sem alterações: peekChar, readChar, skipComment, etc.)
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}
