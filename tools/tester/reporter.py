from rich.console import Console
from rich.table import Table
from rich.panel import Panel
from rich.syntax import Syntax

# Inicializa o console do Rich
console = Console()

def report_failure(test_file, expected, actual, stdout, stderr):
    """
    Exibe um relatório detalhado para um teste que falhou.
    """
    panel_content = ""
    if stdout:
        panel_content += "[bold]Saída (stdout):[/bold]\n"
        panel_content += stdout.strip() + "\n\n"
    if stderr:
        panel_content += "[bold]Erros (stderr):[/bold]\n"
        panel_content += stderr.strip()

    console.print(
        Panel(
            panel_content,
            title=f"[bold red]Detalhes da Falha: {test_file}[/bold red]",
            border_style="red",
            subtitle=f"Esperado: [yellow]{expected}[/yellow] | Recebido: [red]{actual}[/red]",
            expand=False
        )
    )

def report_summary(results, duration):
    """
    Exibe uma tabela com o resumo dos resultados dos testes.
    """
    console.rule("[bold blue]Resumo dos Testes[/bold blue]")

    # Define a cor do resultado geral
    if results["failed"] > 0:
        status_style = "bold red"
        status_text = "FALHOU"
    else:
        status_style = "bold green"
        status_text = "PASSOU"

    # Cria a tabela de resumo
    summary_table = Table(box=None, show_header=False)
    summary_table.add_column(style="bold")
    summary_table.add_column()

    summary_table.add_row("Resultado:", f"[{status_style}]{status_text}[/{status_style}]")
    summary_table.add_row("Total de testes:", str(results['total']))
    summary_table.add_row("[green]Passaram[/green]:", str(results['passed']))
    summary_table.add_row("[red]Falharam[/red]:", str(results['failed']))
    summary_table.add_row("Duração total:", f"{duration:.2f} segundos")

    console.print(summary_table)

