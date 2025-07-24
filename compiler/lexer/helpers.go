// compiler/lexer/helpers.go
package lexer

import "taquion/compiler/token"

// --- NOVA FUNÇÃO PARA LER STRINGS ---
func (l *Lexer) readString() string {
	// A posição inicial é depois do '"' de abertura
	position := l.position + 1
	for {
		l.readChar()
		// Continua lendo até encontrar o '"' de fechamento ou o fim do arquivo
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	// Retorna a fatia da string de entrada que está entre as aspas
	return l.input[position:l.position]
}

// --- Funções Auxiliares ---

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // 0 é o código ASCII para "NUL", representa o fim do arquivo (EOF)
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) logToken(tok token.Token) {
	// Não logar EOF para não poluir o final do log
	if tok.Type == token.EOF {
		return
	}
	l.logger.Printf("Token gerado -> Tipo: %-10s | Literal: '%s'\n", tok.Type, tok.Literal)
}
