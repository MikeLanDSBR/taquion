// O pacote parser constrói a Árvore Sintática Abstrata (AST).
package parser

import (
	"log"
	"os"
	"taquion/compiler/ast"
	"taquion/compiler/lexer"
	"taquion/compiler/token"
)

// prefixParseFn é o tipo de função para analisar expressões que começam com um token.
type prefixParseFn func() ast.Expression

// infixParseFn é o tipo de função para analisar expressões que têm um operador no meio.
type infixParseFn func(ast.Expression) ast.Expression

// Parser mantém o estado do analisador sintático.
type Parser struct {
	l                *lexer.Lexer                      // O lexer para obter tokens
	errors           []string                          // Erros de parsing encontrados
	curToken         token.Token                       // O token atual sendo inspecionado
	peekToken        token.Token                       // O próximo token (olhar à frente)
	prefixParseFns   map[token.TokenType]prefixParseFn // Funções para expressões prefixo
	infixParseFns    map[token.TokenType]infixParseFn  // Funções para expressões infix
	logger           *log.Logger                       // Logger para rastreamento e depuração
	LogFile          *os.File                          // Arquivo de log
	indentationLevel int                               // Nível de indentação para logs
}

// New cria uma nova instância do Parser.
func New(l *lexer.Lexer) *Parser {
	file, err := os.OpenFile("log/parser.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo de log do parser: %v", err)
	}

	p := &Parser{
		l:                l,
		errors:           []string{},
		logger:           log.New(file, "PARSER: ", log.LstdFlags),
		LogFile:          file,
		indentationLevel: 0,
	}
	p.logger.Println("Iniciando nova sessão de parsing.")

	// Inicializa os mapas de funções de parsing.
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.infixParseFns = make(map[token.TokenType]infixParseFn)

	// Registra todas as funções de parsing.
	p.registerPrefixFns()
	p.registerInfixFns()

	// Lê dois tokens para preencher curToken e peekToken.
	p.nextToken()
	p.nextToken()
	return p
}

// ParseProgram analisa o programa completo e retorna o nó raiz da AST.
func (p *Parser) ParseProgram() *ast.Program {
	defer p.traceOut("ParseProgram")
	p.traceIn("ParseProgram")

	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// Continua analisando declarações até o fim do arquivo (EOF).
	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken() // Avança para o próximo token após analisar uma declaração.
	}
	return program
}
