#!/usr/bin/env python3
import sys
import time
from pathlib import Path

HERE = Path(__file__).parent
sys.path.insert(0, str(HERE))

from utils import check_deps, find_examples
from builder import build_taquionc
from runner import run_example

EXPECTED = {
    "start.taq":              200,  # estrutura mínima, ponto de partida
    "const.taq":              10,   # constante simples
    "add.taq":                31,   # operação aritmética básica
    "bool_test.taq":          0,    # teste de booleanos
    "if_statement.taq":       10,   # controle de fluxo básico
    "print_test.taq":         0,    # saída padrão (side-effect)
    "functions_test.taq":     15,   # chamada de função
    "hello_world.taq":        0,    # strings, atribuições
    "advanced_test.taq":      100,  # mistura de várias features
}

def pause():
    input("\n⏎ Pressione Enter para voltar ao menu... ")

def menu():
    print("\n🎯 === Taquion Tester === 🎯")
    print("1 - Iniciar testes de todos exemplos")
    print("2 - Compilar TaquionC")
    print("3 - Ver últimos logs de compilação")
    print("4 - Listar arquivos .taq encontrados")
    print("0 - Sair")
    return input("\n👉 Escolha uma opção: ")

def format_result(name, status, rc, expected, took):
    name_str     = f"[{name}]".ljust(35)
    status_str   = status.ljust(6)
    rc_str       = str(rc).rjust(5)
    expected_str = str(expected).rjust(6) if expected is not None else " None "
    time_str     = f"{took:.2f}s".rjust(7)
    return f"{name_str}{status_str} (⏎ {rc_str} | 🎯 {expected_str}) ⏱️ {time_str}"

def main():
    last_build_output = None
    PROJECT_ROOT = HERE.parent.parent
    EX_DIR = PROJECT_ROOT / "examples"

    while True:
        escolha = menu().strip()

        if escolha == '1':
            print("🔍 Verificando dependências...")
            missing = check_deps()
            if missing:
                print("🚫 Dependências faltando:", ", ".join(missing))
                pause()
                continue
            print("✅ Dependências OK.")

            examples = find_examples(EX_DIR)
            if not examples:
                print("⚠️  Nenhum .taq encontrado em examples/")
                pause()
                continue

            total = 0.0
            for ex in examples:
                name = ex.name
                expected = EXPECTED.get(name, None)
                rc, output, took = run_example(ex)

                if expected is None:
                    passed = rc != 0
                else:
                    passed = (rc == expected)

                status = "✅ OK" if passed else "❌ FAIL"
                print(format_result(name, status, rc, expected, took))

            print(f"\n⏱️ Tempo total: {total:.2f}s")
            pause()

        elif escolha == '2':
            print("⚙️  Compilando TaquionC...")
            ok, out = build_taquionc()
            last_build_output = out
            if ok:
                print("✅ Compilação concluída com sucesso.")
            else:
                print("❌ Falha na compilação:\n")
                print(out.strip())
            pause()

        elif escolha == '3':
            print("📄 Últimos logs de compilação:\n")
            if last_build_output:
                print(last_build_output.strip())
            else:
                print("⚠️  Nenhuma compilação feita ainda.")
            pause()

        elif escolha == '4':
            print("📂 Arquivos .taq disponíveis:\n")
            examples = find_examples(EX_DIR)
            for idx, ex in enumerate(examples, 1):
                print(f"  {idx:02d}. {ex.name}")
            pause()

        elif escolha == '0':
            print("👋 Saindo...")
            break

        else:
            print("❗ Opção inválida.")
            pause()

if __name__ == "__main__":
    main()
