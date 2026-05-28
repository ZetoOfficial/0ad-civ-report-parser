GO       ?= go
BIN_DIR  ?= bin
BIN      ?= $(BIN_DIR)/civreport
CIVS     := athen brit cart gaul germ han iber kush mace maur pers ptol rome sele spart
GAMEDATA ?=
OUT_DIR  ?=
CONFIG   ?=

GAMEDATA_FLAG := $(if $(GAMEDATA),--gamedata $(GAMEDATA),)
OUT_DIR_FLAG  := $(if $(OUT_DIR),--out-dir $(OUT_DIR),)
CONFIG_FLAG   := $(if $(CONFIG),--config $(CONFIG),)

.PHONY: help build all-civs check test clean civ golden-diff $(CIVS) replayreport replay-check

.DEFAULT_GOAL := help

help:
	@echo "Цели:"
	@echo "  build              сборка бинарника civreport"
	@echo "  all-civs           отчёты по всем $(words $(CIVS)) цивилизациям (--all)"
	@echo "  <civ>              отчёт по конкретной циве, напр.: make spart"
	@echo "  civ CIV=<alias>    отчёт по алиасу/русскому имени, напр.: make civ CIV=спарт"
	@echo "  check              smoke-тест без записи файлов (--check)"
	@echo "  golden-diff CIV=germ"
	@echo "                     показать diff overview+structree против testdata/golden/"
	@echo "  test               go test ./..."
	@echo "  clean              удалить сгенерированные .md и бинарник"
	@echo ""
	@echo "Переменные:"
	@echo "  GAMEDATA=<path>    путь к 0ad/binaries/data/mods/public"
	@echo "  OUT_DIR=<dir>      каталог для генерируемых файлов (--out-dir)"
	@echo "  CONFIG=<path>      путь к JSON-конфигу (--config)"
	@echo ""
	@echo "Цивы: $(CIVS)"

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN) ./cmd/civreport

all-civs: build
	$(BIN) $(GAMEDATA_FLAG) $(CONFIG_FLAG) $(OUT_DIR_FLAG) --all

$(CIVS): build
	$(BIN) $(GAMEDATA_FLAG) $(CONFIG_FLAG) $(OUT_DIR_FLAG) $@

civ: build
	@if [ -z "$(CIV)" ]; then echo "usage: make civ CIV=<name|alias>"; exit 1; fi
	$(BIN) $(GAMEDATA_FLAG) $(CONFIG_FLAG) $(OUT_DIR_FLAG) $(CIV)

check: build
	$(BIN) $(GAMEDATA_FLAG) $(CONFIG_FLAG) --check

# Informational only — does not fail the build. Strict diff lands in Epic 4.
golden-diff: build
	@if [ -z "$(CIV)" ]; then echo "usage: make golden-diff CIV=<civcode>"; exit 1; fi
	@tmp=$$(mktemp -d) && \
	 $(BIN) $(GAMEDATA_FLAG) $(CONFIG_FLAG) --out-dir $$tmp $(CIV) && \
	 base=$$($(BIN) --print-basename $(CIV)) && \
	 echo "=== overview diff ===" && \
	 diff -u testdata/golden/$${base}_overview.md $$tmp/$${base}_overview.md || true ; \
	 echo "=== structree diff ===" && \
	 diff -u testdata/golden/$${base}_structree.md $$tmp/$${base}_structree.md || true

test:
	$(GO) test ./...

clean:
	rm -rf out/
	rm -f *_overview.md *_structree.md common.md
	rm -f *_buildings_report.md
	rm -rf $(BIN_DIR)

replayreport:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/replayreport ./cmd/replayreport

replay-check: replayreport
	./$(BIN_DIR)/replayreport --check --all
