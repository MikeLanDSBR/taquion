// O pacote parser constrói a Árvore Sintática Abstrata (AST).
package parser

import (
	"fmt"
	"log"
	"os"
	"strconv" // Necessário para converter literais inteiros
	"taquion/compiler/ast"
	"taquion/compiler/lexer"
	"taquion/compiler/token"
)

// Constantes de precedência de operadores.
// Usadas para determinar a ordem de avaliação das expressões.
const (
	_ int = iota // LOWEST é 0
	LOWEST
	EQUALS      // ==
	LESSGREATER // > ou <
	SUM         // + ou -
	PRODUCT     // * ou /
	PREFIX      // -X ou !X
	CALL        // myFunction(X)
)

// Mapa que associa o tipo de token à sua precedência.
// Isso é fundamental para a análise de expressões infix.
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL, // Para chamadas de função
}

// prefixParseFn é o tipo de função para analisar expressões que começam com um token (operadores prefixo, identificadores, literais).
type prefixParseFn func() ast.Expression

// infixParseFn é o tipo de função para analisar expressões que têm um operador no meio (operadores infix).
type infixParseFn func(ast.Expression) ast.Expression

// Parser mantém o estado do analisador sintático.
type Parser struct {
	l         *lexer.Lexer // O lexer para obter tokens
	errors    []string     // Erros de parsing encontrados
	curToken  token.Token  // O token atual sendo inspecionado
	peekToken token.Token  // O próximo token (olhar à frente)

	prefixParseFns map[token.TokenType]prefixParseFn // Funções para expressões prefixo
	infixParseFns  map[token.TokenType]infixParseFn  // Funções para expressões infix

	logger           *log.Logger // Logger para rastreamento e depuração
	LogFile          *os.File    // Arquivo de log
	indentationLevel int         // Nível de indentação para logs de rastreamento
}

// New cria uma nova instância do Parser.
func New(l *lexer.Lexer) *Parser {
	// Configura o arquivo de log para o parser.
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

	// Registra as funções de parsing prefixo.
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)    // !EXPRESSION
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)   // -EXPRESSION
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression) // (EXPRESSION)

	// Registra as funções de parsing infix.
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

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

// parseStatement analisa uma única declaração.
func (p *Parser) parseStatement() ast.Statement {
	defer p.traceOut("parseStatement")
	p.traceIn("parseStatement")

	switch p.curToken.Type {
	case token.LET:
		p.logTrace("Encontrado token LET, chamando parseLetStatement")
		return p.parseLetStatement()
	case token.FUNCTION:
		p.logTrace("Encontrado token FUNCTION, chamando parseFunctionDeclaration")
		return p.parseFunctionDeclaration()
	case token.RETURN:
		p.logTrace("Encontrado token RETURN, chamando parseReturnStatement")
		return p.parseReturnStatement()
	default:
		// Se não for uma declaração conhecida, tenta analisar como uma ExpressionStatement.
		return p.parseExpressionStatement()
	}
}

// parseLetStatement analisa: `let <ident> = <expressão>;`
func (p *Parser) parseLetStatement() *ast.LetStatement {
	defer p.traceOut("parseLetStatement")
	p.traceIn("parseLetStatement")

	stmt := &ast.LetStatement{Token: p.curToken} // Token 'let'

	// Espera um identificador para o nome da variável.
	if !p.expectPeek(token.IDENT) {
		p.peekError(token.IDENT)
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Espera o operador de atribuição '='.
	if !p.expectPeek(token.ASSIGN) {
		p.peekError(token.ASSIGN)
		return nil
	}
	p.nextToken() // Avança para o token após o '='

	// Analisa a expressão do valor da variável (à direita do '=').
	stmt.Value = p.parseExpression(LOWEST)

	// Opcionalmente, espera um ponto e vírgula.
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseFunctionDeclaration analisa: `func <ident>(<params>) <return_type> { <body> }`
func (p *Parser) parseFunctionDeclaration() *ast.FunctionDeclaration {
	defer p.traceOut("parseFunctionDeclaration")
	p.traceIn("parseFunctionDeclaration")

	decl := &ast.FunctionDeclaration{Token: p.curToken} // Token 'func'

	// Espera o nome da função (um identificador).
	if !p.expectPeek(token.IDENT) {
		p.peekError(token.IDENT)
		return nil
	}
	decl.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Espera o parêntese de abertura para os parâmetros.
	if !p.expectPeek(token.LPAREN) {
		p.peekError(token.LPAREN)
		return nil
	}

	// TODO: Implementar o parsing de parâmetros da função.
	// Por enquanto, apenas avança até o parêntese de fechamento.
	p.nextToken() // Avança para o token dentro dos parênteses ou o RPAREN
	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		// Ignorar tokens dentro dos parênteses por enquanto, se não for RPAREN
		p.nextToken()
	}
	// Se o loop terminou por EOF ou algo inesperado, reporta erro.
	if !p.curTokenIs(token.RPAREN) {
		p.peekError(token.RPAREN)
		return nil
	}

	// Espera o tipo de retorno (um identificador, ex: 'int').
	if !p.peekTokenIs(token.IDENT) && !p.peekTokenIs(token.INT_TYPE) { // Adicionado INT_TYPE
		p.peekError(token.IDENT) // Ou um erro mais específico para tipo
		return nil
	}
	p.nextToken() // Avança para o token do tipo de retorno
	decl.ReturnType = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Espera a chave de abertura para o corpo da função.
	if !p.expectPeek(token.LBRACE) {
		p.peekError(token.LBRACE)
		return nil
	}
	decl.Body = p.parseBlockStatement() // Analisa o corpo da função como um BlockStatement.

	return decl
}

// parseBlockStatement analisa um bloco de código: `{ <statements> }`
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	defer p.traceOut("parseBlockStatement")
	p.traceIn("parseBlockStatement")

	block := &ast.BlockStatement{Token: p.curToken} // Token '{'
	block.Statements = []ast.Statement{}

	p.nextToken() // Avança para o token após o '{'

	// Continua analisando declarações dentro do bloco até encontrar '}' ou EOF.
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken() // Avança para o próximo token dentro do bloco.
	}
	return block
}

// parseReturnStatement analisa: `return <expressão>;`
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	defer p.traceOut("parseReturnStatement")
	p.traceIn("parseReturnStatement")

	stmt := &ast.ReturnStatement{Token: p.curToken} // Token 'return'

	p.nextToken() // Avança para o token após 'return'

	// Analisa a expressão que está sendo retornada.
	stmt.ReturnValue = p.parseExpression(LOWEST)

	// Opcionalmente, espera um ponto e vírgula.
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseExpressionStatement analisa uma declaração que consiste apenas de uma expressão.
// Ex: `x + y;` ou `5;`
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	defer p.traceOut("parseExpressionStatement")
	p.traceIn("parseExpressionStatement")

	stmt := &ast.ExpressionStatement{Token: p.curToken} // O primeiro token da expressão

	// Analisa a expressão completa, começando com a precedência mais baixa.
	stmt.Expression = p.parseExpression(LOWEST)

	// Opcionalmente, espera um ponto e vírgula.
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseExpression é o coração do analisador de expressões, usando o algoritmo Pratt.
// Ele analisa expressões com base na precedência dos operadores.
func (p *Parser) parseExpression(precedence int) ast.Expression {
	defer p.traceOut("parseExpression")
	p.traceIn(fmt.Sprintf("parseExpression (precedência: %d)", precedence))

	// 1. Procura uma função de parsing prefixo para o token atual.
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix() // Analisa a parte inicial da expressão (ex: um identificador 'x' ou um número '5').

	// 2. Continua analisando expressões infix enquanto a precedência do próximo token
	// for maior que a precedência atual.
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp // Não há função infix para o próximo token, então a expressão termina aqui.
		}

		p.nextToken()            // Avança para o token do operador infix (ex: '+').
		leftExp = infix(leftExp) // Chama a função infix para estender a expressão.
	}

	return leftExp
}

// parseIdentifier analisa um identificador e retorna um nó ast.Identifier.
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral analisa um literal inteiro e retorna um nó ast.IntegerLiteral.
func (p *Parser) parseIntegerLiteral() ast.Expression {
	defer p.traceOut("parseIntegerLiteral")
	p.traceIn("parseIntegerLiteral")

	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64) // 0 para base automática
	if err != nil {
		p.logTrace(fmt.Sprintf("ERRO: não foi possível converter '%s' para inteiro", p.curToken.Literal))
		p.errors = append(p.errors, fmt.Sprintf("não foi possível converter %q para inteiro", p.curToken.Literal))
		return nil
	}
	p.logTrace(fmt.Sprintf("Literal de inteiro analisado, valor: %d", value))

	lit.Value = value
	return lit
}

// parsePrefixExpression analisa expressões com operadores prefixo (ex: `!x`, `-5`).
func (p *Parser) parsePrefixExpression() ast.Expression {
	defer p.traceOut("parsePrefixExpression")
	p.traceIn("parsePrefixExpression")

	expression := &ast.PrefixExpression{ // Assumindo que você terá uma struct PrefixExpression em ast.go
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()                                // Avança para o operando da expressão prefixo.
	expression.Right = p.parseExpression(PREFIX) // Analisa o operando com precedência PREFIX.
	return expression
}

// parseInfixExpression analisa expressões com operadores infix (ex: `x + y`, `a == b`).
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	defer p.traceOut("parseInfixExpression")
	p.traceIn("parseInfixExpression")

	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()                  // Pega a precedência do operador infix atual.
	p.nextToken()                                    // Avança para o operando da direita.
	expression.Right = p.parseExpression(precedence) // Analisa o operando da direita com a precedência do operador.
	return expression
}

// parseGroupedExpression analisa expressões entre parênteses (ex: `(x + y)`).
func (p *Parser) parseGroupedExpression() ast.Expression {
	defer p.traceOut("parseGroupedExpression")
	p.traceIn("parseGroupedExpression")

	exp := p.parseExpression(LOWEST) // Analisa a expressão dentro dos parênteses.
	if !p.expectPeek(token.RPAREN) {
		p.peekError(token.RPAREN)
		return nil
	}
	return exp
}

// --- Funções Auxiliares e de Erro ---

// Errors retorna a lista de erros de parsing.
func (p *Parser) Errors() []string { return p.errors }

// nextToken avança os tokens curToken e peekToken.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
	p.logger.Printf("Avançando token: cur=%-10s ('%s') | peek=%-10s ('%s')\n", p.curToken.Type, p.curToken.Literal, p.peekToken.Type, p.peekToken.Literal)
}

// curTokenIs verifica se o token atual é do tipo esperado.
func (p *Parser) curTokenIs(t token.TokenType) bool { return p.curToken.Type == t }

// peekTokenIs verifica se o próximo token é do tipo esperado.
func (p *Parser) peekTokenIs(t token.TokenType) bool { return p.peekToken.Type == t }

// expectPeek verifica se o próximo token é do tipo esperado e, se for, avança.
// Caso contrário, adiciona um erro.
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t) // Adiciona um erro se o token esperado não for encontrado.
	return false
}

// peekPrecedence retorna a precedência do próximo token.
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// curPrecedence retorna a precedência do token atual.
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// noPrefixParseFnError adiciona um erro quando não há função prefixo registrada para um token.
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("nenhuma função de parsing de prefixo encontrada para %s", t)
	p.errors = append(p.errors, msg)
	p.logTrace(msg)
}

// peekError adiciona um erro quando o próximo token não é o esperado.
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("esperava o próximo token ser %s, mas obteve %s (%q)",
		t, p.peekToken.Type, p.peekToken.Literal)
	p.errors = append(p.errors, msg)
	p.logTrace(msg)
}

// registerPrefix registra uma função de parsing prefixo para um tipo de token.
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix registra uma função de parsing infix para um tipo de token.
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// --- Funções de Logging ---

func (p *Parser) logTrace(msg string) {
	indent := ""
	for i := 0; i < p.indentationLevel; i++ {
		indent += "    " // 4 espaços para indentação
	}
	p.logger.Printf("%s%s\n", indent, msg)
}

func (p *Parser) traceIn(funcName string) {
	p.logTrace(">> " + funcName)
	p.indentationLevel++
}

func (p *Parser) traceOut(funcName string) {
	p.indentationLevel--
	p.logTrace("<< " + funcName)
}
