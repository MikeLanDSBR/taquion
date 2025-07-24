// ast.go
package ast

// Node é a interface base para todos os nós da AST.
type Node interface {
	TokenLiteral() string
}

// Statement marca nós que não produzem valor.
type Statement interface {
	Node
	statementNode()
}

// Expression marca nós que produzem valor.
type Expression interface {
	Node
	expressionNode()
}

// Program é o nó raiz: sequência de declarações.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}
