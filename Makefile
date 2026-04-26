GO       ?= go
BIN_DIR  ?= bin
BIN      ?= $(BIN_DIR)/civreport
CIVS     := athen brit cart gaul germ han iber kush mace maur pers ptol rome sele spart
GAMEDATA ?=
OUT      ?=

GAMEDATA_FLAG := $(if $(GAMEDATA),--gamedata $(GAMEDATA),)
OUT_FLAG      := $(if $(OUT),--out $(OUT),)

.PHONY: help build all-civs check test clean civ $(CIVS)

.DEFAULT_GOAL := help

help:
	@echo "Цели:"
	@echo "  build              сборка бинарника civreport"
	@echo "  all-civs           отчёты по всем $(words $(CIVS)) цивилизациям (--all)"
	@echo "  <civ>              отчёт по конкретной циве, напр.: make spart, make athen"
	@echo "  civ CIV=<alias>    отчёт по алиасу/русскому имени, напр.: make civ CIV=спарт"
	@echo "  check              smoke-тест без записи файлов (--check)"
	@echo "  test               go test ./..."
	@echo "  clean              удалить *_buildings_report.md и бинарник"
	@echo ""
	@echo "Переменные:"
	@echo "  GAMEDATA=<path>    путь к 0ad/binaries/data/mods/public"
	@echo "  OUT=<file>         явный путь для одиночного отчёта (--out)"
	@echo ""
	@echo "Цивы: $(CIVS)"

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN) ./cmd/civreport

all-civs: build
	$(BIN) $(GAMEDATA_FLAG) --all

$(CIVS): build
	$(BIN) $(GAMEDATA_FLAG) $(OUT_FLAG) $@

civ: build
	@if [ -z "$(CIV)" ]; then echo "usage: make civ CIV=<name|alias>"; exit 1; fi
	$(BIN) $(GAMEDATA_FLAG) $(OUT_FLAG) $(CIV)

check: build
	$(BIN) $(GAMEDATA_FLAG) --check

test:
	$(GO) test ./...

clean:
	rm -f *_buildings_report.md
	rm -rf $(BIN_DIR)
