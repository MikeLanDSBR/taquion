import shutil
import subprocess
from pathlib import Path

def check_executable(name: str) -> bool:
    return shutil.which(name) is not None

def check_deps() -> list[str]:
    deps = ["go", "clang", "llvm-config"]
    return [d for d in deps if not check_executable(d)]

def get_llvm_flags() -> tuple[str, str]:
    cflags = subprocess.check_output(
        ["llvm-config", "--cflags"], text=True
    ).strip()
    ldflags = subprocess.check_output(
        ["llvm-config", "--ldflags", "--libs", "--system-libs"], text=True
    ).strip()
    return cflags, ldflags

def find_examples(examples_dir: Path) -> list[Path]:
    return sorted(examples_dir.glob("*.taq"))
