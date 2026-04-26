# 0ad-civ-report-parser

CLI на Go, генерирующий русскоязычные markdown-отчёты о цивилизациях
0 A.D. напрямую из XML/JSON шаблонов игры.

## Установка

```bash
make build              # → bin/civreport
go install ./cmd/civreport
```

## Использование

```bash
# Один отчёт — три файла
civreport spart                          # → spartans_overview.md, spartans_structree.md, common.md

# Все цивы
civreport --all                          # 30 файлов + common.md

# Smoke-тест
civreport --check

# Указать каталог вывода
civreport --out-dir /tmp/out spart

# Указать путь к данным игры
civreport --gamedata /path/to/0ad/binaries/data/mods/public spart
OAD_GAMEDATA_ROOT=/path civreport spart

# JSON-конфиг
civreport --config ./config.json spart
```

### config.json

```json
{
  "gamedata": "/Users/zeto/Projects/study/0ad/binaries/data/mods/public",
  "out_dir": ".",
  "lang": "",
  "include_history": false,
  "include_icons": false
}
```

Все поля опциональны; CLI-флаги перекрывают значения из JSON.
В эпике 1 активны: `gamedata`, `out_dir`. Остальные подключаются
в эпиках 2 и 5.

## Make-цели

```bash
make build                              # сборка
make spart                              # отчёт по конкретной циве
make all-civs                           # все 15 цив
make check                              # smoke без записи
make golden-diff CIV=germ               # diff против testdata/golden/germans_*.md
make test                               # go test ./...
make clean                              # удалить сгенерированные .md и bin/
```

## Алиасы цив

`spart`, `спарт`, `Spartans`, `спартанцы` — все резолвятся в код `spart`.
Полный список: athen, brit, cart, gaul, germ, han, iber, kush, mace,
maur, pers, ptol, rome, sele, spart.

## Источник данных

`/Users/zeto/Projects/study/0ad/binaries/data/mods/public` (READ-ONLY).
Подкаталоги — см. `docs/sources.md`.
