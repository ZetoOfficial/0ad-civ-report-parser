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

# Добавить историческую справку в overview
civreport --include-history spart
```

### Флаги

- `--gamedata <path>` — путь к `mods/public` (переопределяет `OAD_GAMEDATA_ROOT` и `config.json`).
- `--out-dir <path>` — каталог вывода (по умолчанию `out/`, создаётся при необходимости).
- `--config <path>` — путь к JSON-конфигу.
- `--include-history` — добавляет блок «Историческая справка» в `<civ>_overview.md` из `Identity/History` Player-шаблона. По умолчанию выключено.
- `--lang ru` — gettext-перевод (не реализовано до эпика 5).
- `--all` — генерировать отчёты для всех 15 цив.
- `--check` — smoke-проверка без записи файлов.

### config.json

```json
{
  "gamedata": "/Users/zeto/Projects/study/0ad/binaries/data/mods/public",
  "out_dir": "out",
  "lang": "",
  "include_history": false,
  "include_icons": false
}
```

Все поля опциональны; CLI-флаги перекрывают значения из JSON.

## Содержимое отчётов

### `<civ>_overview.md`

Семь обязательных секций (Идентичность, Герои, Уникальные строения,
Уникальные технологии, Цивилизационные бонусы, Командный бонус,
Технологии-недоступные) + опциональная Историческая справка + footer-сноска
со ссылкой на `common.md`. Цив-бонусы содержат колонку «Требования» и
per-mod суффикс `(только …)`.

### `<civ>_structree.md`

Поэтапные блоки VILLAGE/TOWN/CITY PHASE с зданиями, таблицами
параметров и подтаблицами «Тренирует» / «Исследует». Приложения:
детальная информация по типам юнитов и сводная таблица строимых зданий.

### `common.md`

Общая для всех 15 цив справка: модификаторы Advanced/Elite, прочие
глобальные авто-эффекты, типы урона, типы ресурсов, статус-эффекты.
Цив-overview ссылается на якоря этого файла.

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
