# runner.py
import time
import subprocess
from pathlib import Path

# Caminhos
BASE_DIR     = Path(__file__).parent.parent.parent       # …/taquion
COMPILER_DIR = BASE_DIR / "compiler"                     # …/taquion/compiler
LOG_DIR      = COMPILER_DIR / "log"
LEXER_LOG    = LOG_DIR / "lexer.log"
PARSER_LOG   = LOG_DIR / "parser.log"
CODEGEN_LOG  = LOG_DIR / "codegen.log"
BUILD_DIR    = BASE_DIR / "build"
TAQC_BIN     = BUILD_DIR / "taquionc.exe"

def run_example(example: Path) -> tuple[int, str, float]:
    """
    Compila e executa um exemplo .taq (ou 'main').

    Agora:
      - O binário final (.exe) é salvo NA MESMA PASTA do exemplo,
        com o nome da PASTA: <nome_da_pasta>.exe
    """
    # Limpa logs antigos
    if LOG_DIR.exists():
        for f in (LEXER_LOG, PARSER_LOG, CODEGEN_LOG):
            try:
                f.unlink()
            except FileNotFoundError:
                pass

    # Saídas intermediárias continuam na build/
    ir_file   = BUILD_DIR / "output.ll"

    # Onde salvar o .exe final
    example_dir  = example.parent
    exe_file     = example_dir / f"{example_dir.name}.exe"

    # Garante que a pasta do exemplo existe (deveria existir)
    example_dir.mkdir(parents=True, exist_ok=True)

    # Remove o exe antigo, se existir (evita confusão)
    if exe_file.exists():
        try:
            exe_file.unlink()
        except PermissionError:
            # No Windows pode estar "em uso"; tenta renomear
            exe_file.rename(exe_file.with_suffix(".old"))

    start = time.time()

    # 1) Gerar LLVM IR
    p1 = subprocess.run(
        [str(TAQC_BIN), str(example), "-o", str(ir_file)],
        cwd=COMPILER_DIR, capture_output=True, text=True
    )
    if p1.returncode != 0:
        return p1.returncode, p1.stderr, time.time() - start

    # 2) Compilar IR -> exe (no diretório do exemplo, com nome da pasta)
    p2 = subprocess.run(
        ["clang", str(ir_file), "-o", str(exe_file)],
        cwd=BASE_DIR, capture_output=True, text=True
    )
    if p2.returncode != 0:
        return p2.returncode, p2.stderr, time.time() - start

    # 3) Executar o .exe
    p3 = subprocess.run(
        [str(exe_file)],
        cwd=example_dir, capture_output=True, text=True
    )
    duration = time.time() - start

    # 4) Lê logs gerados
    log_content = ""
    for log_file in (LEXER_LOG, PARSER_LOG):
        if log_file.exists():
            log_content += f"\n=== {log_file.name} ===\n"
            log_content += log_file.read_text()

    output = (p3.stdout or p3.stderr) + log_content
    return p3.returncode, output, duration
