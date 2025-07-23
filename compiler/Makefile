LLVM_CFLAGS := $(shell llvm-config --cflags)
LLVM_LDFLAGS := $(shell llvm-config --ldflags --libs --system-libs)
TAQ_SRC := ../examples/add.taq
LLVM_IR := output.ll
OUT_EXE := add_test.exe

.PHONY: run clean build test

run: build $(OUT_EXE)
	@echo "==> Executando $(OUT_EXE)..."
	@./$(OUT_EXE)
	@echo "Código de saída: $$?"

build:
	@echo "==> Compilando Taquion com LLVM..."
	@mkdir -p log
	CGO_CFLAGS="$(LLVM_CFLAGS)" CGO_LDFLAGS="$(LLVM_LDFLAGS)" go build -o taquionc.exe -tags="llvm20 byollvm" ./cmd/taquionc

$(LLVM_IR): build
	@echo "==> Gerando LLVM IR com taquionc..."
	./taquionc.exe $(TAQ_SRC)

$(OUT_EXE): $(LLVM_IR)
	@echo "==> Compilando IR com clang..."
	clang $(LLVM_IR) -o $(OUT_EXE)

clean:
	rm -f taquionc.exe $(LLVM_IR) $(OUT_EXE)
	rm -rf log
