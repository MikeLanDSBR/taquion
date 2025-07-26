package token

import (
	"log"
	"os"
	"sync"
)

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT  = "IDENT"
	INT    = "INT"
	STRING = "STRING"

	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	LT = "<"
	GT = ">"

	EQ     = "=="
	NOT_EQ = "!="

	COMMA     = ","
	SEMICOLON = ";"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	// --- NOVO TOKEN ---
	PACKAGE = "PACKAGE"

	FUNCTION = "FUNCTION"
	LET      = "LET"
	RETURN   = "RETURN"
	IF       = "IF"
	ELSE     = "ELSE"
	INT_TYPE = "INT_TYPE" // Pode ser removido futuramente com a inferência de tipos
)

type Token struct {
	Type    TokenType
	Literal string
}

// Mapa de palavras-chave fixas
var keywords = map[string]TokenType{
	// --- NOVA PALAVRA-CHAVE ---
	"package": PACKAGE,
	"func":    FUNCTION,
	"let":     LET,
	"return":  RETURN,
	"if":      IF,
	"else":    ELSE,
	"int":     INT_TYPE,
}

// Logger global e mutex para evitar concorrência
var (
	logger   *log.Logger
	logFile  *os.File
	initOnce sync.Once
)

// Inicializa o logger uma vez só, de forma thread-safe
func initLogger() {
	initOnce.Do(func() {
		var err error
		logFile, err = os.OpenFile("log/token.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("Erro ao abrir arquivo de log token.log: %v", err)
		}
		logger = log.New(logFile, "TOKEN: ", log.LstdFlags)
		logger.Println("=== Nova sessão de log do token iniciada ===")
	})
}

// CloseLogger deve ser chamada no `main` para fechar o arquivo quando o programa terminar.
func CloseLogger() {
	if logFile != nil {
		logger.Println("=== Encerrando sessão de log do token ===")
		logFile.Close()
	}
}

// NewToken cria um token com log em arquivo
func NewToken(t TokenType, lit string) Token {
	initLogger()
	logger.Printf("Criando token - Tipo: %-10s | Literal: %q\n", t, lit)
	return Token{Type: t, Literal: lit}
}

// LookupIdent verifica se um identificador é uma palavra-chave
func LookupIdent(ident string) TokenType {
	initLogger()
	if tok, ok := keywords[ident]; ok {
		logger.Printf("LookupIdent: %q é palavra-chave -> TokenType: %q\n", ident, tok)
		return tok
	}
	logger.Printf("LookupIdent: %q é IDENT padrão\n", ident)
	return IDENT
}
