#!/usr/bin/env python3
import sys
from pathlib import Path

# adiciona a pasta atual ao path para imports locais
HERE = Path(__file__).parent
sys.path.insert(0, str(HERE))

from utils    import check_deps, find_examples
from builder  import build_taquionc
from runner   import run_example
from reporter import show_summary

def main():
    print("\n🛠️  Taquion Tester — Iniciando testes Taquion\n")

    # 1) checa deps básicas
    missing = check_deps()
    if missing:
        print("❌ Dependências faltando:", ", ".join(missing))
        print("   → Instale clang, go e llvm-config e rode de novo.\n")
        sys.exit(1)
    print("✅ Dependências de sistema OK.\n")

    # 2) encontra .taq
    EX_DIR = HERE.parent.parent / "examples"
    examples = find_examples(EX_DIR)
    if not examples:
        print(f"⚠️  Nenhum .taq encontrado em {EX_DIR}")
        sys.exit(1)

    print("📂 Exemplos encontrados:")
    for ex in examples:
        print("  -", ex.name)
    print()

    # 3) compila taquionc
    ok, out = build_taquionc()
    if not ok:
        print("❌ Falha na compilação:\n", out)
        sys.exit(1)
    print("✅ taquionc compilado com sucesso.\n")

    # 4) executa testes
    results, total = [], 0.0
    for ex in examples:
        ok, output, took = run_example(ex)
        tag = "✔️ OK" if ok else "❌ FAIL"
        print(f"[{ex.name}] {tag} ({took:.2f}s)")
        if not ok:
            print("   >", output.replace("\n", "\n     >"))
        total += took
        results.append((ex.name, ok, took))

    # 5) resumo final
    print("\n")
    show_summary(results)
    print(f"\n⏱️ Tempo total de todos: {total:.2f}s\n")

if __name__ == "__main__":
    main()
