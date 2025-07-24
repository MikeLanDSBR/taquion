// function.go
package ast

import "taquion/compiler/token"

// FunctionDeclaration = func name(params) returnType { body }
type FunctionDeclaration struct {
	Token      token.Token // token "func"
	Name       *Identifier
	Parameters []*Identifier
	ReturnType *Identifier
	Body       *BlockStatement
}

func (fd *FunctionDeclaration) statementNode()       {}
func (fd *FunctionDeclaration) TokenLiteral() string { return fd.Token.Literal }
