package codegen

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"taquion/compiler/ast"

	"tinygo.org/x/go-llvm"
)

// (initLogger, CloseLogger, SymbolEntry permanecem os mesmos)
var (
	logger   *log.Logger
	logFile  *os.File
	initOnce sync.Once
)

func initLogger() {
	initOnce.Do(func() {
		if err := os.MkdirAll("log", 0755); err != nil {
			log.Fatalf("Erro ao criar diretório de log: %v", err)
		}
		var err error
		logFile, err = os.OpenFile("log/codegen.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("Erro ao abrir o arquivo de log codegen.log: %v", err)
		}
		logger = log.New(logFile, "CODEGEN: ", log.LstdFlags)
		logger.Println("=== Nova sessão de log do codegen iniciada ===")
	})
}

func CloseLogger() {
	if logFile != nil {
		logger.Println("=== Encerrando sessão de log do codegen ===")
		logFile.Close()
	}
}

type SymbolEntry struct {
	Ptr llvm.Value
	Typ llvm.Type
}

// --- STRUCT MODIFICADA ---
type CodeGenerator struct {
	module           llvm.Module
	builder          llvm.Builder
	context          llvm.Context
	symbolTable      []map[string]SymbolEntry
	indentationLevel int
	// NOVO CAMPO: Armazena o TIPO de retorno da função atual. É mais seguro que armazenar a função inteira.
	currentFunctionReturnType llvm.Type
}

func NewCodeGenerator() *CodeGenerator {
	initLogger()

	ctx := llvm.NewContext()
	cg := &CodeGenerator{
		context:          ctx,
		module:           ctx.NewModule("main_module"),
		builder:          ctx.NewBuilder(),
		symbolTable:      []map[string]SymbolEntry{make(map[string]SymbolEntry)},
		indentationLevel: 0,
	}

	logger.Println("Nova instância de CodeGenerator criada.")
	return cg
}

// (O restante do arquivo permanece o mesmo)
func (c *CodeGenerator) Close() {
	CloseLogger()
}

func (c *CodeGenerator) Generate(program *ast.Program) llvm.Module {
	defer c.trace("Generate")()

	for _, stmt := range program.Statements {
		c.genStatement(stmt)
	}

	return c.module
}

func isBlockTerminated(block llvm.BasicBlock) bool {
	if block.IsNil() {
		return false
	}
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

func (c *CodeGenerator) pushScope() {
	c.logTrace("=> Entrando em novo escopo")
	c.symbolTable = append(c.symbolTable, make(map[string]SymbolEntry))
}

func (c *CodeGenerator) popScope() {
	c.symbolTable = c.symbolTable[:len(c.symbolTable)-1]
	c.logTrace("<= Saindo do escopo")
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

func (c *CodeGenerator) logTrace(msg string) {
	indent := strings.Repeat("    ", c.indentationLevel)
	logger.Printf("%s%s\n", indent, msg)
}

func (c *CodeGenerator) trace(funcName string) func() {
	c.logTrace(">> " + funcName)
	c.indentationLevel++
	return func() {
		c.indentationLevel--
		c.logTrace("<< " + funcName)
	}
}
