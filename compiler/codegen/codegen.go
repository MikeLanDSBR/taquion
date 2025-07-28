package codegen

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"taquion/compiler/ast"

	"github.com/MikeLanDSBR/go-llvm"
)

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
		logFile.Close()
	}
}

type SymbolEntry struct {
	Value     llvm.Value
	Ptr       llvm.Value
	Typ       llvm.Type
	TypeName  string
	ArrayType llvm.Type
	IsLiteral bool
}

type CodeGenerator struct {
	module                    llvm.Module
	builder                   llvm.Builder
	context                   llvm.Context
	symbolTable               []map[string]SymbolEntry
	indentationLevel          int
	currentFunctionReturnType llvm.Type

	printfFunc     llvm.Value
	printfFuncType llvm.Type
	mallocFunc     llvm.Value
	strlenFunc     llvm.Value
	strcpyFunc     llvm.Value
	strcatFunc     llvm.Value

	loopCondBlock llvm.BasicBlock
	loopEndBlock  llvm.BasicBlock

	structTypes        map[string]llvm.Type
	structFieldIndices map[string]map[string]int
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
	cg.declareCFunctions()
	cg.structTypes = make(map[string]llvm.Type)
	cg.structFieldIndices = make(map[string]map[string]int)
	logger.Println("Nova instância de CodeGenerator criada.")
	return cg
}

func (c *CodeGenerator) Close() {
	CloseLogger()
}

func (c *CodeGenerator) Generate(program *ast.Program) llvm.Module {
	defer c.trace("Generate")()
	for _, stmt := range program.Statements {
		c.genStatement(stmt)
	}
	mainFunc := c.module.NamedFunction("main")
	if mainFunc.IsNil() {
		mainFuncType := llvm.FunctionType(c.context.Int32Type(), []llvm.Type{}, false)
		mainFunc = llvm.AddFunction(c.module, "main", mainFuncType)
		entryBlock := c.context.AddBasicBlock(mainFunc, "entry")
		tempBuilder := c.context.NewBuilder()
		defer tempBuilder.Dispose()
		tempBuilder.SetInsertPointAtEnd(entryBlock)
		tempBuilder.CreateRet(llvm.ConstInt(c.context.Int32Type(), 0, false))
	}
	return c.module
}

func (c *CodeGenerator) GetValueTypeSafe(val llvm.Value) llvm.Type {
	if val.IsNil() {
		return llvm.Type{}
	}
	return val.Type()
}

func isBlockTerminated(block llvm.BasicBlock) bool {
	if block.IsNil() || block.LastInstruction().IsNil() {
		return false
	}
	switch block.LastInstruction().InstructionOpcode() {
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
	arrayTypeStr := "nil"
	if !entry.ArrayType.IsNil() {
		arrayTypeStr = entry.ArrayType.String()
	}
	typStr := "nil"
	if !entry.Typ.IsNil() {
		typStr = entry.Typ.String()
	}

	c.logTrace(fmt.Sprintf(
		"Definindo símbolo '%s' no escopo atual. IsLiteral: %t, Ptr: %v, Value: %v, Typ: %s, ArrayType: %s",
		name, entry.IsLiteral, entry.Ptr, entry.Value, typStr, arrayTypeStr,
	))
	c.symbolTable[len(c.symbolTable)-1][name] = entry
}

func (c *CodeGenerator) getSymbol(name string) (SymbolEntry, bool) {
	c.logTrace(fmt.Sprintf("Procurando símbolo '%s'", name))
	for i := len(c.symbolTable) - 1; i >= 0; i-- {
		if entry, ok := c.symbolTable[i][name]; ok {
			arrayTypeStr := "nil"
			if !entry.ArrayType.IsNil() {
				arrayTypeStr = entry.ArrayType.String()
			}
			typStr := "nil"
			if !entry.Typ.IsNil() {
				typStr = entry.Typ.String()
			}
			c.logTrace(fmt.Sprintf(
				"Símbolo '%s' encontrado no escopo %d. IsLiteral: %t, Ptr: %v, Value: %v, Typ: %s, ArrayType: %s",
				name, i, entry.IsLiteral, entry.Ptr, entry.Value, typStr, arrayTypeStr,
			))
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
