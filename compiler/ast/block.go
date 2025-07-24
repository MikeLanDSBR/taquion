// block.go
package ast

import "taquion/compiler/token"

// BlockStatement = { <statements> }
type BlockStatement struct {
	Token      token.Token // o token "{"
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
