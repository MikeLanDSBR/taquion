import os
import subprocess
from pathlib import Path
from utils import get_llvm_flags

# raiz do projeto: taquion/ (3 níveis acima desta pasta)
BASE_DIR = Path(__file__).parent.parent.parent
BUILD_DIR = BASE_DIR / "build"
TAQC_SRC  = BASE_DIR / "cmd" / "taquionc"
TAQC_BIN  = BUILD_DIR / "taquionc.exe"

def build_taquionc() -> tuple[bool, str]:
    BUILD_DIR.mkdir(parents=True, exist_ok=True)
    cflags, ldflags = get_llvm_flags()
    env = os.environ.copy()
    env["CGO_CFLAGS"]  = cflags
    env["CGO_LDFLAGS"] = ldflags

    result = subprocess.run(
        ["go", "build", "-o", str(TAQC_BIN),
         "-tags=llvm20 byollvm", str(TAQC_SRC)],
        env=env, capture_output=True, text=True
    )
    return (result.returncode == 0, result.stdout + result.stderr)
