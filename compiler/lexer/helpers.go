// compiler/lexer/helpers.go
package lexer

import "taquion/compiler/token"

func isLetter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
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

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	if l.ch == '/' && l.peekChar() == '/' {
		// Log para o comentário encontrado
		l.logger.Println("Comentário '//' encontrado, pulando linha.")
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
	}
}

// logToken escreve as informações do token gerado no arquivo de log.
func (l *Lexer) logToken(tok token.Token) {
	// Não logar EOF para não poluir o final do log
	if tok.Type == token.EOF {
		return
	}
	l.logger.Printf("Token gerado -> Tipo: %-10s | Literal: '%s'", tok.Type, tok.Literal)
}
