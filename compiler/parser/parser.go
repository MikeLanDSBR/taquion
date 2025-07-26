package parser

import (
	"log"
	"os"
	"taquion/compiler/ast"
	"taquion/compiler/lexer"
	"taquion/compiler/token"
)

// (O restante da sua struct Parser e das funções prefix/infix permanece o mesmo)
type prefixParseFn func() ast.Expression
type infixParseFn func(ast.Expression) ast.Expression

type Parser struct {
	l                *lexer.Lexer
	errors           []string
	curToken         token.Token
	peekToken        token.Token
	prefixParseFns   map[token.TokenType]prefixParseFn
	infixParseFns    map[token.TokenType]infixParseFn
	logger           *log.Logger
	LogFile          *os.File
	indentationLevel int
}

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

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.infixParseFns = make(map[token.TokenType]infixParseFn)

	p.registerPrefixFns()
	p.registerInfixFns()

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

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		// CORREÇÃO: O parser só avança para o próximo token depois que uma
		// declaração inteira foi analisada. A chamada a nextToken() foi
		// movida para dentro do loop de parseStatement quando necessário.
		// Esta linha foi a causa de muitos erros de sincronia.
		p.nextToken()
	}
	return program
}
