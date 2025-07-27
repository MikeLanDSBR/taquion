import os
import sys
import time
from pathlib import Path
from rich.console import Console
from rich.table import Table

# Setup de paths
HERE = Path(__file__).parent
sys.path.insert(0, str(HERE))

from builder import build_taquionc
from runner import run_example
from reporter import report_failure, report_summary
from utils import find_examples

console = Console()

# Nomes de pastas dentro de /examples e seus códigos esperados
EXPECTED = {
    "start":              200,
    "const":              10,
    "add":                31,
    "bool_test":          1,
    "if_statement":       10,
    "print_test":         0,
    "functions_test":     15,
    "hello_world":        0,
    "advanced_test":      100,
    "math_precendence":   0,
    "nested_ifs":         0,
    "multi_return":       0,
    "function_recursion": 0,
    "scope_shadowing":    0,
    "string_concat":      0,
    "while_loop":         0,
    "break_continue":     0,
    "array_basic":        0,
}

def clear_screen():
    os.system('cls' if os.name == 'nt' else 'clear')

def pause():
    console.input("\n[yellow]Pressione Enter para voltar ao menu...[/yellow]")

def show_menu():
    console.rule("[bold blue]🎯 === Taquion Tester === 🎯[/bold blue]")
    table = Table(show_header=False, box=None)
    table.add_row("[cyan]1[/cyan]", "Iniciar testes de todos exemplos")
    table.add_row("[cyan]2[/cyan]", "Compilar TaquionC")
    table.add_row("[cyan]3[/cyan]", "Ver últimos logs de compilação")
    table.add_row("[cyan]4[/cyan]", "Listar exemplos encontrados")
    table.add_row("[cyan]0[/cyan]", "Sair")
    console.print(table)
    return console.input("[bold]👉 Escolha uma opção: [/bold]")

def run_all_tests_menu(examples_path):
    console.rule("[bold blue]Executando todos os testes do Taquion Compiler[/bold blue]")
    start_time = time.time()
    results = {"passed": 0, "failed": 0, "total": len(EXPECTED)}

    for test_name, expected_code in EXPECTED.items():
        main_path = Path(examples_path) / test_name / "main.taq"
        if not main_path.exists():
            main_path = Path(examples_path) / test_name / "main"
        if not main_path.exists():
            console.print(f"❌ [bold red]FALHOU![/bold red] Arquivo não encontrado: {main_path}")
            results["failed"] += 1
            continue

        console.print(f"\n[cyan]Executando:[/cyan] {test_name} (esperado: {expected_code})")
        return_code, output, _ = run_example(main_path)

        if return_code == expected_code:
            console.print(f"✅ [bold green]PASSOU![/bold green] (código: {return_code})")
            results["passed"] += 1
        else:
            console.print(f"❌ [bold red]FALHOU![/bold red]")
            results["failed"] += 1
            report_failure(test_name, expected_code, return_code, output, "N/A")

    end_time = time.time()
    report_summary(results, end_time - start_time)

def main():
    last_build_output = None
    compiler_path = HERE.parent.parent
    examples_path = compiler_path / "examples"
    build_path = compiler_path / "build"
    os.makedirs(build_path, exist_ok=True)

    while True:
        clear_screen()
        choice = show_menu().strip()

        if choice == '1':
            run_all_tests_menu(examples_path)
            pause()

        elif choice == '2':
            console.print("\n[bold]⚙️  Compilando TaquionC...[/bold]")
            ok, out = build_taquionc()
            last_build_output = out
            console.print("✅ [green]Compilado com sucesso.[/green]" if ok else f"❌ [red]Falha:[/red]\n{out.strip()}")
            pause()

        elif choice == '3':
            console.print("\n[bold]📄 Logs da última compilação:[/bold]\n")
            console.print(last_build_output.strip() if last_build_output else "[yellow]⚠️ Nenhuma compilação feita ainda.[/yellow]")
            pause()

        elif choice == '4':
            console.print("\n[bold]📂 Exemplos encontrados:[/bold]\n")
            for idx, ex in enumerate(find_examples(examples_path), 1):
                console.print(f"[cyan]{idx:02d}.[/cyan] {ex.parent.name}/ → {ex.name}")
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
