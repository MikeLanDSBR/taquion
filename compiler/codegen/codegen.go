// Arquivo: codegen/codegen.go
// Função: Ponto de entrada e funções principais do gerador de código.
package codegen

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync" // Adicionado para inicialização segura do logger
	"taquion/compiler/ast"

	"tinygo.org/x/go-llvm"
)

// --- LOGGER GLOBAL (PADRÃO RECOMENDADO) ---
var (
	logger   *log.Logger
	logFile  *os.File
	initOnce sync.Once
)

// initLogger inicializa o logger global de forma segura usando sync.Once.
// Isso garante que a inicialização ocorra apenas uma vez, mesmo em ambientes concorrentes.
func initLogger() {
	initOnce.Do(func() {
		// Garante que o diretório de log exista
		if err := os.MkdirAll("log", 0755); err != nil {
			log.Fatalf("Erro ao criar diretório de log: %v", err)
		}
		var err error
		// Abre o arquivo de log
		logFile, err = os.OpenFile("log/codegen.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalf("Erro ao abrir o arquivo de log codegen.log: %v", err)
		}
		// Cria a instância do logger
		logger = log.New(logFile, "CODEGEN: ", log.LstdFlags)
		logger.Println("=== Nova sessão de log do codegen iniciada ===")
	})
}

// CloseLogger deve ser chamada (geralmente com defer) na função main do seu programa
// para garantir que o arquivo de log seja fechado corretamente.
func CloseLogger() {
	if logFile != nil {
		logger.Println("=== Encerrando sessão de log do codegen ===")
		logFile.Close()
	}
}

// SymbolEntry armazena o ponteiro e o tipo de uma variável na tabela de símbolos.
type SymbolEntry struct {
	Ptr llvm.Value
	Typ llvm.Type
}

// CodeGenerator mantém o estado durante a geração de código.
type CodeGenerator struct {
	module           llvm.Module
	builder          llvm.Builder
	context          llvm.Context
	symbolTable      []map[string]SymbolEntry // Pilha de mapas para gerenciar escopo com tipos.
	indentationLevel int
}

// NewCodeGenerator cria uma nova instância do gerador de código.
func NewCodeGenerator() *CodeGenerator {
	initLogger() // Garante que o logger global esteja inicializado

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

// Close é um método de conveniência que chama a função CloseLogger do pacote.
// Isso restaura a compatibilidade com o código que chama `defer generator.Close()`.
func (c *CodeGenerator) Close() {
	CloseLogger()
}

// Generate é o ponto de entrada que traduz a AST para um módulo LLVM.
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

// --- FUNÇÕES DE ESCOPO ---

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

// --- FUNÇÕES DE LOGGING AUXILIARES ---

// logTrace usa o logger global para registrar mensagens com indentação.
func (c *CodeGenerator) logTrace(msg string) {
	indent := strings.Repeat("    ", c.indentationLevel)
	logger.Printf("%s%s\n", indent, msg)
}

// trace é uma função auxiliar para rastrear a entrada e saída de funções.
func (c *CodeGenerator) trace(funcName string) func() {
	c.logTrace(">> " + funcName)
	c.indentationLevel++
	return func() {
		c.indentationLevel--
		c.logTrace("<< " + funcName)
	}
}
