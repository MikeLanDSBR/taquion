package parser

import (
	"taquion/compiler/ast"
	"taquion/compiler/lexer"
	"testing"
)

func TestReturnStatement(t *testing.T) {
	input := `
return 5;
return 10;
return 993322;
`
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() retornou nil")
	}
	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements não contém 3 declarações. Recebido=%d", len(program.Statements))
	}

	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt não é *ast.ReturnStatement. Recebido=%T", stmt)
			continue
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("returnStmt.TokenLiteral não é 'return', recebido %q", returnStmt.TokenLiteral())
		}
	}
}
