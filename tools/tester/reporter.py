try:
    from rich.console import Console
    from rich.table import Table
    console = Console()
    def show_summary(results: list[tuple[str, bool, float]]) -> None:
        table = Table(title="Resumo de Testes Taquion")
        table.add_column("Exemplo")
        table.add_column("Status", justify="center")
        table.add_column("Tempo (s)", justify="right")
        for name, ok, t in results:
            status = "[green]OK[/]" if ok else "[red]FAIL[/]"
            table.add_row(name, status, f"{t:.2f}")
        console.print(table)

except ModuleNotFoundError:
    # Fallback simples se Rich não estiver instalado
    def show_summary(results: list[tuple[str, bool, float]]) -> None:
        print("=== Resumo de Testes Taquion ===")
        for name, ok, t in results:
            status = "OK" if ok else "FAIL"
            print(f"{name:20} | {status:4} | {t:.2f}s")
