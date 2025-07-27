import os
import subprocess
from pathlib import Path
from utils import get_llvm_flags

BASE_DIR     = Path(__file__).parent.parent.parent.resolve()
BUILD_DIR    = BASE_DIR / "build"
COMPILER_DIR = BASE_DIR / "compiler"
TAQC_SRC     = COMPILER_DIR / "cmd" / "taquionc"
TAQC_BIN     = BUILD_DIR / "taquionc.exe"

def build_taquionc() -> tuple[bool, str]:
    """
    Compila o compilador TaquionC usando Go + LLVM.
    Agora executa no diretório correto (compiler/) que contém o go.mod.
    """
    BUILD_DIR.mkdir(parents=True, exist_ok=True)

    cflags, ldflags = get_llvm_flags()
    env = os.environ.copy()
    env["CGO_CFLAGS"] = cflags
    env["CGO_LDFLAGS"] = ldflags

    cmd = [
        "go", "build",
        "-o", str(TAQC_BIN),
        "-tags=llvm20 byollvm",
        "./cmd/taquionc"
    ]

    result = subprocess.run(
        cmd,
        cwd=COMPILER_DIR,  # AQUI está o go.mod, então tudo vai funcionar
        env=env,
        capture_output=True,
        text=True
    )

    return (result.returncode == 0, result.stdout + result.stderr)
