// token.go — Define os tipos de token usados pelo lexer e parser da linguagem Taquion.
package token

// TokenType define o tipo de um token (ex: "INT", "RETURN", "+")
type TokenType string

// Token representa uma unidade léxica da linguagem (ex: let, 42, +)
type Token struct {
	Type    TokenType // Tipo do token (ex: LET, INT, IDENT, etc.)
	Literal string    // Texto original do token
}

// Tipos de tokens reconhecidos pela linguagem.
const (
	// Tokens especiais
	ILLEGAL = "ILEGAL" // Token inválido/desconhecido
	EOF     = "EOF"    // Fim de arquivo

	// Identificadores e literais
	IDENT = "IDENT" // Nome de variável, função, etc (ex: x, soma)
	INT   = "INT"   // Número inteiro (ex: 123)

	// Operadores
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	LT     = "<"
	GT     = ">"
	EQ     = "=="
	NOT_EQ = "!="

	// Delimitadores
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"

	// Palavras-chave
	FUNCTION = "FUNCTION" // func
	LET      = "LET"      // let
	RETURN   = "RETURN"   // return
	INT_TYPE = "INT_TYPE" // int (tipo)

	IF   = "IF"   // if
	ELSE = "ELSE" // else
)

// keywords mapeia strings para seus tipos de token, se forem palavras-chave
var keywords = map[string]TokenType{
	"func":   FUNCTION,
	"let":    LET,
	"return": RETURN,
	"int":    INT_TYPE,
	"if":     IF,
	"else":   ELSE,
}

// LookupIdent retorna o tipo de token para identificadores ou palavras-chave.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
