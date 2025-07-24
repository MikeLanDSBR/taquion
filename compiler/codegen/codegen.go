// Arquivo: codegen/codegen.go
// Função: Ponto de entrada e funções principais do gerador de código.
package codegen

import (
	"fmt"
	"log"
	"os"
	"taquion/compiler/ast"

	"tinygo.org/x/go-llvm"
)

// SymbolEntry armazena o ponteiro e o tipo de uma variável na tabela de símbolos.
type SymbolEntry struct {
	Ptr llvm.Value
	Typ llvm.Type
}

// CodeGenerator mantém o estado durante a geração de código.
type CodeGenerator struct {
	module      llvm.Module
	builder     llvm.Builder
	context     llvm.Context
	symbolTable []map[string]SymbolEntry // Pilha de mapas para gerenciar escopo com tipos.

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
		symbolTable:      []map[string]SymbolEntry{make(map[string]SymbolEntry)}, // Escopo global
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

	// MODIFICADO: Agora geramos o código para TODAS as declarações no escopo global.
	// Isso irá processar 'func add(...)' e depois 'func main(...)'.
	for _, stmt := range program.Statements {
		c.genStatement(stmt)
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

// --- FUNÇÕES DE ESCOPO ---

func (c *CodeGenerator) pushScope() {
	c.logTrace("=> Entrando em novo escopo")
	c.symbolTable = append(c.symbolTable, make(map[string]SymbolEntry))
}

func (c *CodeGenerator) popScope() {
	c.logTrace("<= Saindo do escopo")
	c.symbolTable = c.symbolTable[:len(c.symbolTable)-1]
}

func (c *CodeGenerator) setSymbol(name string, entry SymbolEntry) {
	c.logTrace(fmt.Sprintf("Definindo símbolo '%s' no escopo atual", name))
	c.symbolTable[len(c.symbolTable)-1][name] = entry
}

func (c *CodeGenerator) getSymbol(name string) (SymbolEntry, bool) {
	c.logTrace(fmt.Sprintf("Procurando símbolo '%s'", name))
	for i := len(c.symbolTable) - 1; i >= 0; i-- {
		if entry, ok := c.symbolTable[i][name]; ok {
			c.logTrace(fmt.Sprintf("Símbolo '%s' encontrado no escopo %d", name, i))
			return entry, true
		}
	}
	c.logTrace(fmt.Sprintf("Símbolo '%s' não encontrado em nenhum escopo", name))
	return SymbolEntry{}, false
}

// --- FUNÇÕES DE LOGGING AUXILIARES ---

func (c *CodeGenerator) logTrace(msg string) {
	indent := ""
	for i := 0; i < len(c.symbolTable); i++ {
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
