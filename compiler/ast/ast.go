// O pacote ast define as estruturas de dados para a Árvore Sintática Abstrata (AST),
// que é a representação estruturada do código-fonte gerada pelo Parser.
package ast

import "taquion/compiler/token"

// Node é a interface que todos os nós da AST devem implementar.
type Node interface {
	// TokenLiteral retorna o valor literal do token associado ao nó.
	TokenLiteral() string
}

// Statement representa uma instrução que não produz um valor (ex: `return 5;`).
type Statement interface {
	Node
	statementNode() // Método "dummy" para marcar nós como statements.
}

// Expression representa uma instrução que produz um valor (ex: `5 + 5`).
type Expression interface {
	Node
	expressionNode() // Método "dummy" para marcar nós como expressions.
}

// Program é o nó raiz de toda a AST. Um programa é uma sequência de declarações.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// ReturnStatement representa uma declaração de retorno: `return <expressão>;`
type ReturnStatement struct {
	Token       token.Token // O token `token.RETURN`.
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// Identifier representa um nome de variável, função, etc.
type Identifier struct {
	Token token.Token // O token `token.IDENT`.
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// ExpressionStatement é uma declaração que consiste apenas de uma expressão.
type ExpressionStatement struct {
	Token      token.Token // O primeiro token da expressão.
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
