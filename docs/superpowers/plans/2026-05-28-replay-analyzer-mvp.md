# Replay Analyzer MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Один Go-бинарь `replayreport`, который сканит `~/Library/Application Support/0ad/replays/0.28.0/`, парсит `commands.txt` + `metadata.json`, кэширует результат в `analysis.json` рядом с реплеем, и поднимает локальный дашик (`http://localhost:8080`) со списком партий и страницей анализа: build-order timeline, action-density chart, phase markers, engagements, аномалии, итоговое состояние.

**Architecture:** Новый каталог `internal/replay/` — изолированный от существующего `civreport`. Stages — pure functions, единственный канал между ними — `Event`. Output — `analysis.json` со стабильной schema (`schema_version: 1`). Дашик — Go + `templ` + htmx + Plotly-JS (embedded via `//go:embed`). Бинарь `cmd/replayreport/main.go` принимает либо конкретный replay-dir, либо запускается без аргументов и сканит дефолтную папку.

**Tech Stack:** Go 1.25.5 (stdlib + `github.com/a-h/templ`), Plotly-JS (vendored файлом), htmx (vendored файлом). Существующие `internal/{tmpl,civdata,tech,aura,i18n}` не трогаем — в MVP не используются.

**Spec:** [docs/superpowers/specs/2026-05-28-replay-analyzer-mvp-design.md](../specs/2026-05-28-replay-analyzer-mvp-design.md)

**Branch:** `replay-analyzer-mvp` (создать перед стартом).

**Источник данных:** `~/Library/Application Support/0ad/replays/0.28.0/` (read-only пользовательские реплеи; ~86 шт на 2026-05-28). Под integration-тесты — копии в `testdata/replays/` (2 коротких).

---

## ⚠ Отклонение от спецификации (discovery 2026-05-28)

**Спека предполагает** `metadata.json.playerStates[i].sequences` — массивы pop/resources по времени, из которых строятся snapshots. **Реальность:** проверено на всех 86 реплеях пользователя — поля `sequences` **нет ни в одном**. `metadata.json` содержит только финальный (или последний промежуточный) скаляр-снапшот: `popCount`, `resourceCounts`, `phase`, `state`, `researchedTechs`. Это связано с тем, что 0ad пишет `sequences` только при показе summary screen после завершения партии, а пользователь обычно квитает в лобби, не дожидаясь экрана.

**Следствие для v1:** Time-series графики `Population over time`, `Resources gathered over time`, `Active gatherers over time` (Charts 1–3 из спеки) — **не реализуемы без своей reducer-симуляции**, которую спека сама вынесла в out-of-scope ("Real-time / tail-режим — отдельная спека"). Замена в этом плане:

- **`Chart 1: Action density over time`** — стэкнутый bar-chart, 30-секундные бины, цветовая разбивка по категориям команд (military / economy / build / research / other). Из commands.txt всегда доступно.
- **Phase markers** (vertical dashed lines) — из research-команд (`phase_town_*`, `phase_city_*`). Остаются как в спеке.
- **Engagement markers** (vertical solid lines) — из агрегации `attack`-команд. Остаются как в спеке.
- **Build Order sidebar**, **engagements list**, **anomalies panel** (только `panic_garrison`, остальные две аномалии требовали sequences) — остаются.
- **Финальное состояние** (новая панель) — карточка из `metadata.playerStates[i]`: phase / popCount / final resources / researchedTechs count / state. Это единственный source-of-truth по итогам матча.

Snapshot-структура из спеки **не реализуется в v1**. В output JSON поле `snapshots` присутствует как пустой массив (для forward-compat schema); поле `final_state` несёт реальные данные.

---

## Source Data Shape (verified 2026-05-28)

### `commands.txt` line types

```
start {<settings + matchID + timestamp + mods + ...>}
turn <N> <tick_ms>                                           # 200ms per turn at gameSpeed=1
cmd <player_id> {"type":"<kind>", ...}                       # player_id 1-indexed; gaia=0
hash <hex64>
hash-quick <hex32>
end
```

`turn` line preceding a series of `cmd` lines устанавливает игровое время; время команды = время последнего `turn` + 0 (все cmd в одном тике). Время в миллисекундах от старта (turn 0 = t=0; `turn N` несёт длительность тика `tick_ms`, обычно 200).

Известные `cmd.type` (из реальных данных): `walk`, `gather`, `attack`, `train`, `repair`, `set-rallypoint`, `construct`, `research`, `returnresource`, `garrison`, `walk-custom`, `barter`, `stop-production`, `unset-rallypoint`, `unload-all`, `setup-trade-route`, `resign`, `delete-entities`, `construct-wall`. Прочие декодируются в `RawData`.

### `metadata.json` (final state only)

```jsonc
{
  "timeElapsed": 1860400,                       // ms
  "playerStates": [
    {
      "name": "ZetoQqq (1303)",
      "civ": "mace",
      "color": {"r":10,"g":10,"b":190},
      "popCount": 64, "popLimit": 200, "popMax": 200,
      "resourceCounts": {"food":420,"wood":380,"stone":12,"metal":50},
      "phase": "town",                          // "" | "village" | "town" | "city"
      "state": "active",                        // "active" | "won" | "defeated"
      "team": 0, "isAlly":{...}, "isEnemy":{...},
      "researchedTechs": ["phase_town_generic", ...],
      "disabledTechnologies": [], "disabledTemplates": [],
      "classCounts": {...}, "typeCountsByClass": {...},
      "entityCounts": {...}, "matchEntityCounts": {...}
    },
    ...
  ],
  "mapSettings": {...}                          // не используется в MVP
}
```

В этих 86 реплеях `sequences` нет. `popCount=0, state=active` для старых матчей — артефакт того, что игра не успела дойти до второй фазы метадата-сэмплинга. Не падать на этих данных.

### Player numbering

`cmd <P>` использует 1-индексацию (gaia=0 не пишет команд). `metadata.playerStates[i]` использует 0-индексацию, где `i=0` = gaia, `i=1` = player 1. Маппинг: `cmd P → playerStates[P]`.

### Discovery items (выяснить во время Task 4)

1. **Petra AI и `cmd <P>`** — грепнуть тестовый реплей с AI на `cmd 3` / `cmd 4`. Гипотеза: AI пишет команды как и человек. Если ноль — зафиксировать как известное ограничение «анализ Petra AI недоступен в MVP».
2. **`tick_ms` стабильность** — обычно 200ms, но возможны варианты при паузах. Накапливать время как `sum(tick_ms)`, не `turn * 200`.
3. **Outcome detection priority**: `resign event in commands.txt` → defeated. Иначе `metadata.playerStates[P].state`. Иначе — `unknown`.

---

## File Structure

| Path | Action | Responsibility |
|---|---|---|
| `internal/replay/events/events.go` | create | Типы `Event`, `TrainData`, `ConstructData`, `ResearchData`, `AttackData`, `GarrisonData`, `TributeData`, `ResignData`, `ChatData`, `RawData` |
| `internal/replay/commands/reader.go` | create | Stream `commands.txt`: `Reader{file}`, метод `Scan() (Line, bool)`; `Line` — sum-type (`Start`, `Turn`, `Cmd`, `Hash`, `End`) |
| `internal/replay/commands/reader_test.go` | create | Fixture-driven: маленький `commands.txt` → последовательность `Line` |
| `internal/replay/events/decode.go` | create | `Decode(line commands.Cmd, tMs int64) Event` — JSON в типизированный `Event` |
| `internal/replay/events/decode_test.go` | create | Per-type unit-тесты на эталонных JSON-строках |
| `internal/replay/metadata/loader.go` | create | `Load(path string) (*Metadata, error)` — парсинг `metadata.json` + типы `Metadata`, `PlayerState`, `Settings`, `PlayerData` |
| `internal/replay/metadata/loader_test.go` | create | Минимальный JSON → ожидаемые скаляры |
| `internal/replay/analytics/phases.go` | create | `PhaseTimings(events []Event) map[int]map[string]int` — секунды для town/city per player из research events |
| `internal/replay/analytics/engagement.go` | create | `Engagements(events []Event, windowSec int) []Engagement` — склейка attack-cmd в окне 3 сек по `(player, target)` |
| `internal/replay/analytics/anomalies.go` | create | `PanicGarrison(events []Event) []Anomaly` — ≥5 garrison-cmd в одно здание за <3 сек |
| `internal/replay/analytics/density.go` | create | `ActionDensity(events []Event, binSec int) []DensityBin` — категория-агрегация для chart |
| `internal/replay/analytics/analytics_test.go` | create | Все четыре анализатора на синтетических event-списках |
| `internal/replay/output/schema.go` | create | `Analysis`, `GameInfo`, `Player`, `FinalState`, `Engagement`, `Anomaly`, `DensityBin`; const `SchemaVersion = 1` |
| `internal/replay/output/writer.go` | create | `Write(path string, a *Analysis) error` — atomic temp+rename, MarshalIndent |
| `internal/replay/output/writer_test.go` | create | Roundtrip MarshalIndent/Unmarshal; atomic write; mtime cache helper |
| `internal/replay/pipeline.go` | create | `Run(replayDir string) (*output.Analysis, error)` — orchestrates reader → decode → metadata → analytics → analysis |
| `internal/replay/pipeline_test.go` | create | Integration: testdata replay → assert ключевых полей |
| `internal/replay/webui/server.go` | create | `NewServer(repDir string) http.Handler` — роуты `/`, `/replay/{matchID}`, `/static/*` |
| `internal/replay/webui/handlers.go` | create | Список матчей + страница анализа |
| `internal/replay/webui/templates/layout.templ` | create | Базовый layout (head/body/htmx-include) |
| `internal/replay/webui/templates/index.templ` | create | Карточный список матчей |
| `internal/replay/webui/templates/replay.templ` | create | Страница анализа: header / chart / sidebar / final-state / anomalies |
| `internal/replay/webui/static/plotly.min.js` | create | Vendored Plotly-JS (~3MB minified) |
| `internal/replay/webui/static/htmx.min.js` | create | Vendored htmx (~50KB) |
| `internal/replay/webui/static/style.css` | create | Минимальные стили (карточки, chip-цвета) |
| `cmd/replayreport/main.go` | create | CLI: парсинг флагов, либо `replayreport <dir>` (один), либо `replayreport` (сканит и поднимает HTTP) |
| `Makefile` | modify | + `replayreport` target, `replay-check` target |
| `testdata/replays/short-spart-vs-petra/commands.txt` | create | Копия короткого реплея (~50KB) |
| `testdata/replays/short-spart-vs-petra/metadata.json` | create | соответствующий metadata.json |
| `testdata/replays/short-no-metadata/commands.txt` | create | Реплей без metadata.json (на стресс-тест skip-логики) |
| `go.mod` / `go.sum` | modify | + `github.com/a-h/templ` |

Существующие пакеты (`internal/{tmpl,civdata,tech,aura,i18n,render,paths,config}`) **не трогаем**.

---

## Task 1: Skeleton + Analysis schema + roundtrip test

**Files:**
- Create: `internal/replay/output/schema.go`
- Create: `internal/replay/output/writer.go`
- Create: `internal/replay/output/writer_test.go`

- [ ] **Step 1.1: Create branch**

```bash
git checkout -b replay-analyzer-mvp
```

- [ ] **Step 1.2: Write `schema.go`**

```go
package output

const SchemaVersion = 1

type Analysis struct {
	SchemaVersion int        `json:"schema_version"`
	Game          GameInfo   `json:"game"`
	Players       []Player   `json:"players"`
	Events        []Event    `json:"events"`
	Snapshots     []Snapshot `json:"snapshots"` // empty in v1; reserved for forward compat
	FinalState    FinalState `json:"final_state"`
	Metrics       Metrics    `json:"metrics"`
}

type GameInfo struct {
	MatchID       string `json:"match_id"`
	Map           string `json:"map"`
	MapType       string `json:"map_type"`
	Timestamp     int64  `json:"timestamp"`     // unix sec
	DurationMs    int64  `json:"duration_ms"`
	EngineVersion string `json:"engine_version"`
	VictoryConds  []string `json:"victory_conditions"`
}

type Player struct {
	ID    int    `json:"id"`    // 1-based; matches cmd <P>
	Name  string `json:"name"`
	Civ   string `json:"civ"`
	Team  int    `json:"team"`
	IsAI  bool   `json:"is_ai"`
	AIDiff int   `json:"ai_diff,omitempty"`
	Color Color  `json:"color"`
}

type Color struct{ R, G, B int }

type Event struct {
	T      int64           `json:"t"`      // ms from game start
	Player int             `json:"player"` // 1-based
	Type   string          `json:"type"`
	Data   any             `json:"data,omitempty"`
}

type Snapshot struct{} // placeholder; not populated in v1

type FinalState struct {
	Players map[int]PlayerFinalState `json:"players"`
}

type PlayerFinalState struct {
	Phase           string         `json:"phase"`
	State           string         `json:"state"`   // "active" | "won" | "defeated"
	Outcome         string         `json:"outcome"` // resolved (resign event takes precedence)
	PopCount        int            `json:"pop_count"`
	PopLimit        int            `json:"pop_limit"`
	PopMax          int            `json:"pop_max"`
	ResourceCounts  Resources      `json:"resource_counts"`
	ResearchedTechs []string       `json:"researched_techs"`
}

type Resources struct {
	Food, Wood, Stone, Metal int
}

type Metrics struct {
	Players map[int]PlayerMetrics `json:"players"`
	Density []DensityBin          `json:"action_density"`
}

type PlayerMetrics struct {
	PhaseTimings map[string]int `json:"phase_timings"` // sec
	Engagements  []Engagement   `json:"engagements"`
	Anomalies    []Anomaly      `json:"anomalies"`
}

type Engagement struct {
	TStartSec, TEndSec int `json:"t_start_sec,t_end_sec"`
	Target             int `json:"target"`
	PeakUnits          int `json:"peak_units"`
	CommandCount       int `json:"command_count"`
}

type Anomaly struct {
	Type     string         `json:"type"`
	TStartSec int           `json:"t_start_sec"`
	TEndSec  int            `json:"t_end_sec"`
	Severity string         `json:"severity"`
	Details  map[string]any `json:"details,omitempty"`
}

type DensityBin struct {
	TSec   int            `json:"t_sec"`        // bin start
	Counts map[string]int `json:"counts"`       // category → count (per all players)
}
```

- [ ] **Step 1.3: Write writer + failing roundtrip test**

`writer.go`:

```go
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Write atomically marshals a as JSON to path (temp file + rename).
func Write(path string, a *Analysis) error {
	raw, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("output: marshal: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".analysis-*.json")
	if err != nil {
		return fmt.Errorf("output: tempfile: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return fmt.Errorf("output: write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("output: close: %w", err)
	}
	return os.Rename(tmp.Name(), path)
}

// IsFresh reports whether path exists and is newer than the source (commandsTxt).
func IsFresh(path, commandsTxt string) bool {
	a, err := os.Stat(path)
	if err != nil {
		return false
	}
	b, err := os.Stat(commandsTxt)
	if err != nil {
		return false
	}
	return a.ModTime().After(b.ModTime())
}
```

`writer_test.go`:

```go
package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "analysis.json")
	in := &Analysis{
		SchemaVersion: SchemaVersion,
		Game:          GameInfo{MatchID: "ABC", Map: "punjab_2", DurationMs: 1860400},
		Players:       []Player{{ID: 1, Name: "zeto", Civ: "spart"}},
		Events:        []Event{{T: 1200, Player: 1, Type: "train"}},
		Snapshots:     []Snapshot{},
	}
	if err := Write(path, in); err != nil {
		t.Fatalf("Write: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got Analysis
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d, want 1", got.SchemaVersion)
	}
	if got.Game.MatchID != "ABC" {
		t.Errorf("MatchID = %q, want ABC", got.Game.MatchID)
	}
	if len(got.Players) != 1 || got.Players[0].Civ != "spart" {
		t.Errorf("Players = %+v", got.Players)
	}
}

func TestIsFresh(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "commands.txt")
	out := filepath.Join(dir, "analysis.json")
	os.WriteFile(src, []byte("x"), 0o644)
	if IsFresh(out, src) {
		t.Fatal("missing analysis must be stale")
	}
	os.WriteFile(out, []byte("{}"), 0o644)
	// out written after src (Go writes are monotonic in practice; but force via Chtimes)
	now := time.Now()
	os.Chtimes(src, now.Add(-time.Minute), now.Add(-time.Minute))
	os.Chtimes(out, now, now)
	if !IsFresh(out, src) {
		t.Fatal("analysis newer than commands must be fresh")
	}
}
```

- [ ] **Step 1.4: Run tests, expect FAIL → PASS**

```bash
go test ./internal/replay/output/...
```

Expected: PASS (no dependencies; pure stdlib).

- [ ] **Step 1.5: Commit**

```bash
git add internal/replay/output/ docs/superpowers/plans/2026-05-28-replay-analyzer-mvp.md
git commit -m "replay: bootstrap output schema + atomic writer"
```

---

## Task 2: commands.txt reader

**Files:**
- Create: `internal/replay/commands/reader.go`
- Create: `internal/replay/commands/reader_test.go`
- Create: `internal/replay/commands/testdata/sample.txt`

- [ ] **Step 2.1: Write fixture `testdata/sample.txt`**

```
start {"settings":{"PlayerData":[{"Civ":"spart","Name":"zeto"}]},"matchID":"ABC","timestamp":1779953790,"mods":[{"mod":"public","name":"0ad","version":"0.28.0"}]}
turn 0 200
hash deadbeef0000000000000000000000000000000000000000000000000000beef
hash-quick cafebabe00000000000000000000000000000000000000000000000000000000
turn 1 200
cmd 2 {"type":"research","entity":2481,"template":"gather_wicker_baskets","pushFront":false}
cmd 2 {"type":"train","entities":[2481],"template":"units/spart/infantry_spearman_b","count":5,"pushFront":false}
turn 2 200
cmd 2 {"type":"resign"}
end
```

- [ ] **Step 2.2: Write `reader.go`**

```go
// Package commands streams lines from a replay commands.txt file.
package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type LineKind int

const (
	KindStart LineKind = iota
	KindTurn
	KindCmd
	KindHash
	KindHashQuick
	KindEnd
)

type Line struct {
	Kind   LineKind
	StartJSON json.RawMessage // KindStart
	TurnN  int               // KindTurn
	TickMs int               // KindTurn
	Player int               // KindCmd (1-based)
	CmdJSON json.RawMessage  // KindCmd
	Hash   string            // KindHash, KindHashQuick
}

type Reader struct {
	sc *bufio.Scanner
}

func New(r io.Reader) *Reader {
	sc := bufio.NewScanner(r)
	// real replays have very long start lines (>64KB default buffer)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	return &Reader{sc: sc}
}

// Next reads the next line. Returns (Line, true, nil) on success,
// (Line{}, false, nil) on EOF, (Line{}, false, err) on parse error.
func (r *Reader) Next() (Line, bool, error) {
	if !r.sc.Scan() {
		if err := r.sc.Err(); err != nil {
			return Line{}, false, err
		}
		return Line{}, false, nil
	}
	raw := r.sc.Text()
	if raw == "" {
		return r.Next()
	}
	switch {
	case raw == "end":
		return Line{Kind: KindEnd}, true, nil
	case strings.HasPrefix(raw, "start "):
		return Line{Kind: KindStart, StartJSON: json.RawMessage(raw[len("start "):])}, true, nil
	case strings.HasPrefix(raw, "turn "):
		var n, ms int
		if _, err := fmt.Sscanf(raw, "turn %d %d", &n, &ms); err != nil {
			return Line{}, false, fmt.Errorf("commands: bad turn line %q: %w", raw, err)
		}
		return Line{Kind: KindTurn, TurnN: n, TickMs: ms}, true, nil
	case strings.HasPrefix(raw, "cmd "):
		sp1 := strings.IndexByte(raw[4:], ' ')
		if sp1 < 0 {
			return Line{}, false, fmt.Errorf("commands: bad cmd line %q", raw)
		}
		p, err := strconv.Atoi(raw[4 : 4+sp1])
		if err != nil {
			return Line{}, false, fmt.Errorf("commands: bad cmd player %q: %w", raw, err)
		}
		return Line{Kind: KindCmd, Player: p, CmdJSON: json.RawMessage(raw[4+sp1+1:])}, true, nil
	case strings.HasPrefix(raw, "hash-quick "):
		return Line{Kind: KindHashQuick, Hash: raw[len("hash-quick "):]}, true, nil
	case strings.HasPrefix(raw, "hash "):
		return Line{Kind: KindHash, Hash: raw[len("hash "):]}, true, nil
	default:
		return Line{}, false, fmt.Errorf("commands: unknown line %q", raw)
	}
}
```

- [ ] **Step 2.3: Write `reader_test.go`**

```go
package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReaderSample(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "sample.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	r := New(f)
	want := []LineKind{KindStart, KindTurn, KindHash, KindHashQuick, KindTurn, KindCmd, KindCmd, KindTurn, KindCmd, KindEnd}
	var got []LineKind
	for {
		ln, ok, err := r.Next()
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		if !ok {
			break
		}
		got = append(got, ln.Kind)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d lines, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d: got %d, want %d", i, got[i], want[i])
		}
	}
}

func TestReaderTurnFields(t *testing.T) {
	f, _ := os.Open(filepath.Join("testdata", "sample.txt"))
	defer f.Close()
	r := New(f)
	r.Next() // start
	ln, _, _ := r.Next()
	if ln.Kind != KindTurn || ln.TurnN != 0 || ln.TickMs != 200 {
		t.Errorf("turn 0 200: got %+v", ln)
	}
}

func TestReaderCmdPlayer(t *testing.T) {
	f, _ := os.Open(filepath.Join("testdata", "sample.txt"))
	defer f.Close()
	r := New(f)
	for {
		ln, ok, err := r.Next()
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("no cmd found")
		}
		if ln.Kind == KindCmd {
			if ln.Player != 2 {
				t.Errorf("player = %d, want 2", ln.Player)
			}
			return
		}
	}
}
```

- [ ] **Step 2.4: Run, expect PASS**

```bash
go test ./internal/replay/commands/...
```

- [ ] **Step 2.5: Commit**

```bash
git add internal/replay/commands/
git commit -m "replay: line-based commands.txt reader"
```

---

## Task 3: metadata.json loader

**Files:**
- Create: `internal/replay/metadata/loader.go`
- Create: `internal/replay/metadata/loader_test.go`
- Create: `internal/replay/metadata/testdata/sample.json`

- [ ] **Step 3.1: Write fixture `testdata/sample.json`**

```json
{
  "timeElapsed": 1860400,
  "playerStates": [
    {
      "name": "gaia", "civ": "gaia", "state": "active",
      "color": {"r": 255, "g": 255, "b": 255},
      "popCount": 0, "popLimit": 0, "popMax": 0,
      "resourceCounts": {"food": 0, "wood": 0, "stone": 0, "metal": 0},
      "phase": "", "team": -1, "researchedTechs": []
    },
    {
      "name": "zeto", "civ": "spart", "state": "defeated",
      "color": {"r": 10, "g": 10, "b": 190},
      "popCount": 64, "popLimit": 200, "popMax": 200,
      "resourceCounts": {"food": 420, "wood": 380, "stone": 12, "metal": 50},
      "phase": "town", "team": 0,
      "researchedTechs": ["phase_town_generic", "gather_wicker_baskets"]
    }
  ]
}
```

(Включает только то, что код реально читает.)

- [ ] **Step 3.2: Write `loader.go`**

```go
// Package metadata loads metadata.json next to a 0ad replay.
package metadata

import (
	"encoding/json"
	"fmt"
	"os"
)

type Metadata struct {
	TimeElapsed  int64         `json:"timeElapsed"`
	PlayerStates []PlayerState `json:"playerStates"`
}

type PlayerState struct {
	Name             string            `json:"name"`
	Civ              string            `json:"civ"`
	State            string            `json:"state"`
	Phase            string            `json:"phase"`
	PopCount         int               `json:"popCount"`
	PopLimit         int               `json:"popLimit"`
	PopMax           int               `json:"popMax"`
	Team             int               `json:"team"`
	Color            Color             `json:"color"`
	ResourceCounts   map[string]int    `json:"resourceCounts"`
	ResearchedTechs  []string          `json:"researchedTechs"`
}

type Color struct{ R, G, B int }

// Load reads and parses metadata.json at path.
func Load(path string) (*Metadata, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("metadata: read %s: %w", path, err)
	}
	var m Metadata
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("metadata: parse %s: %w", path, err)
	}
	return &m, nil
}
```

- [ ] **Step 3.3: Write `loader_test.go`**

```go
package metadata

import (
	"path/filepath"
	"testing"
)

func TestLoadSample(t *testing.T) {
	m, err := Load(filepath.Join("testdata", "sample.json"))
	if err != nil {
		t.Fatal(err)
	}
	if m.TimeElapsed != 1860400 {
		t.Errorf("TimeElapsed = %d", m.TimeElapsed)
	}
	if len(m.PlayerStates) != 2 {
		t.Fatalf("PlayerStates len = %d", len(m.PlayerStates))
	}
	p := m.PlayerStates[1]
	if p.Civ != "spart" || p.State != "defeated" || p.Phase != "town" {
		t.Errorf("p[1] = %+v", p)
	}
	if p.ResourceCounts["food"] != 420 {
		t.Errorf("food = %d", p.ResourceCounts["food"])
	}
	if len(p.ResearchedTechs) != 2 {
		t.Errorf("techs = %v", p.ResearchedTechs)
	}
}
```

- [ ] **Step 3.4: Run, expect PASS**

```bash
go test ./internal/replay/metadata/...
```

- [ ] **Step 3.5: Commit**

```bash
git add internal/replay/metadata/
git commit -m "replay: metadata.json loader"
```

---

## Task 4: events.Decode — typed events from cmd lines

**Files:**
- Create: `internal/replay/events/events.go`
- Create: `internal/replay/events/decode.go`
- Create: `internal/replay/events/decode_test.go`

- [ ] **Step 4.1: Write `events.go`**

```go
// Package events typed-decodes 0ad replay command JSON.
package events

import "encoding/json"

const (
	TypeTrain      = "train"
	TypeConstruct  = "construct"
	TypeResearch   = "research"
	TypeAttack     = "attack"
	TypeGarrison   = "garrison"
	TypeTribute    = "tribute"
	TypeResign     = "resign"
	TypeChat       = "chat"
	TypeRaw        = "raw"
)

type Event struct {
	TMs    int64       // ms from game start
	Player int         // 1-based; 0 = gaia (won't occur)
	Type   string      // one of Type* constants
	Data   interface{} // type-specific payload below
}

type TrainData struct {
	Template string `json:"template"`
	Count    int    `json:"count"`
	Source   int    `json:"source,omitempty"` // entity ID training; 0 if from entities[]
	Queued   bool   `json:"queued"`
}

type ConstructData struct {
	Template string  `json:"template"`
	X, Z     float64 `json:"x,z"`
	Builders int     `json:"builders"`
}

type ResearchData struct {
	Template string `json:"template"`
	Source   int    `json:"source"`
}

type AttackData struct {
	Target       int  `json:"target"`
	Entities     int  `json:"entities"`
	AllowCapture bool `json:"allow_capture"`
}

type GarrisonData struct {
	Target   int `json:"target"`
	Entities int `json:"entities"`
}

type TributeData struct {
	To                       int `json:"to"`
	Food, Wood, Stone, Metal int
}

type ResignData struct{}

type ChatData struct {
	Text string `json:"text"`
}

type RawData struct {
	Kind string          `json:"kind"`
	JSON json.RawMessage `json:"json"`
}
```

- [ ] **Step 4.2: Write `decode.go`**

```go
package events

import (
	"encoding/json"
)

// Decode turns a single `cmd <P> {json}` line into a typed Event.
// player is 1-based; tMs is the cumulative ms from game start.
func Decode(player int, tMs int64, raw json.RawMessage) Event {
	var head struct {
		Type string `json:"type"`
	}
	_ = json.Unmarshal(raw, &head) // malformed JSON would have failed earlier
	ev := Event{TMs: tMs, Player: player, Type: head.Type}

	switch head.Type {
	case "train":
		var d struct {
			Template string `json:"template"`
			Count    int    `json:"count"`
			Entities []int  `json:"entities"`
			Queued   bool   `json:"queued"`
		}
		_ = json.Unmarshal(raw, &d)
		src := 0
		if len(d.Entities) > 0 {
			src = d.Entities[0]
		}
		ev.Data = TrainData{Template: d.Template, Count: d.Count, Source: src, Queued: d.Queued}
	case "construct":
		var d struct {
			Template string  `json:"template"`
			X        float64 `json:"x"`
			Z        float64 `json:"z"`
			Entities []int   `json:"entities"`
		}
		_ = json.Unmarshal(raw, &d)
		ev.Data = ConstructData{Template: d.Template, X: d.X, Z: d.Z, Builders: len(d.Entities)}
	case "research":
		var d struct {
			Template string `json:"template"`
			Entity   int    `json:"entity"`
		}
		_ = json.Unmarshal(raw, &d)
		ev.Data = ResearchData{Template: d.Template, Source: d.Entity}
	case "attack":
		var d struct {
			Target       int   `json:"target"`
			Entities     []int `json:"entities"`
			AllowCapture bool  `json:"allowCapture"`
		}
		_ = json.Unmarshal(raw, &d)
		ev.Data = AttackData{Target: d.Target, Entities: len(d.Entities), AllowCapture: d.AllowCapture}
	case "garrison":
		var d struct {
			Target   int   `json:"target"`
			Entities []int `json:"entities"`
		}
		_ = json.Unmarshal(raw, &d)
		ev.Data = GarrisonData{Target: d.Target, Entities: len(d.Entities)}
	case "tribute":
		var d struct {
			Player int            `json:"player"`
			Amounts map[string]int `json:"amounts"`
		}
		_ = json.Unmarshal(raw, &d)
		ev.Data = TributeData{
			To:    d.Player,
			Food:  d.Amounts["food"], Wood: d.Amounts["wood"],
			Stone: d.Amounts["stone"], Metal: d.Amounts["metal"],
		}
	case "resign":
		ev.Data = ResignData{}
	case "chat":
		var d struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(raw, &d)
		ev.Data = ChatData{Text: d.Message}
	default:
		ev.Type = TypeRaw
		ev.Data = RawData{Kind: head.Type, JSON: raw}
	}
	return ev
}
```

- [ ] **Step 4.3: Write `decode_test.go`**

```go
package events

import (
	"encoding/json"
	"testing"
)

func TestDecodeTrain(t *testing.T) {
	raw := json.RawMessage(`{"type":"train","template":"units/spart/inf_spearman_b","count":5,"entities":[2481],"queued":false}`)
	ev := Decode(2, 1200, raw)
	if ev.Type != "train" || ev.Player != 2 || ev.TMs != 1200 {
		t.Fatalf("ev = %+v", ev)
	}
	d := ev.Data.(TrainData)
	if d.Template != "units/spart/inf_spearman_b" || d.Count != 5 || d.Source != 2481 {
		t.Errorf("TrainData = %+v", d)
	}
}

func TestDecodeResearchPhase(t *testing.T) {
	raw := json.RawMessage(`{"type":"research","entity":2417,"template":"phase_town_generic","pushFront":false}`)
	ev := Decode(2, 295000, raw)
	d := ev.Data.(ResearchData)
	if d.Template != "phase_town_generic" || d.Source != 2417 {
		t.Errorf("ResearchData = %+v", d)
	}
}

func TestDecodeAttack(t *testing.T) {
	raw := json.RawMessage(`{"type":"attack","entities":[1,2,3],"target":99,"allowCapture":true}`)
	ev := Decode(2, 0, raw)
	d := ev.Data.(AttackData)
	if d.Target != 99 || d.Entities != 3 || !d.AllowCapture {
		t.Errorf("AttackData = %+v", d)
	}
}

func TestDecodeUnknownIsRaw(t *testing.T) {
	raw := json.RawMessage(`{"type":"walk","entities":[1],"x":10,"z":20}`)
	ev := Decode(2, 0, raw)
	if ev.Type != TypeRaw {
		t.Fatalf("type = %q, want raw", ev.Type)
	}
	d := ev.Data.(RawData)
	if d.Kind != "walk" {
		t.Errorf("Kind = %q", d.Kind)
	}
}

func TestDecodeResign(t *testing.T) {
	raw := json.RawMessage(`{"type":"resign"}`)
	ev := Decode(2, 1700000, raw)
	if ev.Type != "resign" {
		t.Errorf("type = %q", ev.Type)
	}
	if _, ok := ev.Data.(ResignData); !ok {
		t.Errorf("Data = %T", ev.Data)
	}
}
```

- [ ] **Step 4.4: Run, expect PASS**

```bash
go test ./internal/replay/events/...
```

- [ ] **Step 4.5: Commit**

```bash
git add internal/replay/events/
git commit -m "replay: typed cmd decoder"
```

---

## Task 5: Analytics (phases, engagements, panic_garrison, action density)

**Files:**
- Create: `internal/replay/analytics/phases.go`
- Create: `internal/replay/analytics/engagement.go`
- Create: `internal/replay/analytics/anomalies.go`
- Create: `internal/replay/analytics/density.go`
- Create: `internal/replay/analytics/analytics_test.go`

- [ ] **Step 5.1: Write `phases.go`**

```go
// Package analytics derives per-player metrics from a typed event stream.
package analytics

import (
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/events"
)

// PhaseTimings: per player, "town" → seconds when phase_town_* was researched.
// Returns the earliest research time per phase.
func PhaseTimings(evs []events.Event) map[int]map[string]int {
	out := map[int]map[string]int{}
	for _, e := range evs {
		if e.Type != events.TypeResearch {
			continue
		}
		d, ok := e.Data.(events.ResearchData)
		if !ok {
			continue
		}
		var phase string
		switch {
		case strings.HasPrefix(d.Template, "phase_town"):
			phase = "town"
		case strings.HasPrefix(d.Template, "phase_city"):
			phase = "city"
		default:
			continue
		}
		if _, exists := out[e.Player]; !exists {
			out[e.Player] = map[string]int{}
		}
		sec := int(e.TMs / 1000)
		if prev, ok := out[e.Player][phase]; !ok || sec < prev {
			out[e.Player][phase] = sec
		}
	}
	return out
}
```

- [ ] **Step 5.2: Write `engagement.go`**

```go
package analytics

import (
	"sort"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/events"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"
)

// Engagements groups consecutive attack commands on the same target by the same
// player; a new engagement starts when the gap to the previous attack exceeds windowMs.
func Engagements(evs []events.Event, windowMs int64) map[int][]output.Engagement {
	type key struct {
		player int
		target int
	}
	type cur struct {
		tStart, tEnd int64
		peak, count  int
	}
	open := map[key]*cur{}
	out := map[int][]output.Engagement{}

	flush := func(k key, c *cur) {
		out[k.player] = append(out[k.player], output.Engagement{
			TStartSec:    int(c.tStart / 1000),
			TEndSec:      int(c.tEnd / 1000),
			Target:       k.target,
			PeakUnits:    c.peak,
			CommandCount: c.count,
		})
	}

	// Stable order
	sorted := append([]events.Event{}, evs...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].TMs < sorted[j].TMs })

	for _, e := range sorted {
		if e.Type != events.TypeAttack {
			continue
		}
		d, ok := e.Data.(events.AttackData)
		if !ok {
			continue
		}
		k := key{player: e.Player, target: d.Target}
		c, exists := open[k]
		if exists && e.TMs-c.tEnd > windowMs {
			flush(k, c)
			exists = false
		}
		if !exists {
			c = &cur{tStart: e.TMs, tEnd: e.TMs, peak: d.Entities, count: 1}
			open[k] = c
			continue
		}
		c.tEnd = e.TMs
		c.count++
		if d.Entities > c.peak {
			c.peak = d.Entities
		}
	}
	for k, c := range open {
		flush(k, c)
	}
	for p := range out {
		es := out[p]
		sort.Slice(es, func(i, j int) bool { return es[i].TStartSec < es[j].TStartSec })
		out[p] = es
	}
	return out
}
```

- [ ] **Step 5.3: Write `anomalies.go`**

```go
package analytics

import (
	"sort"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/events"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"
)

// PanicGarrison flags ≥5 garrison commands into the same building within < 3 sec.
func PanicGarrison(evs []events.Event) map[int][]output.Anomaly {
	type key struct {
		player int
		target int
	}
	type win struct {
		ts []int64
	}
	seen := map[key]*win{}
	out := map[int][]output.Anomaly{}
	sorted := append([]events.Event{}, evs...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].TMs < sorted[j].TMs })

	const windowMs = 3000
	const threshold = 5

	for _, e := range sorted {
		if e.Type != events.TypeGarrison {
			continue
		}
		d, ok := e.Data.(events.GarrisonData)
		if !ok {
			continue
		}
		k := key{player: e.Player, target: d.Target}
		w, ok2 := seen[k]
		if !ok2 {
			w = &win{}
			seen[k] = w
		}
		w.ts = append(w.ts, e.TMs)
		// trim window
		cut := 0
		for cut < len(w.ts) && w.ts[len(w.ts)-1]-w.ts[cut] > windowMs {
			cut++
		}
		w.ts = w.ts[cut:]
		if len(w.ts) >= threshold {
			out[e.Player] = append(out[e.Player], output.Anomaly{
				Type:      "panic_garrison",
				TStartSec: int(w.ts[0] / 1000),
				TEndSec:   int(w.ts[len(w.ts)-1] / 1000),
				Severity:  "warning",
				Details: map[string]any{
					"target": d.Target,
					"count":  len(w.ts),
				},
			})
			w.ts = w.ts[:0] // reset so we don't double-fire
		}
	}
	return out
}
```

- [ ] **Step 5.4: Write `density.go`**

```go
package analytics

import (
	"sort"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/events"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"
)

// ActionDensity groups events into binSec-second buckets, classified by category.
// Categories: military (train/attack/garrison), economy (gather/tribute/barter via raw),
// build (construct), research (research), other.
func ActionDensity(evs []events.Event, binSec int) []output.DensityBin {
	if binSec <= 0 {
		binSec = 30
	}
	classify := func(e events.Event) string {
		switch e.Type {
		case events.TypeTrain, events.TypeAttack, events.TypeGarrison:
			return "military"
		case events.TypeConstruct:
			return "build"
		case events.TypeResearch:
			return "research"
		case events.TypeTribute:
			return "economy"
		case events.TypeRaw:
			d, ok := e.Data.(events.RawData)
			if !ok {
				return "other"
			}
			switch d.Kind {
			case "gather", "returnresource", "barter", "setup-trade-route":
				return "economy"
			default:
				return "other"
			}
		default:
			return "other"
		}
	}
	bins := map[int]map[string]int{}
	maxBin := 0
	for _, e := range evs {
		b := int(e.TMs/1000) / binSec
		if b > maxBin {
			maxBin = b
		}
		if bins[b] == nil {
			bins[b] = map[string]int{}
		}
		bins[b][classify(e)]++
	}
	out := make([]output.DensityBin, 0, maxBin+1)
	for b := 0; b <= maxBin; b++ {
		c := bins[b]
		if c == nil {
			c = map[string]int{}
		}
		out = append(out, output.DensityBin{TSec: b * binSec, Counts: c})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TSec < out[j].TSec })
	return out
}
```

- [ ] **Step 5.5: Write `analytics_test.go`**

```go
package analytics

import (
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/events"
)

func mkResearch(player int, tMs int64, tmpl string) events.Event {
	return events.Event{TMs: tMs, Player: player, Type: events.TypeResearch, Data: events.ResearchData{Template: tmpl}}
}
func mkAttack(player int, tMs int64, target, units int) events.Event {
	return events.Event{TMs: tMs, Player: player, Type: events.TypeAttack, Data: events.AttackData{Target: target, Entities: units}}
}
func mkGarrison(player int, tMs int64, target int) events.Event {
	return events.Event{TMs: tMs, Player: player, Type: events.TypeGarrison, Data: events.GarrisonData{Target: target, Entities: 1}}
}

func TestPhaseTimings(t *testing.T) {
	evs := []events.Event{
		mkResearch(2, 295000, "phase_town_generic"),
		mkResearch(2, 1547000, "phase_city_generic"),
		mkResearch(3, 320000, "phase_town_athen"),
		mkResearch(2, 200000, "gather_wicker_baskets"), // ignored
	}
	out := PhaseTimings(evs)
	if out[2]["town"] != 295 {
		t.Errorf("p2 town = %d", out[2]["town"])
	}
	if out[2]["city"] != 1547 {
		t.Errorf("p2 city = %d", out[2]["city"])
	}
	if out[3]["town"] != 320 {
		t.Errorf("p3 town = %d", out[3]["town"])
	}
}

func TestEngagementsGroups(t *testing.T) {
	evs := []events.Event{
		mkAttack(2, 10000, 99, 5),
		mkAttack(2, 11000, 99, 10), // within 3s window
		mkAttack(2, 12500, 99, 12),
		mkAttack(2, 20000, 99, 6),  // gap > 3s → new engagement
	}
	out := Engagements(evs, 3000)
	if len(out[2]) != 2 {
		t.Fatalf("got %d engagements, want 2: %+v", len(out[2]), out[2])
	}
	if out[2][0].PeakUnits != 12 || out[2][0].CommandCount != 3 {
		t.Errorf("eng[0] = %+v", out[2][0])
	}
}

func TestPanicGarrison(t *testing.T) {
	var evs []events.Event
	for i := int64(0); i < 5; i++ {
		evs = append(evs, mkGarrison(2, 10000+i*400, 77))
	}
	out := PanicGarrison(evs)
	if len(out[2]) != 1 {
		t.Fatalf("got %d anomalies, want 1: %+v", len(out[2]), out[2])
	}
	if out[2][0].Type != "panic_garrison" {
		t.Errorf("Type = %q", out[2][0].Type)
	}
}

func TestPanicGarrisonNoFalsePositive(t *testing.T) {
	evs := []events.Event{
		mkGarrison(2, 0, 77),
		mkGarrison(2, 4000, 77),
		mkGarrison(2, 8000, 77),
		mkGarrison(2, 12000, 77),
		mkGarrison(2, 16000, 77),
	}
	out := PanicGarrison(evs)
	if len(out[2]) != 0 {
		t.Fatalf("got %d, want 0 (gaps too large)", len(out[2]))
	}
}

func TestActionDensity(t *testing.T) {
	evs := []events.Event{
		mkResearch(2, 5000, "x"),                                              // bin 0 research
		mkAttack(2, 15000, 1, 1),                                              // bin 0 military
		mkResearch(2, 35000, "y"),                                             // bin 1 research
		{TMs: 65000, Player: 2, Type: events.TypeConstruct, Data: events.ConstructData{}}, // bin 2 build
	}
	out := ActionDensity(evs, 30)
	if len(out) != 3 {
		t.Fatalf("bins = %d, want 3", len(out))
	}
	if out[0].Counts["research"] != 1 || out[0].Counts["military"] != 1 {
		t.Errorf("bin0 = %+v", out[0].Counts)
	}
	if out[2].Counts["build"] != 1 {
		t.Errorf("bin2 = %+v", out[2].Counts)
	}
}
```

- [ ] **Step 5.6: Run, expect PASS**

```bash
go test ./internal/replay/analytics/...
```

- [ ] **Step 5.7: Commit**

```bash
git add internal/replay/analytics/
git commit -m "replay: phase/engagement/anomaly/density analytics"
```

---

## Task 6: Pipeline.Run + integration test on real replay

**Files:**
- Create: `internal/replay/pipeline.go`
- Create: `internal/replay/pipeline_test.go`
- Create: `testdata/replays/short-spart-vs-petra/commands.txt` (copy from user data; see step 6.1)
- Create: `testdata/replays/short-spart-vs-petra/metadata.json` (same)

- [ ] **Step 6.1: Copy a short real replay as a fixture**

```bash
# Pick the smallest commands.txt under user's replays
mkdir -p testdata/replays/short-spart-vs-petra
SRC="$HOME/Library/Application Support/0ad/replays/0.28.0"
SHORTEST=$(find "$SRC" -name commands.txt -size -100k -print 2>/dev/null | head -1 | xargs -I{} dirname {})
cp "$SHORTEST/commands.txt" testdata/replays/short-spart-vs-petra/
cp "$SHORTEST/metadata.json" testdata/replays/short-spart-vs-petra/
# Verify
ls -lh testdata/replays/short-spart-vs-petra/
```

If no replay <100KB exists, raise the size cap and trim `commands.txt` manually to ~200 lines preserving `start`/`end`.

- [ ] **Step 6.2: Write `pipeline.go`**

```go
// Package replay orchestrates the per-replay pipeline.
package replay

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/analytics"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/commands"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/events"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/metadata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"
)

// Run parses replayDir and returns Analysis. It writes analysis.json next to
// the replay; if analysis.json is newer than commands.txt it is reused.
func Run(replayDir string) (*output.Analysis, error) {
	cmdsPath := filepath.Join(replayDir, "commands.txt")
	metaPath := filepath.Join(replayDir, "metadata.json")
	outPath := filepath.Join(replayDir, "analysis.json")

	if _, err := os.Stat(metaPath); err != nil {
		return nil, fmt.Errorf("replay: %s: metadata.json missing (skipping)", replayDir)
	}

	if output.IsFresh(outPath, cmdsPath) {
		raw, err := os.ReadFile(outPath)
		if err == nil {
			var a output.Analysis
			if err := json.Unmarshal(raw, &a); err == nil && a.SchemaVersion == output.SchemaVersion {
				return &a, nil
			}
		}
	}

	meta, err := metadata.Load(metaPath)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(cmdsPath)
	if err != nil {
		return nil, fmt.Errorf("replay: open commands.txt: %w", err)
	}
	defer f.Close()

	game, players, err := parseStart(f)
	if err != nil {
		return nil, err
	}
	// Re-open for full streaming (parseStart consumed only the first line)
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	evs, durationMs, err := streamEvents(f)
	if err != nil {
		return nil, err
	}
	game.DurationMs = durationMs
	if meta.TimeElapsed > durationMs {
		game.DurationMs = meta.TimeElapsed
	}

	a := buildAnalysis(game, players, meta, evs)

	if err := output.Write(outPath, a); err != nil {
		return nil, fmt.Errorf("replay: write analysis: %w", err)
	}
	return a, nil
}

func parseStart(r io.Reader) (output.GameInfo, []output.Player, error) {
	rd := commands.New(r)
	ln, ok, err := rd.Next()
	if err != nil {
		return output.GameInfo{}, nil, err
	}
	if !ok || ln.Kind != commands.KindStart {
		return output.GameInfo{}, nil, fmt.Errorf("replay: first line is not 'start'")
	}
	var s struct {
		Settings struct {
			MapName           string `json:"mapName"`
			VictoryConditions []string `json:"VictoryConditions"`
			PlayerData        []struct {
				AI       any    `json:"AI"`
				AIDiff   int    `json:"AIDiff"`
				Civ      string `json:"Civ"`
				Name     string `json:"Name"`
				Team     int    `json:"Team"`
				Color    struct{ R, G, B int } `json:"Color"`
			} `json:"PlayerData"`
		} `json:"settings"`
		MatchID  string `json:"matchID"`
		Map      string `json:"map"`
		MapType  string `json:"mapType"`
		Timestamp int64 `json:"timestamp"`
		Mods []struct {
			Version string `json:"version"`
		} `json:"mods"`
	}
	if err := json.Unmarshal(ln.StartJSON, &s); err != nil {
		return output.GameInfo{}, nil, fmt.Errorf("replay: parse start: %w", err)
	}
	matchID := s.MatchID
	if matchID == "" {
		h := sha1.Sum(ln.StartJSON)
		matchID = hex.EncodeToString(h[:8])
	}
	ev := ""
	if len(s.Mods) > 0 {
		ev = s.Mods[0].Version
	}
	g := output.GameInfo{
		MatchID:       matchID,
		Map:           s.Settings.MapName,
		MapType:       s.MapType,
		Timestamp:     s.Timestamp,
		EngineVersion: ev,
		VictoryConds:  s.Settings.VictoryConditions,
	}
	players := make([]output.Player, 0, len(s.Settings.PlayerData))
	for i, pd := range s.Settings.PlayerData {
		_, isAI := pd.AI.(string)
		players = append(players, output.Player{
			ID:    i + 1,
			Name:  pd.Name,
			Civ:   pd.Civ,
			Team:  pd.Team,
			IsAI:  isAI,
			AIDiff: pd.AIDiff,
			Color: output.Color{R: pd.Color.R, G: pd.Color.G, B: pd.Color.B},
		})
	}
	return g, players, nil
}

func streamEvents(r io.Reader) ([]output.Event, int64, error) {
	rd := commands.New(r)
	var evs []output.Event
	var tMs int64
	for {
		ln, ok, err := rd.Next()
		if err != nil {
			return nil, 0, err
		}
		if !ok {
			break
		}
		switch ln.Kind {
		case commands.KindTurn:
			tMs += int64(ln.TickMs)
		case commands.KindCmd:
			ev := events.Decode(ln.Player, tMs, ln.CmdJSON)
			evs = append(evs, output.Event{
				T: ev.TMs, Player: ev.Player, Type: ev.Type, Data: ev.Data,
			})
		}
	}
	return evs, tMs, nil
}

// internalEvents reconstructs the typed events.Event slice for analytics.
func internalEvents(evs []output.Event) []events.Event {
	out := make([]events.Event, len(evs))
	for i, e := range evs {
		out[i] = events.Event{TMs: e.T, Player: e.Player, Type: e.Type, Data: e.Data}
	}
	return out
}

func buildAnalysis(g output.GameInfo, players []output.Player, m *metadata.Metadata, evs []output.Event) *output.Analysis {
	tev := internalEvents(evs)
	phaseT := analytics.PhaseTimings(tev)
	eng := analytics.Engagements(tev, 3000)
	pg := analytics.PanicGarrison(tev)
	density := analytics.ActionDensity(tev, 30)

	resignByPlayer := map[int]bool{}
	for _, e := range tev {
		if e.Type == events.TypeResign {
			resignByPlayer[e.Player] = true
		}
	}

	finalByPlayer := map[int]output.PlayerFinalState{}
	for i, ps := range m.PlayerStates {
		if i == 0 {
			continue // gaia
		}
		outcome := ps.State
		if resignByPlayer[i] {
			outcome = "defeated"
		}
		rc := output.Resources{
			Food:  ps.ResourceCounts["food"],
			Wood:  ps.ResourceCounts["wood"],
			Stone: ps.ResourceCounts["stone"],
			Metal: ps.ResourceCounts["metal"],
		}
		finalByPlayer[i] = output.PlayerFinalState{
			Phase:           ps.Phase,
			State:           ps.State,
			Outcome:         outcome,
			PopCount:        ps.PopCount,
			PopLimit:        ps.PopLimit,
			PopMax:          ps.PopMax,
			ResourceCounts:  rc,
			ResearchedTechs: ps.ResearchedTechs,
		}
	}

	metricsByPlayer := map[int]output.PlayerMetrics{}
	allPlayers := map[int]struct{}{}
	for p := range phaseT {
		allPlayers[p] = struct{}{}
	}
	for p := range eng {
		allPlayers[p] = struct{}{}
	}
	for p := range pg {
		allPlayers[p] = struct{}{}
	}
	for p := range allPlayers {
		metricsByPlayer[p] = output.PlayerMetrics{
			PhaseTimings: phaseT[p],
			Engagements:  eng[p],
			Anomalies:    pg[p],
		}
	}

	return &output.Analysis{
		SchemaVersion: output.SchemaVersion,
		Game:          g,
		Players:       players,
		Events:        evs,
		Snapshots:     []output.Snapshot{},
		FinalState:    output.FinalState{Players: finalByPlayer},
		Metrics:       output.Metrics{Players: metricsByPlayer, Density: density},
	}
}

// Outcome returns a human label for a player given final state and resign events.
func Outcome(ps metadata.PlayerState, resigned bool) string {
	switch {
	case resigned:
		return "defeated"
	case strings.EqualFold(ps.State, "won"):
		return "won"
	case strings.EqualFold(ps.State, "defeated"):
		return "defeated"
	default:
		return ps.State
	}
}
```

- [ ] **Step 6.3: Write `pipeline_test.go`**

```go
package replay

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunOnRealFixture(t *testing.T) {
	dir := filepath.Join("..", "..", "testdata", "replays", "short-spart-vs-petra")
	if _, err := os.Stat(filepath.Join(dir, "metadata.json")); err != nil {
		t.Skipf("fixture missing: %v", err)
	}
	defer os.Remove(filepath.Join(dir, "analysis.json"))

	a, err := Run(dir)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if a.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d", a.SchemaVersion)
	}
	if a.Game.MatchID == "" {
		t.Error("MatchID empty")
	}
	if len(a.Players) == 0 {
		t.Error("no players")
	}
	if a.Game.DurationMs <= 0 {
		t.Errorf("DurationMs = %d", a.Game.DurationMs)
	}
	if len(a.Events) == 0 {
		t.Error("no events decoded")
	}
	// Sanity: phase timings present for at least one player (any short replay should reach town)
	hasPhase := false
	for _, m := range a.Metrics.Players {
		if len(m.PhaseTimings) > 0 {
			hasPhase = true
		}
	}
	if !hasPhase {
		t.Log("WARN: no phase timings (replay may be very short)")
	}
}

func TestRunSkipsMissingMetadata(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "commands.txt"), []byte("start {}\nend\n"), 0o644)
	_, err := Run(dir)
	if err == nil {
		t.Fatal("Run must error on missing metadata.json")
	}
}
```

- [ ] **Step 6.4: Run, expect PASS**

```bash
go test ./internal/replay/...
```

- [ ] **Step 6.5: Commit**

```bash
git add internal/replay/pipeline.go internal/replay/pipeline_test.go testdata/replays/
git commit -m "replay: pipeline orchestration + integration fixture"
```

---

## Task 7: CLI `replayreport` — scan mode (no server yet)

**Files:**
- Create: `cmd/replayreport/main.go`
- Modify: `Makefile`

- [ ] **Step 7.1: Write `cmd/replayreport/main.go`**

```go
// Command replayreport parses 0ad replays and writes analysis.json next to each.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay"
)

const defaultReplayDir = "/Users/zeto/Library/Application Support/0ad/replays/0.28.0"

func main() {
	var (
		all     bool
		check   bool
		repDir  string
	)
	flag.BoolVar(&all, "all", false, "process every replay subdir under replay root")
	flag.BoolVar(&check, "check", false, "validate replays exit with non-zero on any failure (no http)")
	flag.StringVar(&repDir, "replays", defaultReplayDir, "replay root (used when no positional arg)")
	flag.Parse()

	if flag.NArg() == 1 && !all {
		if err := runOne(flag.Arg(0)); err != nil {
			fmt.Fprintln(os.Stderr, "ERROR:", err)
			os.Exit(1)
		}
		return
	}

	root := repDir
	if flag.NArg() == 1 && all {
		root = flag.Arg(0)
	}
	if check || all {
		runScan(root, check)
		return
	}

	// No args, no flags: scan and (Task 8) launch HTTP. For now just scan.
	runScan(root, false)
}

func runOne(dir string) error {
	a, err := replay.Run(dir)
	if err != nil {
		return err
	}
	fmt.Printf("OK %s — %s (%s, %d events)\n", a.Game.MatchID, a.Game.Map, dir, len(a.Events))
	return nil
}

func runScan(root string, strict bool) {
	entries, err := os.ReadDir(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scan:", err)
		os.Exit(1)
	}
	ok, skipped, failed := 0, 0, 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		if _, err := os.Stat(filepath.Join(dir, "metadata.json")); err != nil {
			skipped++
			continue
		}
		if _, err := replay.Run(dir); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", e.Name(), err)
			failed++
			continue
		}
		ok++
	}
	fmt.Printf("scan: %d ok, %d skipped (no metadata), %d failed\n", ok, skipped, failed)
	if strict && failed > 0 {
		os.Exit(2)
	}
}
```

- [ ] **Step 7.2: Add Makefile targets**

Append to existing `Makefile`:

```make
.PHONY: replayreport replay-check
replayreport:
	go build -o bin/replayreport ./cmd/replayreport

replay-check: replayreport
	./bin/replayreport --check --all
```

- [ ] **Step 7.3: Build + smoke against real replays**

```bash
make replayreport
./bin/replayreport testdata/replays/short-spart-vs-petra
./bin/replayreport --check --all
```

Expected: одна `OK ...` строка для fixture; `scan: N ok, M skipped, 0 failed` для пользовательских реплеев.

- [ ] **Step 7.4: Run all tests**

```bash
go test ./...
```

- [ ] **Step 7.5: Commit**

```bash
git add cmd/replayreport/ Makefile
git commit -m "replayreport: CLI scan mode + analysis.json caching"
```

---

## Task 8: Web UI — list page + chart-less analysis page

**Files:**
- Create: `internal/replay/webui/server.go`
- Create: `internal/replay/webui/handlers.go`
- Create: `internal/replay/webui/templates/layout.templ`
- Create: `internal/replay/webui/templates/index.templ`
- Create: `internal/replay/webui/templates/replay.templ`
- Create: `internal/replay/webui/static/style.css`
- Create: `internal/replay/webui/static/htmx.min.js` (vendored)
- Modify: `go.mod`, `go.sum` (add `github.com/a-h/templ`)
- Modify: `cmd/replayreport/main.go` (default mode = serve)

This task **does not** add Plotly yet — chart is plain HTML in Task 9. We end Task 8 with a navigable site backed by parsed analysis files.

- [ ] **Step 8.1: Add templ dependency + tool**

```bash
go get github.com/a-h/templ@latest
go install github.com/a-h/templ/cmd/templ@latest
```

Confirm `templ` is on PATH (`which templ`).

- [ ] **Step 8.2: Vendor htmx**

```bash
mkdir -p internal/replay/webui/static
curl -sSL -o internal/replay/webui/static/htmx.min.js \
  https://unpkg.com/htmx.org@1.9.12/dist/htmx.min.js
ls -lh internal/replay/webui/static/htmx.min.js
```

Expect file ~50KB. If `curl` is unavailable in env, request user download manually and place at that path.

- [ ] **Step 8.3: Write `templates/layout.templ`**

```go
package templates

templ Layout(title string) {
	<!DOCTYPE html>
	<html lang="ru">
	<head>
		<meta charset="utf-8"/>
		<title>{ title }</title>
		<link rel="stylesheet" href="/static/style.css"/>
		<script src="/static/htmx.min.js"></script>
	</head>
	<body>
		<header><a href="/">replayreport</a></header>
		<main>
			{ children... }
		</main>
	</body>
	</html>
}
```

- [ ] **Step 8.4: Write `templates/index.templ`**

```go
package templates

import "fmt"

import "github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"

templ Index(items []ReplayCard) {
	@Layout("0ad replays") {
		<h1>Партии</h1>
		<ul class="cards">
			for _, it := range items {
				<li class="card">
					<a href={ templ.SafeURL(fmt.Sprintf("/replay/%s", it.MatchID)) }>
						<div class="card-title">{ it.Map }</div>
						<div class="card-meta">{ it.When } · { it.Duration }</div>
						<div class="card-players">
							for _, p := range it.Players {
								<span class="chip" style={ fmt.Sprintf("background:rgb(%d,%d,%d)", p.Color.R, p.Color.G, p.Color.B) }>
									{ p.Name } ({ p.Civ })
									if p.IsAI { · AI }
								</span>
							}
						</div>
						<div class="card-outcome">{ it.Outcome }</div>
					</a>
				</li>
			}
		</ul>
	}
}

type ReplayCard struct {
	MatchID  string
	Map      string
	When     string
	Duration string
	Players  []output.Player
	Outcome  string
}
```

- [ ] **Step 8.5: Write `templates/replay.templ` (no charts yet)**

```go
package templates

import "fmt"

import "github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"

templ Replay(a *output.Analysis, buildOrder []BuildOrderRow, anomalies []AnomalyRow) {
	@Layout(fmt.Sprintf("Replay %s", a.Game.MatchID)) {
		<h1>{ a.Game.Map }</h1>
		<section class="header-meta">
			<div>matchID: { a.Game.MatchID }</div>
			<div>длительность: { fmt.Sprintf("%d мин", a.Game.DurationMs/60000) }</div>
			<div>движок: { a.Game.EngineVersion }</div>
		</section>

		<section class="players">
			for _, p := range a.Players {
				<div class="player-row">
					<span class="chip" style={ fmt.Sprintf("background:rgb(%d,%d,%d)", p.Color.R, p.Color.G, p.Color.B) }>
						{ p.Name } · { p.Civ }
						if p.IsAI { · Petra { fmt.Sprintf("(d%d)", p.AIDiff) } }
					</span>
					if fs, ok := a.FinalState.Players[p.ID]; ok {
						<span class="outcome">{ fs.Outcome }</span>
						<span class="phase">фаза: { fs.Phase }</span>
						<span class="pop">поп: { fmt.Sprintf("%d/%d", fs.PopCount, fs.PopLimit) }</span>
					}
				</div>
			}
		</section>

		<section class="density-placeholder">
			<h2>Плотность действий (30-сек бины)</h2>
			<p>График в Task 9. Пока — таблица:</p>
			<table>
				<tr><th>t (сек)</th><th>military</th><th>build</th><th>research</th><th>economy</th><th>other</th></tr>
				for _, b := range a.Metrics.Density {
					<tr>
						<td>{ fmt.Sprint(b.TSec) }</td>
						<td>{ fmt.Sprint(b.Counts["military"]) }</td>
						<td>{ fmt.Sprint(b.Counts["build"]) }</td>
						<td>{ fmt.Sprint(b.Counts["research"]) }</td>
						<td>{ fmt.Sprint(b.Counts["economy"]) }</td>
						<td>{ fmt.Sprint(b.Counts["other"]) }</td>
					</tr>
				}
			</table>
		</section>

		<aside class="build-order">
			<h2>Build order</h2>
			<table>
				<tr><th>t</th><th>player</th><th>event</th></tr>
				for _, r := range buildOrder {
					<tr><td>{ r.Time }</td><td>{ r.Player }</td><td>{ r.Event }</td></tr>
				}
			</table>
		</aside>

		<section class="anomalies">
			<h2>Аномалии</h2>
			if len(anomalies) == 0 {
				<p class="empty">Чисто</p>
			} else {
				<ul>
					for _, an := range anomalies {
						<li>
							<strong>{ an.Type }</strong>
							{ fmt.Sprintf(" %s · t=%d..%d · %s", an.Severity, an.TStart, an.TEnd, an.Detail) }
						</li>
					}
				</ul>
			}
		</section>
	}
}

type BuildOrderRow struct {
	Time   string
	Player string
	Event  string
}

type AnomalyRow struct {
	Type, Severity, Detail string
	TStart, TEnd           int
}
```

- [ ] **Step 8.6: Generate templ Go files**

```bash
templ generate ./internal/replay/webui/templates/...
ls internal/replay/webui/templates/
```

Expect `*_templ.go` files alongside `.templ`.

- [ ] **Step 8.7: Write `handlers.go`**

```go
package webui

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/events"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/webui/templates"
)

//go:embed static/*
var staticFS embed.FS

type handlers struct {
	repRoot string
}

func (h *handlers) index(w http.ResponseWriter, r *http.Request) {
	cards := h.loadAllCards()
	templates.Index(cards).Render(r.Context(), w)
}

func (h *handlers) replay(w http.ResponseWriter, r *http.Request) {
	matchID := strings.TrimPrefix(r.URL.Path, "/replay/")
	a, dir, err := h.findByMatchID(matchID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = dir
	bo := buildOrderRows(a)
	an := anomalyRows(a)
	templates.Replay(a, bo, an).Render(r.Context(), w)
}

func (h *handlers) loadAllCards() []templates.ReplayCard {
	entries, _ := os.ReadDir(h.repRoot)
	var cards []templates.ReplayCard
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(h.repRoot, e.Name())
		if _, err := os.Stat(filepath.Join(dir, "metadata.json")); err != nil {
			continue
		}
		a, err := replay.Run(dir)
		if err != nil {
			continue
		}
		outcome := "—"
		if fs, ok := a.FinalState.Players[1]; ok {
			outcome = fs.Outcome
		}
		cards = append(cards, templates.ReplayCard{
			MatchID:  a.Game.MatchID,
			Map:      a.Game.Map,
			When:     time.Unix(a.Game.Timestamp, 0).Format("02 Jan 15:04"),
			Duration: fmt.Sprintf("%d мин", a.Game.DurationMs/60000),
			Players:  a.Players,
			Outcome:  outcome,
		})
	}
	sort.Slice(cards, func(i, j int) bool { return cards[i].When > cards[j].When })
	return cards
}

func (h *handlers) findByMatchID(matchID string) (*output.Analysis, string, error) {
	entries, _ := os.ReadDir(h.repRoot)
	for _, e := range entries {
		dir := filepath.Join(h.repRoot, e.Name())
		path := filepath.Join(dir, "analysis.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var a output.Analysis
		if err := json.Unmarshal(raw, &a); err != nil {
			continue
		}
		if a.Game.MatchID == matchID {
			return &a, dir, nil
		}
	}
	return nil, "", fmt.Errorf("not found")
}

func buildOrderRows(a *output.Analysis) []templates.BuildOrderRow {
	playerName := map[int]string{}
	for _, p := range a.Players {
		playerName[p.ID] = p.Name
	}
	var rows []templates.BuildOrderRow
	for _, e := range a.Events {
		var label string
		switch e.Type {
		case events.TypeResearch:
			if d, ok := e.Data.(events.ResearchData); ok {
				label = "research " + d.Template
			} else if m, ok := e.Data.(map[string]any); ok {
				if t, _ := m["template"].(string); t != "" {
					label = "research " + t
				}
			}
		case events.TypeConstruct:
			if d, ok := e.Data.(events.ConstructData); ok {
				label = "construct " + d.Template
			} else if m, ok := e.Data.(map[string]any); ok {
				if t, _ := m["template"].(string); t != "" {
					label = "construct " + t
				}
			}
		case events.TypeResign:
			label = "RESIGN"
		default:
			continue
		}
		rows = append(rows, templates.BuildOrderRow{
			Time:   fmt.Sprintf("%02d:%02d", e.T/60000, (e.T/1000)%60),
			Player: playerName[e.Player],
			Event:  label,
		})
	}
	return rows
}

func anomalyRows(a *output.Analysis) []templates.AnomalyRow {
	var rows []templates.AnomalyRow
	for _, m := range a.Metrics.Players {
		for _, an := range m.Anomalies {
			detail := ""
			if t, ok := an.Details["target"]; ok {
				detail = fmt.Sprintf("target=%v", t)
			}
			rows = append(rows, templates.AnomalyRow{
				Type:     an.Type,
				Severity: an.Severity,
				TStart:   an.TStartSec,
				TEnd:     an.TEndSec,
				Detail:   detail,
			})
		}
	}
	return rows
}
```

**Note on `e.Data`:** because `output.Event.Data` is typed `any` and JSON-decoded back into `map[string]any` on cache read, handlers must accept both the typed struct (fresh run) and the map shape (loaded from analysis.json). Both branches are above.

- [ ] **Step 8.8: Write `server.go`**

```go
package webui

import (
	"net/http"
)

func NewServer(repRoot string) http.Handler {
	h := &handlers{repRoot: repRoot}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			h.index(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/replay/", h.replay)
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))
	return mux
}
```

- [ ] **Step 8.9: Write `static/style.css`**

```css
body { font-family: -apple-system, system-ui, sans-serif; margin: 0; background: #fafafa; color: #111; }
header { padding: 12px 24px; background: #222; color: #fff; }
header a { color: #fff; text-decoration: none; font-weight: 600; }
main { max-width: 1100px; margin: 24px auto; padding: 0 24px; }

.cards { list-style: none; padding: 0; display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 12px; }
.card { background: #fff; border: 1px solid #ddd; border-radius: 8px; padding: 14px; }
.card a { text-decoration: none; color: inherit; display: block; }
.card-title { font-weight: 600; font-size: 1.05em; }
.card-meta { color: #888; font-size: 0.85em; margin: 4px 0 8px; }
.card-players { display: flex; flex-wrap: wrap; gap: 4px; margin-bottom: 8px; }
.chip { padding: 2px 8px; border-radius: 999px; color: #fff; font-size: 0.85em; text-shadow: 0 1px 1px rgba(0,0,0,.3); }
.card-outcome { font-size: 0.85em; color: #555; }

.header-meta { display: flex; gap: 24px; color: #555; font-size: 0.9em; margin-bottom: 12px; }
.player-row { display: flex; align-items: center; gap: 12px; margin: 4px 0; }
.player-row .outcome { font-weight: 600; }
section.density-placeholder, aside.build-order, section.anomalies { background: #fff; padding: 16px; border-radius: 8px; margin: 16px 0; border: 1px solid #ddd; }
table { width: 100%; border-collapse: collapse; font-size: 0.9em; }
th, td { padding: 4px 8px; border-bottom: 1px solid #eee; text-align: left; }
.empty { color: #888; font-style: italic; }
```

- [ ] **Step 8.10: Wire up CLI default mode = serve**

Edit `cmd/replayreport/main.go`:

Replace the `// No args, no flags: scan and (Task 8) launch HTTP. For now just scan.` block with:

```go
	addr := ":8080"
	fmt.Printf("scanning %s …\n", root)
	scanCountOnly(root)
	fmt.Printf("serving http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, webui.NewServer(root)); err != nil {
		fmt.Fprintln(os.Stderr, "http:", err)
		os.Exit(1)
	}
```

Add the supporting helper at the bottom of `main.go`:

```go
func scanCountOnly(root string) {
	entries, _ := os.ReadDir(root)
	ok, skipped, failed := 0, 0, 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		if _, err := os.Stat(filepath.Join(dir, "metadata.json")); err != nil {
			skipped++
			continue
		}
		if _, err := replay.Run(dir); err != nil {
			failed++
			continue
		}
		ok++
	}
	fmt.Printf("scan: %d ok, %d skipped, %d failed\n", ok, skipped, failed)
}
```

And add imports `"net/http"` and `"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/webui"`.

- [ ] **Step 8.11: Build, smoke-run**

```bash
templ generate ./internal/replay/webui/templates/...
make replayreport
./bin/replayreport &
SERVER_PID=$!
sleep 1
curl -s http://localhost:8080/ | head -30
kill $SERVER_PID
```

Expected: HTML с `<h1>Партии</h1>` и хотя бы одной карточкой.

- [ ] **Step 8.12: Run all tests**

```bash
go test ./...
```

- [ ] **Step 8.13: Commit**

```bash
git add internal/replay/webui/ cmd/replayreport/main.go go.mod go.sum
git commit -m "replayreport: web ui — index + replay page (no charts)"
```

---

## Task 9: Plotly action-density chart on replay page

**Files:**
- Create: `internal/replay/webui/static/plotly.min.js` (vendored)
- Modify: `internal/replay/webui/handlers.go` (add chart-data JSON helper)
- Modify: `internal/replay/webui/templates/replay.templ` (replace placeholder with `<div id="chart">` + inline JSON + Plotly init)

- [ ] **Step 9.1: Vendor Plotly-JS**

```bash
curl -sSL -o internal/replay/webui/static/plotly.min.js \
  https://cdn.plot.ly/plotly-2.32.0.min.js
ls -lh internal/replay/webui/static/plotly.min.js
```

Expect ~3.5MB. If size matters, swap for `plotly-basic-*.min.js` (~1MB) — basic includes bar charts.

- [ ] **Step 9.2: Replace placeholder section in `replay.templ`**

Find `<section class="density-placeholder">…</section>` and replace with:

```go
		<section class="density-chart">
			<h2>Плотность действий</h2>
			<div id="chart" style="width:100%;height:360px"></div>
			<script>
				const DENSITY = { @templ.JSONScript("density-data", chartData) };
				const PHASES = { @templ.JSONScript("phase-data", phaseMarkers) };
				const ENGS = { @templ.JSONScript("eng-data", engagementMarkers) };
			</script>
			<script src="/static/plotly.min.js"></script>
			<script src="/static/replay-chart.js"></script>
		</section>
```

Wait — templ doesn't have `JSONScript`. Use a simpler approach: emit JSON into hidden `<script type="application/json" id="...">` blocks; init JS reads them.

Replace with:

```go
		<section class="density-chart">
			<h2>Плотность действий (30 сек)</h2>
			<div id="chart" style="width:100%;height:360px"></div>
			<script type="application/json" id="density-data">{ chartDataJSON }</script>
			<script type="application/json" id="phase-data">{ phaseMarkersJSON }</script>
			<script type="application/json" id="eng-data">{ engagementMarkersJSON }</script>
			<script src="/static/plotly.min.js"></script>
			<script src="/static/replay-chart.js"></script>
		</section>
```

And change the `templ Replay(...)` signature to accept those three `string` JSON parameters.

Updated template signature:

```go
templ Replay(a *output.Analysis, buildOrder []BuildOrderRow, anomalies []AnomalyRow,
	chartDataJSON, phaseMarkersJSON, engagementMarkersJSON string) {
```

- [ ] **Step 9.3: Compute JSON in handlers.go**

Inside `(h *handlers) replay(...)`:

```go
	chartData := buildDensityChartData(a)
	phaseMarkers := buildPhaseMarkers(a)
	engMarkers := buildEngagementMarkers(a)
	chartJSON, _ := json.Marshal(chartData)
	phaseJSON, _ := json.Marshal(phaseMarkers)
	engJSON, _ := json.Marshal(engMarkers)
	templates.Replay(a, bo, an, string(chartJSON), string(phaseJSON), string(engJSON)).Render(r.Context(), w)
```

Add helpers (still in handlers.go):

```go
type densityTrace struct {
	Name string `json:"name"`
	X    []int  `json:"x"`
	Y    []int  `json:"y"`
}

type phaseMarker struct {
	X     int    `json:"x"`
	Label string `json:"label"`
}

type engMarker struct {
	X    int `json:"x"`
	Size int `json:"size"`
}

func buildDensityChartData(a *output.Analysis) []densityTrace {
	cats := []string{"military", "build", "research", "economy", "other"}
	x := make([]int, len(a.Metrics.Density))
	for i, b := range a.Metrics.Density {
		x[i] = b.TSec
	}
	out := make([]densityTrace, 0, len(cats))
	for _, c := range cats {
		y := make([]int, len(a.Metrics.Density))
		for i, b := range a.Metrics.Density {
			y[i] = b.Counts[c]
		}
		out = append(out, densityTrace{Name: c, X: x, Y: y})
	}
	return out
}

func buildPhaseMarkers(a *output.Analysis) []phaseMarker {
	var out []phaseMarker
	for _, m := range a.Metrics.Players {
		for name, t := range m.PhaseTimings {
			out = append(out, phaseMarker{X: t, Label: name})
		}
	}
	return out
}

func buildEngagementMarkers(a *output.Analysis) []engMarker {
	const minPeak = 5
	var out []engMarker
	for _, m := range a.Metrics.Players {
		for _, e := range m.Engagements {
			if e.PeakUnits < minPeak {
				continue
			}
			out = append(out, engMarker{X: e.TStartSec, Size: e.PeakUnits})
		}
	}
	return out
}
```

- [ ] **Step 9.4: Write `static/replay-chart.js`**

```js
(function () {
	const density = JSON.parse(document.getElementById("density-data").textContent);
	const phases = JSON.parse(document.getElementById("phase-data").textContent);
	const engs = JSON.parse(document.getElementById("eng-data").textContent);

	const traces = density.map((d) => ({
		name: d.name,
		type: "bar",
		x: d.x,
		y: d.y,
	}));

	const shapes = phases.map((p) => ({
		type: "line", xref: "x", yref: "paper",
		x0: p.x, x1: p.x, y0: 0, y1: 1,
		line: { dash: "dash", width: 1, color: "#555" },
	})).concat(engs.map((e) => ({
		type: "line", xref: "x", yref: "paper",
		x0: e.x, x1: e.x, y0: 0, y1: 1,
		line: { dash: "solid", width: Math.min(1 + Math.log2(e.size), 4), color: "rgba(220,0,0,0.5)" },
	})));

	const annotations = phases.map((p) => ({
		x: p.x, y: 1, yref: "paper",
		text: p.label, showarrow: false,
		font: { size: 10, color: "#555" },
	}));

	Plotly.newPlot("chart", traces, {
		barmode: "stack",
		margin: { t: 24, r: 16, l: 40, b: 32 },
		xaxis: { title: "сек" },
		yaxis: { title: "команд / 30 сек" },
		shapes,
		annotations,
		legend: { orientation: "h", y: -0.2 },
	}, { responsive: true });
})();
```

- [ ] **Step 9.5: Regenerate templ, build, smoke**

```bash
templ generate ./internal/replay/webui/templates/...
make replayreport
./bin/replayreport &
SERVER_PID=$!
sleep 1
curl -s http://localhost:8080/ -o /tmp/index.html
curl -s "http://localhost:8080/replay/$(jq -r .game.match_id testdata/replays/short-spart-vs-petra/analysis.json)" -o /tmp/replay.html
grep -c 'id="chart"' /tmp/replay.html
grep -c 'density-data' /tmp/replay.html
kill $SERVER_PID
```

Expect `1` for both greps. Open in browser locally if you want eyeballs on the chart.

- [ ] **Step 9.6: Run all tests**

```bash
go test ./...
```

- [ ] **Step 9.7: Commit**

```bash
git add internal/replay/webui/
git commit -m "replayreport: action density chart with phase + engagement markers"
```

---

## Task 10: README + verification on full replay corpus

**Files:**
- Modify: `README.md` (add `replayreport` section)
- Modify: `CLAUDE.md` (one-paragraph note about new bin + caveat re: sequences)

- [ ] **Step 10.1: Update README**

Append section to `README.md`:

```markdown
## replayreport

`replayreport` парсит реплеи 0 A.D. и поднимает локальный дашик.

```bash
make replayreport
./bin/replayreport                     # сканит ~/Library/Application Support/0ad/replays/0.28.0/
                                       # → http://localhost:8080
./bin/replayreport <replay-dir>        # парсит один replay
./bin/replayreport --check --all       # CI-mode: exit 2 если хоть один реплей упал
```

Для каждого replay-dir пишется `analysis.json` рядом с `commands.txt` (кэш по mtime).
Replay-dir без `metadata.json` пропускаются (игра упала до summary screen).
```

- [ ] **Step 10.2: Update CLAUDE.md project notes**

Add a paragraph under "Стек" or a new "Replay analyzer" section:

```markdown
## Replay analyzer (`cmd/replayreport`, `internal/replay/`)

Vertical MVP. CLI `replayreport` парсит `commands.txt` + `metadata.json`
из replay-dir, пишет `analysis.json` (schema v1) рядом, поднимает дашик
на :8080. Time-series графики **не реализованы** в v1: `metadata.json`
реальных пользовательских реплеев не содержит `sequences` (см.
`docs/superpowers/specs/2026-05-28-replay-analyzer-mvp-design.md` и
discovery в `docs/superpowers/plans/2026-05-28-replay-analyzer-mvp.md`).
Замена — action-density chart по command stream. Reducer-симуляция
снапшотов отложена в фазу 3 (real-time).
```

- [ ] **Step 10.3: Verification — full corpus smoke**

```bash
./bin/replayreport --check --all
```

Expected: `scan: N ok, M skipped, 0 failed` где `N+M ≈ 86`. Если `failed > 0` — починить прежде чем мерджить.

- [ ] **Step 10.4: Final commit**

```bash
git add README.md CLAUDE.md
git commit -m "docs: replayreport usage + sequences caveat"
```

- [ ] **Step 10.5: Run `verification-before-completion` skill checklist**

Before opening PR, run through `superpowers:verification-before-completion`:
- `go test ./...` → all PASS
- `./bin/replayreport --check --all` → 0 failed
- `./bin/replayreport` → manual browser smoke at `localhost:8080`
- `git status` clean
- `git log --oneline master..replay-analyzer-mvp` shows clean commit history

---

## Self-Review

**Spec coverage:**

| Spec section | Task(s) |
|---|---|
| `commands.txt` reader | Task 2 |
| `metadata.json` loader | Task 3 |
| Typed events (train/construct/research/attack/garrison/tribute/resign/chat + raw) | Task 4 |
| Snapshots-from-metadata.sequences | **Dropped (discovery)** — replaced with action density |
| `phase_timings` | Task 5 |
| `engagements` | Task 5 |
| `pop_crash` anomaly | **Dropped** (needs sequences) |
| `resource_stall` anomaly | **Dropped** (needs sequences) |
| `panic_garrison` anomaly | Task 5 |
| `analysis.json` schema v1 + caching | Task 1, Task 6 |
| CLI `replayreport <dir>` / no-args / `--all` | Task 7, Task 8 |
| Index page (cards) | Task 8 |
| Replay page (header, build order, anomalies, final state) | Task 8 |
| Replay page (action density chart with phase + eng markers) | Task 9 |
| `Chart 2: Resources gathered`, `Chart 3: Active gatherers` | **Dropped** (no sequences); replaced by density chart |
| `testdata/replays/` short fixture | Task 6 |
| Discovery item: Petra AI commands | Manual inspect in Task 6 (note in plan) |
| Discovery item: tick_ms accumulation | Implemented in `streamEvents` (Task 6) — accumulates `tick_ms`, not `turn*200` |
| Discovery item: outcome detection | Implemented in `buildAnalysis` (Task 6) — resign event takes precedence |
| README / CLAUDE.md update | Task 10 |

**Placeholder scan:** No "TBD", no "add appropriate error handling", every code block is concrete and runnable.

**Type consistency check:** `output.Engagement` fields are `TStartSec / TEndSec / PeakUnits / CommandCount` — used identically in `engagement.go` (Task 5) and `replay-chart.js` (Task 9 reads `e.TStartSec` via JSON tag `t_start_sec`). `Anomaly` fields `TStartSec / TEndSec` consistent across analytics + handlers. `output.Event.Data` is `any` and handlers cope with both struct (fresh run) and `map[string]any` (cache load) — explicitly addressed in Task 8.7.

**Carry-forward risks (engineer should know):**
- `templ generate` must run before `go build` whenever `.templ` files change. CI/Makefile target could add this; intentionally not in MVP — engineer runs by hand.
- Plotly-JS vendored as 3.5MB binary. Acceptable for local dashboard. If pushing to remote it should be GZ-compressed (handled by `net/http` default? no — manual middleware). Out of scope v1.
- `analysis.json` events use `Data any` → on re-load from JSON it round-trips as `map[string]any`. Handlers must accept both shapes (already shown in Task 8.7 buildOrderRows). Future iteration: define `UnmarshalJSON` per Data variant.
