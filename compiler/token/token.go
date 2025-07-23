// O pacote 'token' define as constantes, tipos e estruturas
// que representam as unidades léxicas da linguagem Takion.
package token

// TokenType é uma string que representa o tipo de um token.
// Usar uma string torna a depuração mais fácil no início.
type TokenType string

// Token é a estrutura que representa um token individual.
// Ele tem um tipo (ex: NÚMERO) e um literal (o valor real, ex: "42").
type Token struct {
	Type    TokenType
	Literal string
}

// Constantes para todos os tipos de token da linguagem
const (
	// Tokens especiais
	ILLEGAL = "ILEGAL" // Token/caractere desconhecido
	EOF     = "EOF"    // End of File (Fim de Arquivo)

	// Identificadores + Literais
	IDENT = "IDENT" // main, x, minhaFuncao
	INT   = "INT"   // 123, 42

	// Operadores (vamos adicionar mais depois)
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	// Delimitadores
	COMMA     = ","
	SEMICOLON = ";"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	// Palavras-chave
	FUNCTION = "FUNCTION"
	RETURN   = "RETURN"
	// Adicionaremos 'let', 'true', 'false', 'if', 'else' etc. no futuro
)
