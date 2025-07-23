package lexer

import (
	"taquion/compiler/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `func main() int {
		return 42
	}`

	// Esta é a lista de tokens que esperamos que nosso Lexer produza.
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.FUNCTION, "func"},
		{token.IDENT, "main"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LookupIdent("int"), "int"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.INT, "42"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

	// Cria um novo Lexer com nosso input.
	l := New(input)

	// Itera sobre nossa lista de testes.
	for i, tt := range tests {
		// Pede ao Lexer pelo próximo token.
		tok := l.NextToken()

		// Verifica se o TIPO do token é o que esperávamos.
		if tok.Type != tt.expectedType {
			t.Fatalf("teste[%d] - tipo de token errado. esperado=%q, recebido=%q",
				i, tt.expectedType, tok.Type)
		}

		// Verifica se o LITERAL do token é o que esperávamos.
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("teste[%d] - literal errado. esperado=%q, recebido=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// Adicione esta função ao final de lexer_test.go

func TestMultipleReturnStatements(t *testing.T) {
	input := `
return 5;
return 10;
return 993322;
`

	// A lista de tokens que esperamos que o Lexer produza
	expectedTokens := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.RETURN, "return"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.RETURN, "return"},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.RETURN, "return"},
		{token.INT, "993322"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""}, // O Fim de Arquivo deve ser o último
	}

	l := New(input)

	// Itera sobre a lista de tokens esperados e compara com o que o lexer produz
	for i, tt := range expectedTokens {
		tok := l.NextToken()

		// Imprime o token gerado para podermos ver o que está acontecendo
		t.Logf("Token %d: Type=[%s] Literal=[%s]", i, tok.Type, tok.Literal)

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tipo de token errado. esperado=%q, recebido=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal errado. esperado=%q, recebido=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
