import subprocess
import time
import os
import platform
import statistics
import shutil
import sys
from pathlib import Path

# Tenta importar a biblioteca 'rich' para uma interface mais bonita.
try:
    from rich.console import Console
    from rich.table import Table
    from rich.rule import Rule
    RICH_AVAILABLE = True
except ImportError:
    RICH_AVAILABLE = False
    class Table:
        def __init__(self, title="", header_style="", **kwargs): self._title = title.replace("[bold magenta]", "").replace("[/bold magenta]", ""); self._rows = []; self._columns = []
        def add_column(self, header, **kwargs): self._columns.append(header)
        def add_row(self, *args): self._rows.append(args)
        def __rich_console__(self, console, options):
            if self._title: print(self._title)
            print(" | ".join(self._columns))
            for row in self._rows: print(" | ".join(map(str, row)))
            yield
    class Console:
        def print(self, *args, **kwargs): print(*args)
        def rule(self, text, **kwargs): print(f"\n--- {text.replace('[bold blue]', '').replace('[/bold blue]', '')}")
        def input(self, prompt=""): return input(prompt.replace('[bold]', '').replace('[/bold]', '').replace('[yellow]', '').replace('[/yellow]', ''))

# --- Configurações ---
BENCHMARK_NAME = "Fibonacci Recursivo (N=42)"
EXPECTED_RESULT = str(267914296)
RUNS = 3
WARMUP = 1

# --- Caminhos ---
TOOL_DIR = Path(__file__).parent.resolve()
BASE_DIR = TOOL_DIR.parent.parent
BUILD_DIR = BASE_DIR / "build"
COMPILER_DIR = BASE_DIR / "compiler"
TAQC_SRC_DIR = COMPILER_DIR / "cmd" / "taquionc"
EXECUTABLE_SUFFIX = ".exe" if platform.system() == "Windows" else ""
TAQC_BIN = BUILD_DIR / f"taquionc{EXECUTABLE_SUFFIX}"
BENCH_DIR = TOOL_DIR / "fibonacci"

# --- Definições das Linguagens ---
LANGUAGES = {
    "taquion": {"src": "benchmark.taquion", "ir": "benchmark.ll", "exe": f"benchmark_taquion{EXECUTABLE_SUFFIX}"},
    "c":       {"src": "benchmark.c", "exe": f"benchmark_c{EXECUTABLE_SUFFIX}"},
    "c++":     {"src": "benchmark.cpp", "exe": f"benchmark_cpp{EXECUTABLE_SUFFIX}"},
    "rust":    {"src": "benchmark.rs", "exe": f"benchmark_rust{EXECUTABLE_SUFFIX}"},
    "go":      {"src": "benchmark.go", "exe": f"benchmark_go{EXECUTABLE_SUFFIX}"},
    "js":      {"src": "benchmark.js"},
    "python":  {"src": "benchmark.py"},
}

console = Console()

# --- Funções Auxiliares (Baseadas no seu código) ---

def run_cmd(cmd, cwd=None, env=None):
    """Executa um comando de forma segura."""
    try:
        result = subprocess.run(cmd, cwd=str(cwd) if cwd else None, env=env,
                                check=True, capture_output=True, text=True, encoding='utf-8')
        return True, result.stdout, result.stderr
    except FileNotFoundError:
        return False, "", f"Comando não encontrado: {cmd[0]}"
    except subprocess.CalledProcessError as e:
        return False, e.stdout, e.stderr

def get_llvm_flags():
    """Busca as flags de compilação do LLVM no sistema."""
    try:
        cflags = subprocess.check_output(["llvm-config", "--cflags"], text=True).strip()
        ldflags = subprocess.check_output(["llvm-config", "--ldflags", "--libs", "--system-libs"], text=True).strip()
        return cflags, ldflags
    except Exception: return None, None

def check_deps():
    """Verifica se as dependências essenciais estão instaladas."""
    deps = ["go", "clang", "llvm-config", "node", "python"]
    return [d for d in deps if not shutil.which(d)]

def clear():
    """Limpa a tela do console."""
    os.system('cls' if os.name == 'nt' else 'clear')
    
# --- Lógica Principal da Ferramenta ---

def build_taquionc():
    """Compila o TaquionC com as flags do LLVM."""
    console.rule("[bold blue]→ Compilando o Compilador TaquionC[/bold blue]")
    cflags, ldflags = get_llvm_flags()
    if not cflags:
        console.print("[bold red]ERRO: Não foi possível obter as flags do LLVM.[/bold red]")
        return False
    
    BUILD_DIR.mkdir(parents=True, exist_ok=True)
    env = os.environ.copy()
    env["CGO_CFLAGS"] = cflags
    env["CGO_LDFLAGS"] = ldflags
    
    console.print(f"Construindo [cyan]taquionc[/cyan] em [yellow]{TAQC_BIN}[/yellow]...")
    cmd = ["go", "build", "-o", str(TAQC_BIN), "-tags=llvm20 byollvm", str(TAQC_SRC_DIR)]
    
    ok, out, err = run_cmd(cmd, cwd=COMPILER_DIR, env=env)
    if not ok:
        console.print("[bold red]FALHA[/bold red]")
        console.print(f"[red]{out}{err}[/red]")
        return False
    
    console.print("[bold green]OK[/bold green]")
    return True
    
def compile_benchmarks():
    """Compila todos os benchmarks necessários."""
    console.rule(f"[bold blue]→ Compilando Benchmarks[/bold blue]")
    built = set()
    for lang, info in LANGUAGES.items():
        if 'exe' not in info: continue
        
        console.print(f"Verificando [cyan]{lang:<8}[/cyan]...", end="")
        src = BENCH_DIR / info['src']
        exe = BENCH_DIR / info['exe']

        if exe.exists():
            console.print("[yellow] já compilado, pulando.[/yellow]")
            built.add(lang)
            continue
        
        console.print("[magenta] compilando...[/magenta]", end="")

        steps = []
        if lang == 'taquion':
            ir_file = BENCH_DIR / info['ir']
            steps.append({'cmd': [str(TAQC_BIN), str(src), '-o', str(ir_file)], 'cwd': COMPILER_DIR})
            steps.append({'cmd': ['clang','-O3', str(ir_file), '-o', str(exe)], 'cwd': BASE_DIR})
        else:
            compiler_map = {'c': 'clang', 'cpp': 'clang++', 'rust': 'rustc', 'go': 'go build'}
            opts = ['-O3'] if lang in ['c', 'cpp'] else ['-C', 'opt-level=3'] if lang == 'rust' else []
            cmd = [compiler_map[lang]] + opts + ['-o', str(exe), str(src)]
            if lang == 'go':
                cmd = ['go', 'build', '-o', str(exe), str(src)]
            steps.append({'cmd': cmd, 'cwd': BASE_DIR})

        lang_ok = True
        for step in steps:
            ok, out, err = run_cmd(step['cmd'], cwd=step['cwd'])
            if not ok:
                console.print(f"\n[bold red]ERRO ao compilar {lang}[/bold red]")
                console.print(f"[red]{out}{err}[/red]")
                lang_ok = False
                break
        
        if lang_ok:
            built.add(lang)
            console.print(f"\rVerificando [cyan]{lang:<8}[/cyan]... [bold green]OK[/bold green]              ")

    return built

def run_benchmarks(built):
    """Executa os benchmarks e mede o tempo."""
    console.rule(f"[bold blue]→ Executando {BENCHMARK_NAME}[/bold blue]")
    results = {}
    for lang, info in LANGUAGES.items():
        console.print(f"Executando [cyan]{lang:<8}[/cyan]...", end="")
        
        if 'exe' in info and lang not in built:
            console.print("[yellow] pulando (falha na compilação).[/yellow]")
            results[lang] = {'status': 'COMPILAÇÃO FALHOU', 'avg_time': float('inf')}
            continue
        
        # --- LÓGICA CORRIGIDA PARA CONSTRUÇÃO DE COMANDOS ---
        cmd = []
        if 'exe' in info:
            # Para linguagens compiladas, o comando é o caminho para o executável.
            cmd = [str(BENCH_DIR / info['exe'])]
        elif lang == 'js':
            # Para Node.js, usamos 'node' e o caminho para o script 
            # como dois itens separados na lista.
            cmd = ['node', str(BENCH_DIR / info['src'])]
        elif lang == 'python':
            # Para Python, usamos sys.executable (o caminho completo para o python.exe)
            # e o caminho para o script como dois itens separados na lista.
            cmd = [sys.executable, str(BENCH_DIR / info['src'])]

        times = []
        ok, err_output = True, ""
        for i in range(RUNS + WARMUP):
            start = time.perf_counter()
            ok_run, out, err = run_cmd(cmd) # A função run_cmd agora recebe a lista correta
            elapsed = time.perf_counter() - start
            
            if not ok_run or EXPECTED_RESULT not in out:
                ok = False
                err_output = err if err else f"Saída inesperada: {out.strip()}"
                break
            if i >= WARMUP:
                times.append(elapsed)
        
        if ok:
            avg = statistics.mean(times)
            console.print(f"[bold green]OK[/bold green] (média: [yellow]{avg:.4f}s[/yellow])")
            results[lang] = {'status': 'OK', 'avg_time': avg}
        else:
            console.print(f"[bold red]FALHOU[/bold red]")
            console.print(f"[red]{err_output}[/red]")
            results[lang] = {'status': 'ERRO', 'avg_time': float('inf')}
    return results

def print_results(results):
    """Exibe os resultados finais em uma tabela formatada."""
    console.rule("[bold blue]→ Resultados Finais do Benchmark[/bold blue]")
    sorted_langs = sorted(results.items(), key=lambda x: x[1]['avg_time'])
    
    ok_results = [res for res in sorted_langs if res[1]['status'] == 'OK']
    if not ok_results:
        console.print("[yellow]Nenhum benchmark foi concluído com sucesso.[/yellow]")
        return
        
    baseline_time = ok_results[0][1]['avg_time']
    baseline_lang = ok_results[0][0]

    table = Table(title=f"{BENCHMARK_NAME} | Média de {RUNS} execuções", header_style="bold magenta")
    table.add_column("Posição", style="cyan", justify="center"); table.add_column("Linguagem", style="bold"); table.add_column("Tempo Médio (s)", style="yellow", justify="right"); table.add_column("Relativo ao Melhor", style="green", justify="left"); table.add_column("Status", justify="center")
    
    for i, (lang, data) in enumerate(sorted_langs):
        status, status_style = data['status'], "green"
        if status == "ERRO": status_style = "red"
        elif status == "COMPILAÇÃO FALHOU": status_style = "yellow"
        
        status_colored = f"[bold {status_style}]{status}[/bold {status_style}]" if RICH_AVAILABLE else status
        
        if data['status'] == "OK":
            time_str = f"{data['avg_time']:.4f}"
            relative = data['avg_time'] / baseline_time
            relative_str = "O Mais Rápido" if lang == baseline_lang else f"{relative:.2f}x mais lento"
        else:
            time_str, relative_str = ("-" * 10), "N/A"
        
        table.add_row(f"{i+1}", lang.capitalize(), time_str, relative_str, status_colored)
    console.print(table)

def main_menu():
    """Menu principal da ferramenta."""
    miss = check_deps()
    if miss:
        console.print(f"[bold red]Dependências faltando:[/bold red] {', '.join(miss)}")
        return

    while True:
        clear()
        console.rule("\n[bold blue]🚀 === Ferramenta de Benchmark Taquion === 🚀[/bold blue]")
        table = Table(show_header=False, box=None)
        table.add_row("[cyan]1[/cyan]", "Executar todos os benchmarks")
        table.add_row("[cyan]2[/cyan]", "Apenas compilar benchmarks")
        table.add_row("[cyan]3[/cyan]", "Recompilar o compilador TaquionC")
        table.add_row("[cyan]0[/cyan]", "Sair")
        console.print(table)
        choice = console.input("[bold]👉 Escolha uma opção: [/bold]").strip()

        if choice == '1':
            if not TAQC_BIN.exists():
                console.print(f"[bold yellow]Compilador TaquionC não encontrado. Tentando compilar agora...[/bold yellow]")
                if not build_taquionc():
                    console.input("\n[yellow]Pressione Enter para voltar ao menu...[/yellow]")
                    continue
            
            built_langs = compile_benchmarks()
            results = run_benchmarks(built_langs)
            print_results(results)

        elif choice == '2':
            compile_benchmarks()

        elif choice == '3':
            build_taquionc()
        
        elif choice == '0':
            console.print("\n[bold]👋 Saindo...[/bold]")
            break
        
        else:
            console.print("[bold red]❗ Opção inválida.[/bold red]")
            time.sleep(1)
        
        console.input("\n[yellow]Pressione Enter para continuar...[/yellow]")

if __name__ == '__main__':
    try:
        if not RICH_AVAILABLE:
            print("AVISO: Biblioteca 'rich' não encontrada. Para uma visualização mais bonita, instale com: pip install rich")
        main_menu()
    except KeyboardInterrupt:
        console.print("\n\n[bold yellow]Execução interrompida pelo usuário.[/bold yellow]")
    except Exception as e:
        console.print(f"\n[bold red]Ocorreu um erro inesperado na ferramenta:[/bold red]")
        console.print(e)
    finally:
        # Garante que a janela não feche sozinha se não for um terminal interativo
        if not sys.stdout.isatty():
             input("\nExecução finalizada. Pressione Enter para fechar.")