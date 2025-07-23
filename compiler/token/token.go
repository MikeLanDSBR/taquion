// O pacote token define as constantes e estruturas que representam as unidades
// léxicas da linguagem Taquion, como palavras-chave, operadores e identificadores.
package token

// TokenType é o tipo que representa a categoria de um token (ex: "INT", "RETURN").
// Usar uma string facilita a depuração.
type TokenType string

// Token representa uma unidade léxica individual identificada pelo Lexer.
type Token struct {
	Type    TokenType // O tipo do token, ex: token.RETURN.
	Literal string    // O valor literal do token, ex: "return".
}

// Definição de todos os tipos de tokens da linguagem.
const (
	// Especiais
	ILLEGAL = "ILEGAL" // Representa um token que não é reconhecido pela linguagem.
	EOF     = "EOF"    // Representa o fim do arquivo (End Of File).

	// Identificadores e Literais
	IDENT = "IDENT" // Nomes de variáveis e funções: main, x, minhaFuncao.
	INT   = "INT"   // Literais de números inteiros: 123, 42.

	// Operadores
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	// Delimitadores
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"

	// Palavras-chave
	FUNCTION = "FUNCTION"
	RETURN   = "RETURN"
	LET      = "LET"
)

// keywords é um mapa que associa as strings das palavras-chave aos seus
// respectivos tipos de token.
var keywords = map[string]TokenType{
	"func":   FUNCTION,
	"return": RETURN,
	"let":    LET,
}

// LookupIdent verifica se um identificador é uma palavra-chave reservada.
// Se for, retorna o TokenType da palavra-chave; senão, retorna token.IDENT.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
