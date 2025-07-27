# Variáveis de configuração do LLVM
LLVM_CFLAGS := $(shell llvm-config --cflags)
LLVM_LDFLAGS := $(shell llvm-config --ldflags --libs --system-libs)
EXAMPLE_NAME ?= type
OUT_NAME := $(EXAMPLE_NAME)
TAQ_SRC := ../examples/$(EXAMPLE_NAME)/main.taq
BUILD_DIR := ../build
TAQC_BIN := $(BUILD_DIR)/taquionc.exe
LLVM_IR := $(BUILD_DIR)/$(OUT_NAME).ll
OUT_EXE := $(BUILD_DIR)/$(OUT_NAME).exe

.PHONY: all run clean build taqc

# Regra padrão: compila tudo
all: build taqc $(OUT_EXE)

# Compila o taquionc
taqc:
	@echo "==> Compilando taquionc atualizado..."
	@mkdir -p $(BUILD_DIR)
	CGO_CFLAGS="$(LLVM_CFLAGS)" \
	CGO_LDFLAGS="$(LLVM_LDFLAGS)" \
	go build -o $(TAQC_BIN) -tags="llvm20 byollvm" ./cmd/taquionc

# Gera IR e compila o executável
$(OUT_EXE): taqc
	@echo "==> Gerando LLVM IR: $(LLVM_IR) a partir de $(TAQ_SRC)"
	$(TAQC_BIN) $(TAQ_SRC) -o $(LLVM_IR)
	@echo "==> Compilando IR com clang: $(OUT_EXE)"
	clang $(LLVM_IR) -o $(OUT_EXE)

# Executa o programa compilado
run: all
	@echo "==> Executando $(OUT_EXE)..."
	@$(OUT_EXE)
	@echo "Código de saída: $$?"

# Apenas constrói (compila taquionc e o exemplo)
build: taqc $(OUT_EXE)

# Limpa os arquivos gerados
clean:
	@echo "==> Limpando tudo em $(BUILD_DIR)"
	rm -rf $(BUILD_DIR)

