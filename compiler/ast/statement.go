package ast

import (
	"bytes"
	"taquion/compiler/token"
)

// --- DECLARAÇÕES (STATEMENTS) ---

// PackageStatement representa uma declaração 'package <nome>'
type PackageStatement struct {
	Token token.Token // O token 'package'
	Name  *Identifier // O nome do pacote
}

func (ps *PackageStatement) statementNode()       {}
func (ps *PackageStatement) TokenLiteral() string { return ps.Token.Literal }
func (ps *PackageStatement) String() string {
	initLogger()
	logger.Printf("Gerando string para PackageStatement: %s %s", ps.TokenLiteral(), ps.Name.String())
	return ps.TokenLiteral() + " " + ps.Name.String() + ";"
}

// --- NOVA STRUCT PARA CONST ---
// ConstStatement = const <nome> = <valor>;
type ConstStatement struct {
	Token token.Token // O token 'const'
	Name  *Identifier
	Value Expression
}

func (cs *ConstStatement) statementNode()       {}
func (cs *ConstStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ConstStatement) String() string {
	initLogger()
	logger.Printf("Gerando string para ConstStatement: %s", cs.Name.Value)
	var out bytes.Buffer
	out.WriteString(cs.TokenLiteral() + " ")
	out.WriteString(cs.Name.String())
	out.WriteString(" = ")
	if cs.Value != nil {
		out.WriteString(cs.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// LetStatement = let <nome> = <valor>;
type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	initLogger()
	logger.Printf("Gerando string para LetStatement: %s", ls.Name.Value)
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// ReturnStatement = return <expr>;
type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	initLogger()
	logger.Println("Gerando string para ReturnStatement")
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

// ExpressionStatement = <expr>;
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	initLogger()
	logger.Println("Gerando string para ExpressionStatement")
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// AssignmentStatement = <nome> = <valor>;
type AssignmentStatement struct {
	Token token.Token // O token '='
	Name  *Identifier
	Value Expression
}

func (as *AssignmentStatement) statementNode()       {}
func (as *AssignmentStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignmentStatement) String() string {
	initLogger()
	logger.Printf("Gerando string para AssignmentStatement: %s", as.Name.Value)
	var out bytes.Buffer
	out.WriteString(as.Name.String())
	out.WriteString(" = ")
	if as.Value != nil {
		out.WriteString(as.Value.String())
	}
	out.WriteString(";")
	return out.String()
}
