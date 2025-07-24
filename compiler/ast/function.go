package ast

import (
	"bytes"
	"strings"
	"taquion/compiler/token"
)

// Parameter representa um parâmetro de função, como `a int`
type Parameter struct {
	Token token.Token // O token do nome do parâmetro
	Name  *Identifier
	Type  *Identifier
}

func (p *Parameter) Node()                {}
func (p *Parameter) TokenLiteral() string { return p.Token.Literal }
func (p *Parameter) String() string       { return p.Name.String() + " " + p.Type.String() }

// FunctionDeclaration representa a definição de uma função.
type FunctionDeclaration struct {
	Token      token.Token // O token 'func'
	Name       *Identifier
	Parameters []*Parameter // MODIFICADO: Usa a nova struct Parameter
	ReturnType *Identifier
	Body       *BlockStatement
}

func (fd *FunctionDeclaration) statementNode()       {}
func (fd *FunctionDeclaration) TokenLiteral() string { return fd.Token.Literal }
func (fd *FunctionDeclaration) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range fd.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(fd.TokenLiteral() + " ")
	out.WriteString(fd.Name.String())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	if fd.ReturnType != nil {
		out.WriteString(fd.ReturnType.String() + " ")
	}
	out.WriteString(fd.Body.String())
	return out.String()
}
