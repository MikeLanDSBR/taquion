// O pacote lexer transforma o código-fonte em uma sequência de tokens.
package lexer

import "taquion/compiler/token"

// Lexer é o analisador léxico. Ele percorre o input e gera tokens.
type Lexer struct {
	input        string // O código-fonte a ser analisado.
	position     int    // Posição atual no input (aponta para l.ch).
	readPosition int    // Próxima posição a ser lida.
	ch           byte   // Caractere atual sob exame.
}

// New cria e retorna um novo Lexer.
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar() // Lê o primeiro caractere para inicializar o estado.
	return l
}

// readChar lê o próximo caractere do input e avança os ponteiros do lexer.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // 0 é o código ASCII para "NUL", representando o fim do arquivo (EOF).
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// NextToken analisa o caractere atual e retorna o token correspondente.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace() // Pula espaços em branco, quebras de linha, etc.

	switch l.ch {
	// Operadores
	case '=':
		tok = newToken(token.ASSIGN, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '!':
		tok = newToken(token.BANG, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '/':
		tok = newToken(token.SLASH, l.ch)
	// Delimitadores
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	// Fim do arquivo
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok // Retorna aqui pois readIdentifier já avançou o cursor.
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok // Retorna aqui pois readNumber já avançou o cursor.
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar() // Avança para o próximo caractere.
	return tok
}

// newToken cria um token a partir de um tipo e um caractere.
func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

// skipWhitespace consome todos os caracteres de espaço em branco consecutivos.
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// readIdentifier lê um identificador (letras e '_') e avança o cursor.
func (l *Lexer) readIdentifier() string {
	startPosition := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[startPosition:l.position]
}

// isLetter verifica se o caractere é uma letra ou underscore.
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// readNumber lê um número inteiro e avança o cursor.
func (l *Lexer) readNumber() string {
	startPosition := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[startPosition:l.position]
}

// isDigit verifica se o caractere é um número.
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
