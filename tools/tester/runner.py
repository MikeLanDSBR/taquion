import time
import subprocess
from pathlib import Path
from builder import TAQC_BIN

# raiz do projeto: taquion/
BASE_DIR = Path(__file__).parent.parent.parent
BUILD_DIR = BASE_DIR / "build"

def run_example(example: Path) -> tuple[bool, str, float]:
    ir_file  = BUILD_DIR / f"{example.stem}.ll"
    exe_file = BUILD_DIR / f"{example.stem}.exe"
    start = time.time()

    # 1) Gera IR
    ir_res = subprocess.run(
        [str(TAQC_BIN), str(example), "-o", str(ir_file)],
        capture_output=True, text=True
    )
    if ir_res.returncode != 0:
        return False, ir_res.stderr, time.time() - start

    # 2) Compila com clang
    clang_res = subprocess.run(
        ["clang", str(ir_file), "-o", str(exe_file)],
        capture_output=True, text=True
    )
    if clang_res.returncode != 0:
        return False, clang_res.stderr, time.time() - start

    # 3) Executa
    run_res = subprocess.run(
        [str(exe_file)], capture_output=True, text=True
    )
    duration = time.time() - start
    return (run_res.returncode == 0, run_res.stdout or run_res.stderr, duration)
