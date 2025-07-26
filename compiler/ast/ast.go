package ast

import (
	"bytes"
	"log"
	"os"
	"sync"
)

// --- SISTEMA DE LOG ---
var (
	logger   *log.Logger
	logFile  *os.File
	initOnce sync.Once
)

// initLogger inicializa o logger global de forma segura.
func initLogger() {
	initOnce.Do(func() {
		if err := os.MkdirAll("log", 0755); err != nil {
			log.Fatalf("Erro ao criar diretório de log: %v", err)
		}
		var err error
		logFile, err = os.OpenFile("log/ast.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("Erro ao abrir o arquivo de log ast.log: %v", err)
		}
		logger = log.New(logFile, "AST: ", log.LstdFlags)
		logger.Println("=== Nova sessão de log da AST iniciada ===")
	})
}

// CloseLogger deve ser chamada no main para fechar o arquivo de log.
func CloseLogger() {
	if logFile != nil {
		logger.Println("=== Encerrando sessão de log da AST ===")
		logFile.Close()
	}
}

// --- INTERFACES BASE ---

// Node é a interface base para todos os nós da AST.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement marca nós que são declarações (não produzem valor).
type Statement interface {
	Node
	statementNode()
}

// Expression marca nós que são expressões (produzem valor).
type Expression interface {
	Node
	expressionNode()
}

// --- NÓ RAIZ ---

// Program é o nó raiz da AST: uma sequência de declarações.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	initLogger()
	logger.Println("Gerando string para o nó Program")
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}
