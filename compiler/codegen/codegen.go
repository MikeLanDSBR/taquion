// O pacote codegen é responsável por traduzir a Árvore Sintática Abstrata (AST)
// para a Representação Intermediária (IR) do LLVM.
package codegen

import (
	"fmt"
	"strconv"
	"taquion/compiler/ast"
	"taquion/compiler/token" // Importar o pacote token para acessar os tipos de operadores

	"tinygo.org/x/go-llvm" // A biblioteca correta e mantida
)

// CodeGenerator mantém o estado durante a geração de código.
type CodeGenerator struct {
	module      llvm.Module
	builder     llvm.Builder
	context     llvm.Context
	symbolTable map[string]llvm.Value // Tabela de símbolos para rastrear variáveis.
}

// NewCodeGenerator cria uma nova instância do gerador de código.
func NewCodeGenerator() *CodeGenerator {
	// A API do go-llvm requer um "Contexto" para gerenciar a memória e os tipos.
	ctx := llvm.NewContext()
	return &CodeGenerator{
		context:     ctx,
		module:      ctx.NewModule("main_module"),
		builder:     ctx.NewBuilder(),
		symbolTable: make(map[string]llvm.Value),
	}
}

// Generate é o ponto de entrada que traduz a AST para um módulo LLVM.
func (c *CodeGenerator) Generate(program *ast.Program) llvm.Module {
	// Itera sobre as declarações de alto nível no programa.
	for _, stmt := range program.Statements {
		// Procura pela declaração da função 'main'.
		if funcDecl, ok := stmt.(*ast.FunctionDeclaration); ok && funcDecl.Name.Value == "main" {
			// Define o tipo da função main: int main()
			// Retorna i32 (inteiro de 32 bits), sem parâmetros, não variadic.
			funcType := llvm.FunctionType(c.context.Int32Type(), []llvm.Type{}, false)
			// Adiciona a função 'main' ao módulo LLVM.
			mainFunc := llvm.AddFunction(c.module, "main", funcType)
			// Cria um bloco básico de entrada para a função.
			entryBlock := llvm.AddBasicBlock(mainFunc, "entry")
			// Define o ponto de inserção do builder para o final do bloco de entrada.
			c.builder.SetInsertPointAtEnd(entryBlock)

			// Gera o código para todas as declarações dentro do corpo da função main.
			for _, bodyStmt := range funcDecl.Body.Statements {
				c.genStatement(bodyStmt)
			}

			// Após gerar o corpo da função, verifica se o último bloco básico tem um terminador.
			// Isso é crucial para garantir que o LLVM IR seja válido.
			currentBlock := c.builder.GetInsertBlock()
			lastInst := currentBlock.LastInstruction()
			// Se o bloco não tem uma última instrução ou a última instrução não é um retorno,
			// adiciona um retorno padrão (ret 0).
			if lastInst.IsNil() || lastInst.IsAReturnInst().IsNil() {
				// Adiciona um `return 0` padrão se nenhum `return` explícito for encontrado.
				c.builder.CreateRet(llvm.ConstInt(c.context.Int32Type(), 0, false))
			}
			break // Sai do loop após encontrar e processar a função main.
		}
	}

	return c.module // Retorna o módulo LLVM completo.
}

// genStatement gera código LLVM IR para uma declaração AST.
func (c *CodeGenerator) genStatement(stmt ast.Statement) {
	switch node := stmt.(type) {
	case *ast.LetStatement:
		// Para uma declaração 'let x = 10;'
		// 1. Gera o código para a expressão à direita do '=' (o valor inicial).
		val := c.genExpression(node.Value)
		// 2. Aloca espaço na pilha para a variável.
		// O nome da variável no LLVM IR será o nome do identificador.
		ptr := c.builder.CreateAlloca(c.context.Int32Type(), node.Name.Value)
		// 3. Armazena o valor da expressão na memória alocada.
		c.builder.CreateStore(val, ptr)
		// 4. Adiciona a variável e seu ponteiro LLVM à tabela de símbolos.
		c.symbolTable[node.Name.Value] = ptr
	case *ast.ReturnStatement:
		// Para uma declaração 'return x + y;'
		// 1. Gera o código para a expressão que deve ser retornada.
		val := c.genExpression(node.ReturnValue)
		// 2. Cria a instrução de retorno LLVM com o valor calculado.
		c.builder.CreateRet(val)
	// Adicione outros tipos de declarações aqui conforme a linguagem evolui (ex: IfStatement, WhileStatement).
	default:
		// Log ou erro para declarações não suportadas.
		fmt.Printf("Declaração não suportada: %T\n", node)
	}
}

// genExpression gera código LLVM IR para uma expressão AST e retorna o llvm.Value resultante.
func (c *CodeGenerator) genExpression(expr ast.Expression) llvm.Value {
	switch node := expr.(type) {
	case *ast.IntegerLiteral:
		// Para um literal inteiro (ex: 10, 32)
		val, err := strconv.ParseInt(node.TokenLiteral(), 10, 64)
		if err != nil {
			// Lidar com erro de parsing de inteiro, talvez retornar um valor padrão ou panicar.
			panic(fmt.Sprintf("Erro ao parsear literal inteiro: %s", node.TokenLiteral()))
		}
		// Cria uma constante inteira LLVM com o valor e o tipo i32.
		return llvm.ConstInt(c.context.Int32Type(), uint64(val), false)
	case *ast.Identifier:
		// Para um identificador (ex: x, y)
		// 1. Procura o ponteiro da variável na tabela de símbolos.
		if ptr, ok := c.symbolTable[node.Value]; ok {
			// 2. Carrega o valor da variável da memória.
			return c.builder.CreateLoad(c.context.Int32Type(), ptr, node.Value+"_val")
		}
		// Se a variável não for encontrada, é um erro.
		panic(fmt.Sprintf("variável não definida: %s", node.Value))
	case *ast.InfixExpression:
		// Para uma expressão infix (ex: x + y)
		// 1. Gera o código para a expressão do lado esquerdo.
		left := c.genExpression(node.Left)
		// 2. Gera o código para a expressão do lado direito.
		right := c.genExpression(node.Right)

		// 3. Com base no operador, cria a instrução LLVM apropriada.
		switch node.Operator {
		case token.PLUS:
			// Cria uma instrução de adição de inteiros.
			return c.builder.CreateAdd(left, right, "addtmp")
		case token.MINUS:
			// Cria uma instrução de subtração de inteiros.
			return c.builder.CreateSub(left, right, "subtmp")
		case token.ASTERISK:
			// Cria uma instrução de multiplicação de inteiros.
			return c.builder.CreateMul(left, right, "multmp")
		case token.SLASH:
			// Cria uma instrução de divisão de inteiros (assumindo divisão com sinal por padrão).
			return c.builder.CreateSDiv(left, right, "divtmp")
		// Adicione outros operadores aqui (ex: comparação, lógicos).
		default:
			// Operador não suportado.
			panic(fmt.Sprintf("operador infix não suportado: %s", node.Operator))
		}
	// Adicione outros tipos de expressões aqui conforme a linguagem evolui (ex: CallExpression).
	default:
		// Log ou erro para expressões não suportadas.
		fmt.Printf("Expressão não suportada: %T\n", node)
		return llvm.Value{} // Retorna um valor LLVM nulo ou um erro.
	}
}
