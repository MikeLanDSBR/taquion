// Função: Ponto de entrada e funções principais do gerador de código.
package codegen

import (
	"fmt"
	"log"
	"os"
	"taquion/compiler/ast"

	"tinygo.org/x/go-llvm"
)

// CodeGenerator mantém o estado durante a geração de código.
type CodeGenerator struct {
	module      llvm.Module
	builder     llvm.Builder
	context     llvm.Context
	symbolTable map[string]llvm.Value

	// --- CAMPOS DE LOGGING ---
	logger           *log.Logger
	logFile          *os.File
	indentationLevel int
}

// NewCodeGenerator cria uma nova instância do gerador de código.
func NewCodeGenerator() *CodeGenerator {
	file, err := os.OpenFile("log/codegen.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo de log do codegen: %v", err)
	}

	ctx := llvm.NewContext()
	cg := &CodeGenerator{
		context:          ctx,
		module:           ctx.NewModule("main_module"),
		builder:          ctx.NewBuilder(),
		symbolTable:      make(map[string]llvm.Value),
		logger:           log.New(file, "CODEGEN: ", log.LstdFlags),
		logFile:          file,
		indentationLevel: 0,
	}

	cg.logger.Println("Iniciando nova sessão de geração de código.")
	return cg
}

// Close encerra os recursos do gerador de código, como o arquivo de log.
func (c *CodeGenerator) Close() {
	if c.logFile != nil {
		c.logger.Println("Encerrando sessão de geração de código.")
		c.logFile.Close()
	}
}

// Generate é o ponto de entrada que traduz a AST para um módulo LLVM.
func (c *CodeGenerator) Generate(program *ast.Program) llvm.Module {
	defer c.traceOut("Generate")
	c.traceIn("Generate")

	for _, stmt := range program.Statements {
		if funcDecl, ok := stmt.(*ast.FunctionDeclaration); ok && funcDecl.Name.Value == "main" {
			c.logTrace(fmt.Sprintf("Encontrada a função 'main'. Gerando corpo..."))

			funcType := llvm.FunctionType(c.context.Int32Type(), []llvm.Type{}, false)
			mainFunc := llvm.AddFunction(c.module, "main", funcType)
			entryBlock := llvm.AddBasicBlock(mainFunc, "entry")
			c.builder.SetInsertPointAtEnd(entryBlock)

			for _, bodyStmt := range funcDecl.Body.Statements {
				c.genStatement(bodyStmt)
			}

			currentBlock := c.builder.GetInsertBlock()
			lastInst := currentBlock.LastInstruction()
			if lastInst.IsNil() || lastInst.IsAReturnInst().IsNil() {
				c.logTrace("Nenhum 'return' explícito encontrado no final da função 'main'. Adicionando 'return 0' padrão.")
				c.builder.CreateRet(llvm.ConstInt(c.context.Int32Type(), 0, false))
			}
			if !isBlockTerminated(currentBlock) {
				c.logTrace("Nenhum terminador explícito encontrado no final da função 'main'. Adicionando 'return 0' padrão.")
				c.builder.CreateRet(llvm.ConstInt(c.context.Int32Type(), 0, false))
			}
			break
		}
	}

	return c.module
}

func isBlockTerminated(block llvm.BasicBlock) bool {
	lastInst := block.LastInstruction()
	if lastInst.IsNil() {
		return false
	}
	opcode := lastInst.InstructionOpcode()
	switch opcode {
	case llvm.Ret, llvm.Br, llvm.Switch, llvm.IndirectBr, llvm.Invoke, llvm.Unreachable, llvm.Resume:
		return true
	default:
		return false
	}
}

func (c *CodeGenerator) logTrace(msg string) {
	indent := ""
	for i := 0; i < c.indentationLevel; i++ {
		indent += "    "
	}
	c.logger.Printf("%s%s\n", indent, msg)
}

func (c *CodeGenerator) traceIn(funcName string) {
	c.logTrace(">> " + funcName)
	c.indentationLevel++
}

func (c *CodeGenerator) traceOut(funcName string) {
	c.indentationLevel--
	c.logTrace("<< " + funcName)
}
