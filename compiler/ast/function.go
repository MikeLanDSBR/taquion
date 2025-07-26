package ast

import (
	"bytes"
	"strings"
	"taquion/compiler/token"
)

// Parameter representa um parâmetro de função.
// Com a remoção de tipos explícitos, ele contém apenas um nome.
type Parameter struct {
	Token token.Token // O token do nome do parâmetro
	Name  *Identifier
	Type  *Identifier // Mantido para possível análise semântica futura, mas não preenchido pelo parser.
}

func (p *Parameter) Node()                {}
func (p *Parameter) TokenLiteral() string { return p.Token.Literal }
func (p *Parameter) String() string {
	// Garante que o logger seja inicializado. Uma chamada a String() em qualquer
	// nó da AST deve ser suficiente para criar o arquivo de log.
	initLogger()
	// Como não temos mais tipos explícitos na sintaxe, apenas retornamos o nome.
	return p.Name.String()
}

// FunctionDeclaration representa a definição de uma função.
type FunctionDeclaration struct {
	Token      token.Token // O token 'func'
	Name       *Identifier
	Parameters []*Parameter
	ReturnType *Identifier // Mantido para possível análise semântica futura.
	Body       *BlockStatement
}

func (fd *FunctionDeclaration) statementNode()       {}
func (fd *FunctionDeclaration) TokenLiteral() string { return fd.Token.Literal }

// String() agora garante a inicialização do log e registra a sua própria execução.
func (fd *FunctionDeclaration) String() string {
	// Garante que o logger da AST esteja inicializado.
	initLogger()

	// Loga a chamada para este método específico para facilitar a depuração.
	logger.Printf("Gerando string para FunctionDeclaration: %s", fd.Name.Value)

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
	// Não imprimimos mais o tipo de retorno, pois ele é inferido.
	if fd.Body != nil {
		out.WriteString(fd.Body.String())
	}
	return out.String()
}
