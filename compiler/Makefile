LLVM_CFLAGS := $(shell llvm-config --cflags)
LLVM_LDFLAGS := $(shell llvm-config --ldflags --libs --system-libs)

TAQ_NAME   := math_precendence
TAQ_SRC    := ../examples/$(TAQ_NAME).taq
BUILD_DIR  := ../build

TAQC_BIN   := $(BUILD_DIR)/taquionc.exe
LLVM_IR    := $(BUILD_DIR)/output.ll
OUT_EXE    := $(BUILD_DIR)/$(TAQ_NAME).exe

.PHONY: all run clean build taqc

all: build taqc $(OUT_EXE)

# Compila o taquionc
taqc:
	@echo "==> Compilando taquionc atualizado..."
	@mkdir -p $(BUILD_DIR)
	CGO_CFLAGS="$(LLVM_CFLAGS)" \
	CGO_LDFLAGS="$(LLVM_LDFLAGS)" \
	go build -o $(TAQC_BIN) -tags="llvm20 byollvm" ./cmd/taquionc

# Gera IR e compila .exe em seguida
$(OUT_EXE): taqc
	@echo "==> Gerando LLVM IR: $(LLVM_IR)"
	$(TAQC_BIN) $(TAQ_SRC) -o $(LLVM_IR)
	@echo "==> Compilando IR com clang: $(OUT_EXE)"
	clang $(LLVM_IR) -o $(OUT_EXE)

run: all
	@echo "==> Executando $(OUT_EXE)..."
	@$(OUT_EXE)
	@echo "Código de saída: $$?"

build: taqc

clean:
	@echo "==> Limpando tudo em $(BUILD_DIR)"
	rm -rf $(BUILD_DIR)

# go build -o taquionc.exe -tags="llvm20 byollvm" ./cmd/taquionc