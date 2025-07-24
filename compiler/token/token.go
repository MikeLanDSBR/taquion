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

	FUNCTION = "FUNCTION"
	LET      = "LET"
	RETURN   = "RETURN"
	IF       = "IF"
	ELSE     = "ELSE"
	INT_TYPE = "INT_TYPE"
)

type Token struct {
	Type    TokenType
	Literal string
}

// Mapa keywords fixas
var keywords = map[string]TokenType{
	"func":   FUNCTION,
	"let":    LET,
	"return": RETURN,
	"if":     IF,
	"else":   ELSE,
	"int":    INT_TYPE,
}

// Logger global e mutex pra evitar rixa de concorrência
var (
	logger   *log.Logger
	logFile  *os.File
	initOnce sync.Once
)

// Inicializa o logger uma vez só, thread-safe
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

// Chame essa função no `main` para fechar o arquivo quando o programa terminar.
func CloseLogger() {
	if logFile != nil {
		logger.Println("=== Encerrando sessão de log do token ===")
		logFile.Close()
	}
}

// Cria token com log em arquivo
func NewToken(t TokenType, lit string) Token {
	initLogger()
	logger.Printf("Criando token - Tipo: %-10s | Literal: %q\n", t, lit)
	return Token{Type: t, Literal: lit}
}

// LookupIdent com log em arquivo
func LookupIdent(ident string) TokenType {
	initLogger()
	if tok, ok := keywords[ident]; ok {
		logger.Printf("LookupIdent: %q é palavra-chave -> TokenType: %q\n", ident, tok)
		return tok
	}
	logger.Printf("LookupIdent: %q é IDENT padrão\n", ident)
	return IDENT
}
