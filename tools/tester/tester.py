import os
import sys
import time
from pathlib import Path
from rich.console import Console
from rich.table import Table

# Adiciona o diretório atual ao path para encontrar os módulos locais
HERE = Path(__file__).parent
sys.path.insert(0, str(HERE))

# Importa as funções dos outros módulos da ferramenta
from builder import build_taquionc
from runner import run_example
from reporter import report_failure, report_summary
from utils import find_examples

# Inicializa o console do Rich para uma saída mais bonita
console = Console()

# Dicionário de testes esperados e seus códigos de saída.
# A ordem dos itens neste dicionário define a sequência de execução dos testes.
EXPECTED = {
    # --- Testes existentes ---
    "start.taq":              200,  # estrutura mínima, ponto de partida
    "const.taq":              10,   # constante simples
    "add.taq":                31,   # operação aritmética básica
    "bool_test.taq":          1,    # teste de booleanos
    "if_statement.taq":       10,   # controle de fluxo básico
    "print_test.taq":         0,    # saída padrão (side-effect)
    "functions_test.taq":     15,   # chamada de função
    "hello_world.taq":        0,    # strings, atribuições
    "advanced_test.taq":      100,  # mistura de várias features

    # --- Novos testes, ordenados por complexidade ---
    "math_precendence.taq":   0,    # Valida a precedência de operadores
    "nested_ifs.taq":         0,    # Valida condicionais aninhadas e escopo
    "multi_return.taq":       0,    # Garante que múltiplos returns em condicionais funcionam
    "function_recursion.taq": 0,    # Testa a pilha de chamadas com recursão
    "scope_shadowing.taq":    0,    # Valida o sombreamento de variáveis (let vs. const)
    "string_concat.taq":      0,    # Testa a concatenação de strings
    "while_loop.taq":         0,    # Implementação de loop 'while'
    "break_continue.taq":     0,    # Controle de fluxo dentro de loops
    "array_basic.taq":        0,    # Suporte básico para arrays (declaração e acesso)
}

def clear_screen():
    """Limpa a tela do console, compatível com Windows, Mac e Linux."""
    os.system('cls' if os.name == 'nt' else 'clear')

def pause():
    """Pausa a execução e espera o usuário pressionar Enter."""
    console.input("\n[yellow]Pressione Enter para voltar ao menu...[/yellow]")

def show_menu():
    """Exibe o menu principal de opções."""
    console.rule("\n[bold blue]🎯 === Taquion Tester === 🎯[/bold blue]")
    table = Table(show_header=False, box=None)
    table.add_row("[cyan]1[/cyan]", "Iniciar testes de todos exemplos")
    table.add_row("[cyan]2[/cyan]", "Compilar TaquionC")
    table.add_row("[cyan]3[/cyan]", "Ver últimos logs de compilação")
    table.add_row("[cyan]4[/cyan]", "Listar arquivos .taq encontrados")
    table.add_row("[cyan]0[/cyan]", "Sair")
    console.print(table)
    return console.input("[bold]👉 Escolha uma opção: [/bold]")

def run_all_tests_menu(compiler_path, examples_path, build_path):
    """
    Executa todos os testes definidos no dicionário EXPECTED na ordem especificada.
    """
    console.rule("[bold blue]Executando todos os testes do Taquion Compiler[/bold blue]")
    start_time = time.time()

    # A compilação de cada teste é tratada pela função run_example.
    # Removida a compilação global para alinhar com o comportamento anterior.
    
    results = {"passed": 0, "failed": 0, "total": len(EXPECTED)}

    for test_file, expected_code in EXPECTED.items():
        console.print(f"\n[cyan]Executando teste:[/cyan] {test_file} (esperado: {expected_code})")

        source_path = Path(examples_path) / test_file
        if not source_path.exists():
            console.print(f"❌ [bold red]FALHOU![/bold red] Arquivo de teste não encontrado: {source_path}")
            results["failed"] += 1
            continue

        return_code, output, _ = run_example(source_path)
        
        passed = (return_code == expected_code)

        if passed:
            console.print(f"✅ [bold green]PASSOU![/bold green] (código de saída: {return_code})")
            results["passed"] += 1
        else:
            console.print(f"❌ [bold red]FALHOU![/bold red]")
            results["failed"] += 1
            report_failure(test_file, expected_code, return_code, output, "N/A (saída combinada)")

    end_time = time.time()
    report_summary(results, end_time - start_time)

def main():
    """Função principal que gerencia o menu e o fluxo do programa."""
    last_build_output = None
    compiler_path = HERE.parent.parent
    examples_path = compiler_path / "examples"
    build_path = compiler_path / "build"

    os.makedirs(build_path, exist_ok=True)

    while True:
        clear_screen()  # Limpa a tela antes de mostrar o menu
        choice = show_menu().strip()

        if choice == '1':
            run_all_tests_menu(str(compiler_path), str(examples_path), str(build_path))
            pause()

        elif choice == '2':
            console.print("\n[bold]⚙️  Compilando TaquionC...[/bold]")
            ok, out = build_taquionc()
            last_build_output = out
            if ok:
                console.print("✅ [green]Compilação concluída com sucesso.[/green]")
            else:
                console.print("❌ [red]Falha na compilação:[/red]\n")
                console.print(out.strip())
            pause()

        elif choice == '3':
            console.print("\n[bold]📄 Últimos logs de compilação:[/bold]\n")
            if last_build_output:
                console.print(last_build_output.strip())
            else:
                console.print("[yellow]⚠️  Nenhuma compilação feita ainda nesta sessão.[/yellow]")
            pause()

        elif choice == '4':
            console.print("\n[bold]📂 Arquivos .taq disponíveis:[/bold]\n")
            try:
                examples = find_examples(examples_path)
                for idx, ex in enumerate(examples, 1):
                    console.print(f"  [cyan]{idx:02d}.[/cyan] {ex.name}")
            except FileNotFoundError:
                console.print(f"[red]Diretório de exemplos não encontrado em: {examples_path}[/red]")
            pause()

        elif choice == '0':
            clear_screen()
            console.print("\n[bold]👋 Saindo...[/bold]")
            break

        else:
            console.print("\n[bold red]❗ Opção inválida.[/bold red]")
            time.sleep(1)

if __name__ == "__main__":
    main()
