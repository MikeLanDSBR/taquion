// Arquivo: ast/ast.go
package ast

import (
	"bytes"
	"log"
	"os"
	"sync"
	"taquion/compiler/token"
)

var (
	logger   *log.Logger
	logFile  *os.File
	initOnce sync.Once
)

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

func CloseLogger() {
	if logFile != nil {
		logger.Println("=== Encerrando sessão de log da AST ===")
		logFile.Close()
	}
}

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

type CompositeLiteral struct {
	Token    token.Token // o token de abertura '{'
	TypeName *Identifier // Pessoa
	Fields   []*KeyValueExpr
}

func (cl *CompositeLiteral) String() string {
	var out bytes.Buffer

	out.WriteString(cl.TypeName.String())
	out.WriteString(" { ")

	for i, field := range cl.Fields {
		out.WriteString(field.String())
		if i < len(cl.Fields)-1 {
			out.WriteString(", ")
		}
	}

	out.WriteString(" }")
	return out.String()
}

func (cl *CompositeLiteral) expressionNode()      {}
func (cl *CompositeLiteral) TokenLiteral() string { return cl.Token.Literal }

type KeyValueExpr struct {
	Key   *Identifier
	Value Expression
}

func (kv *KeyValueExpr) String() string {
	return kv.Key.String() + ": " + kv.Value.String()
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

type Identifier struct {
	Token token.Token
	Value string
	Type  *Identifier
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

func (i *Identifier) String() string {
	if i.Type != nil {
		return i.Value + ": " + i.Type.String()
	}
	return i.Value
}

func (i *Identifier) TypeNode() bool {
	return i.Type != nil
}

func (i *Identifier) TypeString() string {
	if i.Type != nil {
		return i.Type.String()
	}
	return "unknown"
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}
