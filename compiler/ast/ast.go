package ast

import "bytes"

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
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}
