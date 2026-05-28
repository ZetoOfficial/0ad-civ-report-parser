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

## replayreport

`replayreport` парсит реплеи 0 A.D. (commands.txt + metadata.json), пишет
`analysis.json` рядом с replay-dir, и поднимает локальный дашик с графиком
плотности команд, фазовыми маркерами, build-order и аномалиями.

```bash
make replayreport                            # → bin/replayreport
./bin/replayreport                           # сканит ~/Library/Application Support/0ad/replays/0.28.0/
                                             # → http://localhost:8080
./bin/replayreport <replay-dir>              # парсит один replay (без HTTP)
./bin/replayreport --check --all             # CI-режим: exit 2 если хоть один реплей упал
./bin/replayreport --replays /path/to/dir    # альтернативный корень
./bin/replayreport --addr :9000              # альтернативный порт
```

Для каждого replay-dir пишется `analysis.json` (schema v1) рядом с
`commands.txt`. Кэш по mtime: если `analysis.json` свежее `commands.txt` —
не парсим повторно (старт дашика на 80+ реплеях за секунды).

Replay-dir без `metadata.json` пропускаются (игра упала / не дошла до
summary screen). В v1 это ожидаемое поведение.

### Что в дашике

- **Список партий** — карточки с датой, картой, длительностью, цивами и исходом.
- **Страница партии** — header + игроки с финальным состоянием + Plotly-график
  плотности команд (30-сек бины × 5 категорий) с маркерами фаз (пунктир)
  и крупных боёв (сплошные красные линии) + build-order + аномалии.

### Что НЕ в v1

Time-series графики population / resources / active gatherers — `metadata.json`
реальных реплеев не содержит `sequences` (0ad пишет их только при показе
summary screen, который пользователь обычно скипает). Замена — `Chart 1`
(плотность команд). См. `docs/superpowers/specs/2026-05-28-replay-analyzer-mvp-design.md`
и `docs/superpowers/plans/2026-05-28-replay-analyzer-mvp.md` для деталей.

Также вне scope v1: APM/EAPM, кросс-партийная аналитика, real-time mode,
markdown-отчёт по партии, idle worker detection.
