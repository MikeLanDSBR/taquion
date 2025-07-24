package ast

import (
	"bytes"
	"taquion/compiler/token"
)

// BlockStatement representa um bloco de código: { ...declarações... }
type BlockStatement struct {
	Token      token.Token // O token '{'
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}
