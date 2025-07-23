// compiler/codegen/codegen.go

package codegen

import (
	"strconv"

	"taquion/compiler/ast"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

// Generate traduz a AST de um programa Taquion para um módulo LLVM IR.
func Generate(program *ast.Program) *ir.Module {
	module := ir.NewModule()
	mainFunc := module.NewFunc("main", types.I32)
	entryBlock := mainFunc.NewBlock("entry")

	// CORREÇÃO: Itera nas declarações para encontrar a FUNÇÃO "main".
	for _, stmt := range program.Statements {
		if funcDecl, ok := stmt.(*ast.FunctionDeclaration); ok && funcDecl.Name.Value == "main" {

			// Agora, itera dentro do CORPO da função encontrada.
			for _, bodyStmt := range funcDecl.Body.Statements {
				// Procuramos pela declaração de retorno DENTRO do corpo.
				if returnStmt, ok := bodyStmt.(*ast.ReturnStatement); ok {

					// Verificamos se o valor de retorno é um número inteiro.
					if intLit, ok := returnStmt.ReturnValue.(*ast.IntegerLiteral); ok {
						val, _ := strconv.ParseInt(intLit.TokenLiteral(), 10, 32)
						returnValue := constant.NewInt(types.I32, val)

						// Gera a instrução de retorno que estava faltando!
						entryBlock.NewRet(returnValue)
						break // Para o loop interno, já que encontramos o return.
					}
				}
			}
			break // Para o loop externo, já que encontramos e processamos a função main.
		}
	}

	// MEDIDA DE SEGURANÇA:
	// Se, por algum motivo, o bloco `entry` ainda não tiver um terminador
	// (ex: função main vazia), adicionamos um `return 0` padrão para garantir um IR válido.
	if entryBlock.Term == nil {
		defaultReturn := constant.NewInt(types.I32, 0)
		entryBlock.NewRet(defaultReturn)
	}

	return module
}
