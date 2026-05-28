---
status: draft
priority: high
---

# Replay Analyzer MVP — Design

## Goal

End-to-end vertical MVP, который позволяет после проигранной партии открыть локальный
дашборд, выбрать недавний повтор и за 10 секунд увидеть **что пошло не так**: провалы
популяции, ресурсные стопоры, поздние phase ups, крупные бои.

Артефакт: одна команда `replayreport`, поднимающая локальный сервер. Никакого
real-time, никакой кросс-партийной аналитики в v1 — только «открыл, выбрал, понял».

Парсер живёт в том же монорепо что и `civreport`, переиспользуя `internal/tmpl/civdata/tech/aura/i18n`
(хотя в MVP реально нужен только loader `metadata.json` и парсер `commands.txt` для
build-order events — глубокого обогащения через templates пока нет).

## User Story

> Я — zeto, играю в 0 A.D. за Спарту против Petra-5. Только что просрал.
> Открываю `~/Library/Application Support/0ad/replays/0.28.0/<latest>` глазами не хочу.
> Запускаю `replayreport`, в браузере вижу карточку последней игры, кликаю.
> На экране — графики популяции и ресурсов, маркеры боёв, список аномалий.
> Через 30 секунд понимаю: «застрял на камне 12 минут, потерял 77 поп в одном замесе,
> Phase City в 25:47 — поздно на 10 минут».

## In Scope (v1)

### Парсер — минимально достаточный

- **Reader** `commands.txt` (line-by-line, `bufio.Scanner`, принимает `io.Reader`)
- **Loader** `metadata.json` (опционально — без него replay-dir скипается в v1)
- **Events** (только необходимые типы): `construct`, `train`, `research`, `attack`,
  `garrison`, `tribute`, `resign`, `chat`. Остальные (walk, stance, formation, …)
  декодируются как «прочее» с raw data — не теряются, но и не типизируются
- **Snapshots** — собираются напрямую из `metadata.json.playerStates[i].sequences`.
  Своя in-memory reducer-симуляция в MVP не пишется
- **Metrics** в v1:
  - `phase_timings` — для каждого игрока: когда исследована town/city
  - `engagements` — агрегация подряд идущих `attack` команд на одну цель (3-секундное
    окно склейки) в `Engagement{t_start, t_end, target, units_engaged_peak, cmd_count}`
  - `anomalies`:
    - `pop_crash` — детект из sequences.populationCount: просадка ≥ 20 пунктов за ≤ 60 сек
    - `resource_stall` — `sequences.resourcesGathered.{res}` плоско ≥ 5 минут
    - `panic_garrison` — ≥ 5 garrison-команд за < 3 сек в одно здание
- **Outputs**: `analysis.json` рядом с replay-dir; кэшируется (если есть и mtime
  свежее commands.txt — переиспользуем)

### Дашик — минимально достаточный

Стек: **Go + templ + htmx + Plotly-JS**. Один бинарь `replayreport`, два режима:
- `replayreport <replay-dir>` — парсит один и поднимает дашик
- `replayreport` (без аргументов) — сканит дефолтную папку реплеев

Страницы:

1. **`/` — список повторов.** Карточки: дата, карта, длительность, цивы, результат.
   Скан + парсинг на старте (с кешированием).
2. **`/replay/{match_id}` — анализ.**
   - Header: game info + player chips (с цветами civ)
   - **Chart 1**: Population over time (Plotly line, все игроки overlaid, vertical
     dashed lines для phase ups, vertical solid lines для engagements ≥ 30 юнитов)
   - **Chart 2**: Resources gathered over time (4 subplots: food/wood/stone/metal,
     все игроки overlaid)
   - **Chart 3**: Active gatherers over time (4 area subplots)
   - **Sidebar**: Build Order — табличка значимых событий (construct, research, train
     champion/hero, фазы), `time | event`
   - **Bottom panel**: Anomalies — текст-список карточек с типом, окном времени,
     severity и деталями

## Out of Scope (v1)

- **Real-time / tail-режим** — отдельная спека (фаза 3)
- **Markdown-отчёт** — Python `replay_to_readable.py` пока остаётся, не трогаем
- **SQLite-индекс / cross-replay queries** — потом, когда наберётся достаточно
  партий и появится потребность в фильтрации «все мои Спарты с phase city > 20 мин»
- **APM / EAPM** — потом
- **Enrichment через templates** (cost/HP/build_time, классификация зданий) — потом,
  когда дойдём до coach-mode («ты потратил 1500 food на чемпионов, которых слил»)
- **Civ-specific benchmarks** (тайминги фаз vs ideal) — потом, нужна табличка по
  всем 15 цивам и валидация цифр
- **Heatmaps координатных данных** — координаты сохраняем в events, но не визуализируем
- **Engagement outcome inference** (кто выиграл бой) — невозможно без полноценной
  sim, нет данных о смертях в реальном времени
- **Idle worker detection** — слишком спекулятивно без симуляции, только сырые
  `ActiveGatherers` в snapshots
- **Replay-dirs без metadata.json** (игра упала / в процессе) — в v1 скипаются

## Architecture

```
replay-dir/                   ┌─→ analysis.json (per replay)
├── commands.txt   ──┐        │
└── metadata.json  ──┤        │
                     ▼        │
              ┌─────────────┐ │   replayreport
              │ commands/   │ │   ↓
              │  reader.go  │ │   HTTP server (localhost:8080)
              └──────┬──────┘ │   ↓
                     │        │   templ + htmx + Plotly-JS
                     ▼        │
              ┌─────────────┐ │
              │  events/    │ │
              │  decode.go  │ │
              └──────┬──────┘ │
                     │        │
                     ▼        │
              ┌─────────────┐ │
              │  metadata/  │ │
              │  loader.go  │ │
              └──────┬──────┘ │
                     │        │
                     ▼        │
              ┌─────────────┐ │
              │ analytics/  │ │
              │  + snapshot │ │
              └──────┬──────┘ │
                     │        │
                     ▼        │
              ┌─────────────┐ │
              │  output/    ├─┘
              │  writer.go  │
              └─────────────┘
```

**Принципы:**

1. Стадии — pure functions, кроме reader (читает файл) и snapshot/merge (мерджит с
   metadata)
2. **Events — единственный канал** между стадиями. REDUCER появится потом для real-time,
   сейчас не нужен — snapshots берём напрямую из metadata.json
3. Output JSON — стабильная schema (поле `schema_version: 1`)
4. **Идемпотентность парсинга** — если `analysis.json` свежее `commands.txt`, не парсим
   повторно (быстрый старт дашика на 86 существующих реплеях)

## Module Structure

```
0ad-civ-report-parser/
├── cmd/
│   ├── civreport/                   ← существующее
│   └── replayreport/                ← НОВОЕ
│       └── main.go                  ← CLI + HTTP server bootstrap
├── internal/
│   ├── tmpl/ civdata/ tech/ aura/ i18n/   ← reuse, не трогаем
│   └── replay/                      ← НОВОЕ
│       ├── commands/
│       │   ├── reader.go            ← line-based читалка commands.txt
│       │   └── reader_test.go
│       ├── metadata/
│       │   ├── loader.go            ← JSON struct + Load(path)
│       │   └── loader_test.go
│       ├── events/
│       │   ├── events.go            ← типизированные структуры
│       │   ├── decode.go            ← raw JSON cmd → typed event
│       │   └── decode_test.go
│       ├── snapshot/
│       │   ├── snapshot.go          ← Snapshot, PlayerSnapshot
│       │   └── from_metadata.go     ← разворачиваем metadata.sequences в []Snapshot
│       ├── analytics/
│       │   ├── phases.go            ← phase_timings из research events
│       │   ├── engagement.go        ← группировка attack-команд → engagements
│       │   ├── anomalies.go         ← pop_crash / resource_stall / panic_garrison
│       │   └── analytics_test.go
│       ├── output/
│       │   ├── schema.go            ← публичные типы (Analysis, Game, Player, …)
│       │   ├── writer.go            ← MarshalIndent + atomic write
│       │   └── version.go           ← const SchemaVersion = 1
│       ├── pipeline.go              ← Run(replayDir) (Analysis, error)
│       └── webui/
│           ├── server.go            ← HTTP сервер + роуты
│           ├── templates/
│           │   ├── layout.templ
│           │   ├── index.templ      ← список повторов
│           │   └── replay.templ     ← страница анализа
│           ├── static/              ← Plotly-JS (embedded via //go:embed)
│           └── handlers.go          ← хендлеры HTMX-фрагментов
└── testdata/replays/                ← 2–3 коротких real-replay для integration test
```

## Data Model

### Event

```go
type Event struct {
    T      int64  `json:"t"`        // ms от начала игры
    Turn   int    `json:"turn"`
    Player int    `json:"player"`
    Type   string `json:"type"`     // "train" | "construct" | …
    Data   any    `json:"data"`     // type-specific (TrainData | ConstructData | …)
}
```

Подтипы данных в `events.go`:
- `TrainData{Template, Count, Source, Queued}`
- `ConstructData{Template, X, Z, Angle, Builders, Foundation}`
- `ResearchData{Template, Source}`
- `AttackData{Target, Entities, AllowCapture, Queued}`
- `GarrisonData{Target, Entities, Queued}`
- `TributeData{To, Food, Wood, Stone, Metal}`
- `ResignData{}`
- `ChatData{Text}`
- `RawData{Type, JSON}` — для всех остальных типов команд (сохраняем без потерь)

### Snapshot

Один на 30 секунд игры (период sampling-а metadata.json sequences).

```go
type Snapshot struct {
    TSec    int                          `json:"t"`
    Players map[int]PlayerSnapshot       `json:"players"`
}

type PlayerSnapshot struct {
    Phase             string             `json:"phase"`
    PopCount          int                `json:"pop_count"`
    PopLimit          int                `json:"pop_limit"`
    PopMax            int                `json:"pop_max"`
    ResourcesOnHand   Resources          `json:"resources_on_hand"`
    ActiveGatherers   Resources          `json:"active_gatherers"`
    ResourcesGathered Resources          `json:"resources_gathered_cum"`
    MapExploredPct    float64            `json:"map_explored_pct"`
    MapControlledPct  float64            `json:"map_controlled_pct"`
    UnitsLostCum      map[string]int     `json:"units_lost_cum"`     // by class
    EnemyUnitsKilled  map[string]int     `json:"enemy_units_killed_cum"`
    BuildingsBuilt    map[string]int     `json:"buildings_built_cum"`
}

type Resources struct {
    Food, Wood, Stone, Metal int `json:"food,wood,stone,metal"`
}
```

### Metrics

```go
type Metrics struct {
    Players map[int]PlayerMetrics `json:"players"`
}

type PlayerMetrics struct {
    PhaseTimings map[string]int `json:"phase_timings"`  // "town": 295, "city": 1547 (секунды)
    Engagements  []Engagement   `json:"engagements"`
    Anomalies    []Anomaly      `json:"anomalies"`
}

type Engagement struct {
    TStart, TEnd      int     `json:"t_start,t_end"`   // секунды
    Target            int     `json:"target"`           // entity ID
    UnitsEngagedPeak  int     `json:"units_engaged_peak"`
    CommandCount      int     `json:"command_count"`
}

type Anomaly struct {
    Type      string         `json:"type"`     // "pop_crash" | "resource_stall" | "panic_garrison"
    TStart    int            `json:"t_start"`
    TEnd      int            `json:"t_end"`
    Severity  string         `json:"severity"` // "minor" | "warning" | "critical"
    Details   map[string]any `json:"details"`
}
```

### Top-level Analysis

```go
type Analysis struct {
    SchemaVersion int        `json:"schema_version"`
    Game          GameInfo   `json:"game"`
    Players       []Player   `json:"players"`
    Events        []Event    `json:"events"`
    Snapshots     []Snapshot `json:"snapshots"`
    FinalState    FinalState `json:"final_state"`
    Metrics       Metrics    `json:"metrics"`
}
```

## Discovery Items (выяснить при имплементации)

1. **Petra AI 0-commands** — в существующем `replay_analysis.md` AI показал «0 commands
   total». Грепнуть `commands.txt` на `cmd 2` / `cmd 3`. Если AI-команды отсутствуют —
   зафиксировать как известное ограничение MVP («анализ доступен полноценно только для
   человека-игрока»). Если есть — починить парсер.

2. **Sampling interval `metadata.json sequences`** — предполагаем 30 секунд, проверить
   на нескольких разных по длине партиях; если переменное — взять `time[i+1] - time[i]`.

3. **Resign / wonder-victory / capture-CC** — как надёжно определить outcome когда
   formal `state` остался `active`. Простая эвристика для v1: если есть `resign` event
   → defeated. Если game завершилась а у игрока live CC нет → defeated. Иначе — active/won.

4. **Engine version display** (F16) — брать из `mods[0].version`, не из
   `engine_serialization_version`.

5. **Outcome определяется из metadata.json или из последнего snapshot?** Простой ответ:
   `metadata.playerStates[i].state` + сверка с наличием `resign` event-а.

## Testing

- **Unit:** каждый пакет имеет `*_test.go` рядом. Heavy — analytics (pop_crash detector,
  engagement aggregation), events.decode (все типы команд).
- **Integration:** `testdata/replays/short-spart-vs-petra/` (≈ 50KB commands.txt + metadata).
  Полный pipeline → ассерт ключевых полей `analysis.json`.
- **Smoke:** `replayreport --all` на всей папке реплеев пользователя — должен пройти
  без ошибок (но не падает на replay-dir-ах без metadata.json — скипает).

## Process / Pace

Реализуется единым blob-ом за один сеанс, без эпиков. План реализации (writing-plans)
будет коротким — 4–6 шагов:

1. Skeleton: пакеты, типы данных, пустые функции
2. Парсер `commands.txt` + `metadata.json` + events.decode
3. Snapshot extraction из metadata
4. Analytics: phases / engagements / anomalies
5. CLI + HTTP server + templ-шаблоны
6. Plotly-графики и сборка страницы

Промежуточных коммитов минимум — один по концу каждого крупного шага либо один большой
в конце.

## Open Decisions

| Решение | Принято в брейнсторме |
|---|---|
| Монорепо vs отдельный репо | Монорепо (расширяем `0ad-civ-report-parser`, опц. переименовать в `0ad-tools`) |
| Output: JSON vs SQLite vs оба | Только JSON в v1. SQLite добавим когда понадобится cross-replay |
| Дашик: Streamlit vs Go+templ vs static HTML | Go + templ + htmx + Plotly-JS |
| Парсер: один проход vs streaming | Streaming-friendly (`bufio.Scanner` + `io.Reader`), но в MVP вызывается батчем |
| Markdown-отчёт | Откладываем, Python-скрипт остаётся как есть |
| Real-time mode | Отдельная спека (фаза 3), не в MVP |
| Когда parsing-кеширование | `analysis.json` свежее `commands.txt` mtime → не парсим повторно |
