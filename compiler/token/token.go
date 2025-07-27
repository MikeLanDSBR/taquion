package token

import (
	"log"
	"os"
	"sync"
)

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	// Caracteres especiais
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identificadores + literais
	IDENT  = "IDENT"
	INT    = "INT"
	STRING = "STRING"

	// Operadores
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	MODULO   = "%"
	LT       = "<"
	GT       = ">"
	EQ       = "=="
	NOT_EQ   = "!="

	// Delimitadores
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	DOT       = "."

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Palavras-chave
	PACKAGE  = "PACKAGE"
	FUNCTION = "FUNCTION"
	CONST    = "CONST"
	LET      = "LET"
	RETURN   = "RETURN"
	IF       = "IF"
	ELSE     = "ELSE"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	WHILE    = "WHILE"
	BREAK    = "BREAK"
	CONTINUE = "CONTINUE"
	TYPE     = "TYPE"
)

var keywords = map[string]TokenType{
	"package":  PACKAGE,
	"func":     FUNCTION,
	"const":    CONST,
	"return":   RETURN,
	"if":       IF,
	"else":     ELSE,
	"true":     TRUE,
	"false":    FALSE,
	"let":      LET,
	"while":    WHILE,
	"break":    BREAK,
	"continue": CONTINUE,
	"type":     TYPE,
}

var (
	logger   *log.Logger
	logFile  *os.File
	initOnce sync.Once
)

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

func CloseLogger() {
	if logFile != nil {
		logger.Println("=== Encerrando sessão de log do token ===")
		logFile.Close()
	}
}

func NewToken(t TokenType, lit string) Token {
	initLogger()
	logger.Printf("Criando token - Tipo: %-10s | Literal: %q\n", t, lit)
	return Token{Type: t, Literal: lit}
}

func LookupIdent(ident string) TokenType {
	initLogger()
	if tok, ok := keywords[ident]; ok {
		logger.Printf("LookupIdent: %q é palavra-chave -> TokenType: %q\n", ident, tok)
		return tok
	}
	logger.Printf("LookupIdent: %q é IDENT padrão\n", ident)
	return IDENT
}
