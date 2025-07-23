// O pacote lexer é responsável por pegar o código-fonte como uma string
// e transformá-lo em uma sequência de tokens.
package lexer

import "taquion/compiler/token"

// Lexer é a estrutura que representa nosso analisador léxico.
// Ele mantém o controle da posição atual no código para poder gerar tokens.
type Lexer struct {
	input        string // O código-fonte completo.
	position     int    // Posição atual no input (aponta para o caractere atual).
	readPosition int    // Próxima posição de leitura no input (depois do caractere atual).
	ch           byte   // O caractere atual que estamos examinando.
}

// New cria e retorna um novo Lexer pronto para ser usado.
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar() // Inicializa o lexer lendo o primeiro caractere.
	return l
}

// readChar lê o próximo caractere do input e avança as posições.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // 0 é o código ASCII para "NUL", que usaremos para significar Fim de Arquivo (EOF).
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

// NextToken é o coração do Lexer. Ele olha para o caractere atual (l.ch)
// e retorna o token correspondente.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace() // Ignora espaços em branco, tabulações e quebras de linha.

	switch l.ch {
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case 0: // Fim do arquivo
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			// Se for uma letra, pode ser uma palavra-chave ou um identificador.
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal) // Verifica se é 'func' ou 'return'.
			return tok                                // Retornamos aqui porque readIdentifier já avançou o cursor.
		} else if isDigit(l.ch) {
			// Se for um dígito, é um número inteiro.
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok // Retornamos aqui porque readNumber já avançou o cursor.
		} else {
			// Se não for nada conhecido, é um token ilegal.
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar() // Avança para o próximo caractere antes de retornar.
	return tok
}

// newToken é uma função auxiliar para criar um novo token.
func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

// skipWhitespace avança o cursor do lexer ignorando todos os espaços em branco.
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// readIdentifier lê uma sequência de letras e retorna a palavra completa.
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// isLetter verifica se um caractere é uma letra (simplificado para a-z, A-Z e _).
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// readNumber lê uma sequência de dígitos e retorna o número completo como string.
func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// isDigit verifica se um caractere é um dígito.
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
