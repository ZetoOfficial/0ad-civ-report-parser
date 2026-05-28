# React + REST Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Снести templ + htmx UI; превратить Go-бэк в чистый REST JSON API; сделать React + Vite + TypeScript + Tailwind фронт в отдельной папке `web/`; в prod-сборке Go embed-ит `web/dist/` и отдаёт SPA на `/`.

**Architecture:** Парсер (`internal/replay/{commands,metadata,events,analytics,output,pipeline}`) не трогаем — он уже чистый бэк. `internal/replay/webui/` теперь только REST-эндпоинты, JSON-сериализация, SPA-fallback роутинг. Фронт в `web/` — отдельный npm-проект, в dev режиме vite на :5173 проксирует `/api/*` на Go :8080; в prod `vite build` → `web/dist/`, Go embed-ит через `//go:embed all:web/dist` и отдаёт всё статикой + index.html на любой `/replay/*` для SPA-роутинга.

**Tech Stack:** Go 1.25.5 (stdlib), React 18, TypeScript, Vite 5, Tailwind CSS 3, react-router-dom v6, react-plotly.js. Без state-менеджеров (Zustand/Redux), без UI-китов — кастомный Tailwind.

**Branch:** продолжаем на `replay-analyzer-mvp` (HEAD = `69dfed1`). templ/htmx уходит из истории как промежуточный шаг — squash при мерже опционально.

**Spec reference:** старая [replay-analyzer-mvp-design.md](../specs/2026-05-28-replay-analyzer-mvp-design.md) была про templ+htmx; этот план её заменяет в части UI-стека. Парсер и data model — без изменений.

---

## File Structure

| Path | Action | Responsibility |
|---|---|---|
| `internal/replay/webui/handlers.go` | rewrite | JSON-only REST handlers. Все templ-ссылки удалены. |
| `internal/replay/webui/server.go` | rewrite | Routes: `/api/replays`, `/api/replays/{id}`, embedded SPA fallback на `/` |
| `internal/replay/webui/templates/` | DELETE | Все `.templ` + `_templ.go` файлы — больше не нужны |
| `internal/replay/webui/static/htmx.min.js` | DELETE | |
| `internal/replay/webui/static/plotly.min.js` | DELETE | Будет re-vendored через npm в web/ |
| `internal/replay/webui/static/style.css` | DELETE | |
| `internal/replay/webui/static/replay-chart.js` | DELETE | |
| `internal/replay/webui/api_test.go` | create | Unit-тесты REST handlers (JSON shape, 404) |
| `go.mod`, `go.sum` | modify | Удалить `github.com/a-h/templ` через `go mod tidy` |
| `web/` | create | Новая папка для React app |
| `web/package.json` | create | npm deps + scripts |
| `web/vite.config.ts` | create | Vite config + proxy /api → :8080 в dev |
| `web/tsconfig.json` | create | TS config |
| `web/tsconfig.node.json` | create | TS config для vite.config.ts |
| `web/tailwind.config.js` | create | Tailwind config |
| `web/postcss.config.js` | create | PostCSS для Tailwind |
| `web/index.html` | create | Vite entry HTML |
| `web/src/main.tsx` | create | React entry + router setup |
| `web/src/App.tsx` | create | Layout shell + routes |
| `web/src/api.ts` | create | Типизированный fetch-клиент |
| `web/src/types.ts` | create | TS-типы, зеркало `output.Analysis` schema |
| `web/src/pages/IndexPage.tsx` | create | Список партий (карточки) |
| `web/src/pages/ReplayPage.tsx` | create | Страница партии (header + final state + chart + build order + anomalies) |
| `web/src/components/PlayerChip.tsx` | create | Цветной chip с цив-цветом |
| `web/src/components/DensityChart.tsx` | create | Plotly stacked-bar action density + markers |
| `web/src/components/BuildOrderList.tsx` | create | Build order таблица |
| `web/src/components/AnomalyList.tsx` | create | Anomalies список |
| `web/src/index.css` | create | Tailwind directives + минимум кастома |
| `web/dist/.gitkeep` | create | Чтобы Go `//go:embed` не падал до первой `npm run build` |
| `web/.gitignore` | create | `node_modules/`, `dist/*` кроме `.gitkeep` |
| `Makefile` | modify | Новые цели: `web-install`, `web-build`, `web-dev`; `replayreport` теперь зависит от `web-build` |
| `cmd/replayreport/main.go` | modify | Минимальные правки: вывод подсказки про `make web-dev` |
| `README.md` | modify | Новый dev/prod workflow |
| `CLAUDE.md` | modify | Обновить раздел про webui |

**File responsibility check:**
- Парсер каталоги (`commands/`, `metadata/`, `events/`, `analytics/`, `output/`, `pipeline.go`) **не трогаются** — они чистый бэк, JSON-сериализуемый
- `internal/replay/webui/` уменьшается с ~7 файлов до 3 (server.go, handlers.go, api_test.go)
- `web/` — изолированный фронт, может полностью заменяться без влияния на парсер
- Один файл = одна ответственность: api.ts только fetch, types.ts только типы, каждая страница — один файл, каждый компонент — один файл

---

## REST API Contract

Зеркало `output.Analysis` schema из Task 1 предыдущего плана.

### `GET /api/replays`

Список всех партий с минимумом меты для карточек.

```json
[
  {
    "match_id": "B3018EAD0161CF7E",
    "map": "Watering Holes (4 players)",
    "timestamp": 1779953790,
    "duration_ms": 622800,
    "players": [
      {"id": 1, "name": "zeto", "civ": "germ", "color": {"r": 10, "g": 10, "b": 190}, "is_ai": false},
      {"id": 2, "name": "Cleisthenes", "civ": "athen", "color": {"r": 125, "g": 235, "b": 15}, "is_ai": true, "ai_diff": 3}
    ],
    "outcome": "defeated"
  }
]
```

Отсортирован по `timestamp` desc (newest first). Outcome — для player 1 (предполагается «zeto»). Парсинг и кэш через `replay.Run` как раньше.

### `GET /api/replays/{matchID}`

Полный analysis.json. Тело = `output.Analysis` as-is, MarshalIndent НЕ нужен.

404 → `{"error": "not found"}`.

### SPA Fallback

`GET /` → embedded `web/dist/index.html`
`GET /replay/{id}` → embedded `web/dist/index.html` (client-side роутинг)
`GET /assets/*` → embedded `web/dist/assets/*`
`GET /api/*` → REST handlers выше; всё остальное на `/` уходит в SPA fallback

---

## Task 1: Strip templ/htmx, REST API in Go

**Files:**
- Delete: `internal/replay/webui/templates/` (вся папка)
- Delete: `internal/replay/webui/static/htmx.min.js`
- Delete: `internal/replay/webui/static/plotly.min.js`
- Delete: `internal/replay/webui/static/style.css`
- Delete: `internal/replay/webui/static/replay-chart.js`
- Rewrite: `internal/replay/webui/handlers.go`
- Rewrite: `internal/replay/webui/server.go`
- Create: `internal/replay/webui/api_test.go`
- Create: `web/dist/.gitkeep`
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1.1: Delete templ + htmx files**

```bash
rm -rf internal/replay/webui/templates/
rm -f internal/replay/webui/static/htmx.min.js
rm -f internal/replay/webui/static/plotly.min.js
rm -f internal/replay/webui/static/style.css
rm -f internal/replay/webui/static/replay-chart.js
mkdir -p web/dist
touch web/dist/.gitkeep
```

`internal/replay/webui/static/` остаётся пустой папкой — её тоже можно удалить:
```bash
rmdir internal/replay/webui/static/ 2>/dev/null || true
```

- [ ] **Step 1.2: Write new `internal/replay/webui/handlers.go` (JSON-only)**

```go
package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"
)

type handlers struct {
	repRoot string
}

type replayListItem struct {
	MatchID    string          `json:"match_id"`
	Map        string          `json:"map"`
	Timestamp  int64           `json:"timestamp"`
	DurationMs int64           `json:"duration_ms"`
	Players    []output.Player `json:"players"`
	Outcome    string          `json:"outcome"`
}

func (h *handlers) listReplays(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(h.repRoot)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("scan: %v", err))
		return
	}
	var items []replayListItem
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
		items = append(items, replayListItem{
			MatchID:    a.Game.MatchID,
			Map:        a.Game.Map,
			Timestamp:  a.Game.Timestamp,
			DurationMs: a.Game.DurationMs,
			Players:    a.Players,
			Outcome:    outcome,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Timestamp > items[j].Timestamp })
	writeJSON(w, http.StatusOK, items)
}

func (h *handlers) getReplay(w http.ResponseWriter, r *http.Request) {
	matchID := strings.TrimPrefix(r.URL.Path, "/api/replays/")
	if matchID == "" {
		writeJSONError(w, http.StatusBadRequest, "missing matchID")
		return
	}
	a, err := h.findByMatchID(matchID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *handlers) findByMatchID(matchID string) (*output.Analysis, error) {
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
			return &a, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
```

- [ ] **Step 1.3: Write new `internal/replay/webui/server.go` (REST + SPA fallback)**

```go
package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:../../../web/dist
var spaFS embed.FS

func NewServer(repRoot string) http.Handler {
	h := &handlers{repRoot: repRoot}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/replays", h.listReplays)
	mux.HandleFunc("/api/replays/", h.getReplay)

	dist, err := fs.Sub(spaFS, "../../../web/dist")
	if err != nil {
		panic("webui: embed dist subtree: " + err.Error())
	}
	staticFS := http.FileServer(http.FS(dist))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/" {
			serveIndex(w, r, dist)
			return
		}
		if _, err := fs.Stat(dist, strings.TrimPrefix(r.URL.Path, "/")); err == nil {
			staticFS.ServeHTTP(w, r)
			return
		}
		serveIndex(w, r, dist)
	})

	return mux
}

func serveIndex(w http.ResponseWriter, r *http.Request, dist fs.FS) {
	f, err := dist.Open("index.html")
	if err != nil {
		http.Error(w, "frontend not built — run `make web-build`", http.StatusServiceUnavailable)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	stat, _ := f.Stat()
	if rs, ok := f.(interface {
		ReadSeeker() http.File
	}); ok {
		http.ServeContent(w, r, "index.html", stat.ModTime(), rs.ReadSeeker())
		return
	}
	// Fallback: read whole file (small)
	buf := make([]byte, stat.Size())
	if _, err := f.Read(buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(buf)
}
```

Note: `//go:embed all:../../../web/dist` uses the `all:` prefix to include files starting with `.` or `_` (Vite generates `_assets` in some configs). The relative path `../../../web/dist` walks up from `internal/replay/webui/` to repo root. If your Go version disallows `..` in embed paths, fall back to `//go:embed all:web/dist` and move server.go to a package at repo root — but Go 1.25 supports the relative form.

**Update:** Go's `//go:embed` does NOT allow `..` in paths. Embedded files MUST be in the same package directory or a subdirectory. **Therefore:** copy the `dist/` contents into `internal/replay/webui/dist/` during build, OR move embed declaration to a package located such that `web/dist/` is below it.

**Chosen approach:** create a separate package `internal/replay/webui/spa/` that lives next to a symlinked or copied `dist/`. Actually the cleanest: post-build step in Makefile copies `web/dist/` to `internal/replay/webui/dist/`. Then:

```go
//go:embed all:dist
var spaFS embed.FS
```

inside `server.go` (which lives in `internal/replay/webui/`).

Rewrite `server.go` with this approach instead:

```go
package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var spaFS embed.FS

func NewServer(repRoot string) http.Handler {
	h := &handlers{repRoot: repRoot}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/replays", h.listReplays)
	mux.HandleFunc("/api/replays/", h.getReplay)

	dist, err := fs.Sub(spaFS, "dist")
	if err != nil {
		panic("webui: embed dist subtree: " + err.Error())
	}
	staticFS := http.FileServer(http.FS(dist))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/" {
			serveIndex(w, dist)
			return
		}
		clean := strings.TrimPrefix(r.URL.Path, "/")
		if _, err := fs.Stat(dist, clean); err == nil {
			staticFS.ServeHTTP(w, r)
			return
		}
		serveIndex(w, dist)
	})

	return mux
}

func serveIndex(w http.ResponseWriter, dist fs.FS) {
	raw, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		http.Error(w, "frontend not built — run `make web-build`", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(raw)
}
```

Create stub now so Go compiles immediately:

```bash
mkdir -p internal/replay/webui/dist
echo '<!doctype html><html><body>frontend not built — run <code>make web-build</code></body></html>' > internal/replay/webui/dist/index.html
```

- [ ] **Step 1.4: Write `internal/replay/webui/api_test.go`**

```go
package webui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"
)

func TestListReplaysEmptyRoot(t *testing.T) {
	dir := t.TempDir()
	srv := NewServer(dir)
	req := httptest.NewRequest(http.MethodGet, "/api/replays", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q", ct)
	}
	body := strings.TrimSpace(w.Body.String())
	if body != "null" && body != "[]" {
		t.Errorf("body = %q, want null or []", body)
	}
}

func TestGetReplayNotFound(t *testing.T) {
	dir := t.TempDir()
	srv := NewServer(dir)
	req := httptest.NewRequest(http.MethodGet, "/api/replays/DOESNOTEXIST", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("body parse: %v", err)
	}
	if body["error"] == "" {
		t.Errorf("body = %v, want error field", body)
	}
}

func TestGetReplayFound(t *testing.T) {
	dir := t.TempDir()
	// Use the existing fixture
	src := filepath.Join("..", "..", "..", "testdata", "replays", "short-germ-vs-3p")
	if _, err := os.Stat(src); err != nil {
		t.Skipf("fixture missing: %v", err)
	}
	rdir := filepath.Join(dir, "short-germ-vs-3p")
	if err := os.Mkdir(rdir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"commands.txt", "metadata.json"} {
		raw, err := os.ReadFile(filepath.Join(src, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(rdir, name), raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	a, err := replay.Run(rdir)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	srv := NewServer(dir)
	req := httptest.NewRequest(http.MethodGet, "/api/replays/"+a.Game.MatchID, nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var got output.Analysis
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Game.MatchID != a.Game.MatchID {
		t.Errorf("MatchID = %q, want %q", got.Game.MatchID, a.Game.MatchID)
	}
	if len(got.Events) == 0 {
		t.Error("no events in response")
	}
}

func TestSpaFallback(t *testing.T) {
	dir := t.TempDir()
	srv := NewServer(dir)
	req := httptest.NewRequest(http.MethodGet, "/replay/DOESNOTEXIST", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d (SPA fallback should 200)", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q", ct)
	}
}
```

- [ ] **Step 1.5: Add `.gitignore` for `web/dist/`-related files**

The Go embed needs `internal/replay/webui/dist/` to exist. But built assets shouldn't be in git history (they're machine-generated). Plan:
- `internal/replay/webui/dist/` IS gitignored (except `.gitkeep` and the stub `index.html`)
- Makefile build copies `web/dist/*` into `internal/replay/webui/dist/`

Create `internal/replay/webui/dist/.gitkeep`:

```bash
mkdir -p internal/replay/webui/dist
touch internal/replay/webui/dist/.gitkeep
```

Add lines to repo root `.gitignore` (create if missing):

```
internal/replay/webui/dist/*
!internal/replay/webui/dist/.gitkeep
!internal/replay/webui/dist/index.html
web/node_modules/
web/dist/
```

Then create a STUB `internal/replay/webui/dist/index.html` (1 line) that's committed:

```bash
echo '<!doctype html><meta charset="utf-8"><body>frontend not built — run <code>make web-build</code></body>' > internal/replay/webui/dist/index.html
```

- [ ] **Step 1.6: Update CLI main.go banner**

Edit `cmd/replayreport/main.go`. Find the printf that says `"serving http://localhost%s\n"` and replace with:

```go
	fmt.Printf("serving http://localhost%s\n", addr)
	fmt.Println("  (dev: run `make web-dev` for hot-reload at :5173 → :8080)")
```

- [ ] **Step 1.7: Remove templ dep from go.mod**

```bash
go mod tidy
```

Expected: `github.com/a-h/templ` is removed from `go.mod` and `go.sum`.

- [ ] **Step 1.8: Run tests**

```bash
go test ./...
```

Expected: all tests pass; analytics/output/pipeline/etc. unchanged; `webui` tests pass (4 of them).

- [ ] **Step 1.9: Build + smoke**

```bash
make replayreport
./bin/replayreport &
SERVER_PID=$!
sleep 1
curl -s -i http://localhost:8080/api/replays | head -10
curl -s http://localhost:8080/api/replays | head -c 200; echo
curl -s -i http://localhost:8080/replay/anything | head -10  # should 200 + HTML (stub)
kill $SERVER_PID
```

Expected: `/api/replays` returns `application/json` and JSON array; `/replay/anything` returns 200 + HTML (the stub index.html).

- [ ] **Step 1.10: Commit**

```bash
git add internal/replay/webui/ web/dist/.gitkeep .gitignore go.mod go.sum cmd/replayreport/main.go
git commit -m "replayreport: rewrite webui as REST API + SPA fallback (drop templ/htmx)"
```

---

## Task 2: Scaffold React + Vite + TS + Tailwind

**Files:**
- Create: `web/package.json`, `web/package-lock.json`
- Create: `web/vite.config.ts`
- Create: `web/tsconfig.json`, `web/tsconfig.node.json`
- Create: `web/tailwind.config.js`, `web/postcss.config.js`
- Create: `web/index.html`
- Create: `web/src/main.tsx`
- Create: `web/src/App.tsx`
- Create: `web/src/index.css`
- Create: `web/src/api.ts`
- Create: `web/src/types.ts`
- Create: `web/.gitignore`

Hello-world page that fetches `/api/replays` and prints count. Cards/routing/styling come in Task 3+.

- [ ] **Step 2.1: Scaffold via vite create**

```bash
cd web
npm create vite@latest . -- --template react-ts
```

Vite asks confirmation since `.gitkeep` exists in `dist/` — answer yes. It scaffolds `package.json`, `tsconfig.json`, `vite.config.ts`, `index.html`, `src/main.tsx`, `src/App.tsx`, `src/index.css`, etc.

- [ ] **Step 2.2: Install additional deps**

```bash
cd web
npm install react-router-dom plotly.js react-plotly.js
npm install -D @types/react-plotly.js tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

Expected: `package.json` now has `react`, `react-dom`, `react-router-dom`, `plotly.js`, `react-plotly.js` in deps; `tailwindcss`, `postcss`, `autoprefixer`, `@types/react-plotly.js` in devDeps. Also creates `tailwind.config.js` and `postcss.config.js`.

- [ ] **Step 2.3: Configure Tailwind content paths**

Edit `web/tailwind.config.js` to:

```js
/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: { extend: {} },
  plugins: [],
};
```

Edit `web/src/index.css` — replace its contents with:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

html, body, #root { height: 100%; }
body { background: #fafafa; color: #111; font-family: -apple-system, system-ui, sans-serif; }
```

- [ ] **Step 2.4: Configure Vite proxy**

Replace `web/vite.config.ts` with:

```ts
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: false,
      },
    },
  },
  build: {
    outDir: "dist",
    assetsDir: "assets",
    emptyOutDir: true,
  },
});
```

- [ ] **Step 2.5: Write `web/src/types.ts` (TS mirror of `output.Analysis`)**

```ts
export interface Color {
  r: number;
  g: number;
  b: number;
}

export interface Player {
  id: number;
  name: string;
  civ: string;
  team: number;
  is_ai: boolean;
  ai_diff?: number;
  color: Color;
}

export interface Resources {
  food: number;
  wood: number;
  stone: number;
  metal: number;
}

export interface PlayerFinalState {
  phase: string;
  state: string;
  outcome: string;
  pop_count: number;
  pop_limit: number;
  pop_max: number;
  resource_counts: Resources;
  researched_techs: string[];
}

export interface FinalState {
  players: Record<string, PlayerFinalState>; // JSON keys are stringified ints
}

export interface Engagement {
  t_start_sec: number;
  t_end_sec: number;
  target: number;
  peak_units: number;
  command_count: number;
}

export interface Anomaly {
  type: string;
  t_start_sec: number;
  t_end_sec: number;
  severity: string;
  details?: Record<string, unknown>;
}

export interface DensityBin {
  t_sec: number;
  counts: Record<string, number>;
}

export interface PlayerMetrics {
  phase_timings: Record<string, number>;
  engagements: Engagement[];
  anomalies: Anomaly[];
}

export interface Metrics {
  players: Record<string, PlayerMetrics>;
  action_density: DensityBin[];
}

export interface GameInfo {
  match_id: string;
  map: string;
  map_type: string;
  timestamp: number;
  duration_ms: number;
  engine_version: string;
  victory_conditions: string[];
}

export interface ReplayEvent {
  t: number;
  player: number;
  type: string;
  data?: unknown;
}

export interface Analysis {
  schema_version: number;
  game: GameInfo;
  players: Player[];
  events: ReplayEvent[];
  snapshots: unknown[];
  final_state: FinalState;
  metrics: Metrics;
}

export interface ReplayListItem {
  match_id: string;
  map: string;
  timestamp: number;
  duration_ms: number;
  players: Player[];
  outcome: string;
}
```

- [ ] **Step 2.6: Write `web/src/api.ts`**

```ts
import type { Analysis, ReplayListItem } from "./types";

const base = ""; // same-origin; vite proxies /api in dev

export async function listReplays(): Promise<ReplayListItem[]> {
  const r = await fetch(`${base}/api/replays`);
  if (!r.ok) throw new Error(`/api/replays: ${r.status}`);
  return (await r.json()) ?? [];
}

export async function getReplay(matchID: string): Promise<Analysis> {
  const r = await fetch(`${base}/api/replays/${encodeURIComponent(matchID)}`);
  if (!r.ok) throw new Error(`/api/replays/${matchID}: ${r.status}`);
  return r.json();
}
```

- [ ] **Step 2.7: Replace `web/src/App.tsx` with hello-world that fetches /api/replays**

```tsx
import { useEffect, useState } from "react";
import { listReplays } from "./api";
import type { ReplayListItem } from "./types";

export default function App() {
  const [items, setItems] = useState<ReplayListItem[] | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    listReplays().then(setItems).catch((e) => setErr(String(e)));
  }, []);

  return (
    <main className="max-w-4xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">replayreport</h1>
      {err && <p className="text-red-600">Error: {err}</p>}
      {!err && items === null && <p>Loading…</p>}
      {items && <p>Found {items.length} replays.</p>}
    </main>
  );
}
```

- [ ] **Step 2.8: Update `web/src/main.tsx` if scaffold gave a different shape**

Verify content of `web/src/main.tsx` is approximately:

```tsx
import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.tsx";
import "./index.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
```

If vite scaffold left it different, replace with this.

- [ ] **Step 2.9: Update `web/.gitignore`**

```
node_modules/
dist/
*.log
.DS_Store
```

- [ ] **Step 2.10: Smoke-test dev mode**

In one terminal: build & run Go binary.

```bash
make replayreport
./bin/replayreport &
echo $! > /tmp/replayreport.pid
sleep 1
```

In another step (or background-spawn vite):

```bash
cd web
npm run dev &
VITE_PID=$!
sleep 3
curl -s http://localhost:5173/ | grep -c '<div id="root">'  # → 1
curl -s http://localhost:5173/api/replays | head -c 200; echo
kill $VITE_PID
kill $(cat /tmp/replayreport.pid)
```

Expected: vite serves HTML on :5173; proxy forwards `/api/replays` to Go and returns JSON.

- [ ] **Step 2.11: Smoke-test prod build path**

```bash
cd web
npm run build
ls -la dist/
ls dist/assets/ | head
```

Expected: `dist/index.html` exists; `dist/assets/*.js`, `*.css` populated.

- [ ] **Step 2.12: Commit**

```bash
git add web/
git commit -m "web: scaffold react + vite + ts + tailwind + plotly"
```

Note: `web/dist/` is gitignored — only `web/.gitignore`, source, and configs are committed. `package-lock.json` IS committed.

---

## Task 3: List page (cards)

**Files:**
- Create: `web/src/pages/IndexPage.tsx`
- Create: `web/src/components/PlayerChip.tsx`
- Modify: `web/src/App.tsx` (add router, mount IndexPage at `/`)

- [ ] **Step 3.1: Write `web/src/components/PlayerChip.tsx`**

```tsx
import type { Player } from "../types";

interface Props {
  player: Player;
  compact?: boolean;
}

export function PlayerChip({ player, compact }: Props) {
  const bg = `rgb(${player.color.r}, ${player.color.g}, ${player.color.b})`;
  return (
    <span
      className="inline-block px-2 py-0.5 rounded-full text-white text-xs font-medium"
      style={{ backgroundColor: bg, textShadow: "0 1px 1px rgba(0,0,0,.3)" }}
    >
      {player.name} ({player.civ})
      {player.is_ai && ` · AI d${player.ai_diff ?? "?"}`}
      {!compact && player.team >= 0 && ` · team ${player.team}`}
    </span>
  );
}
```

- [ ] **Step 3.2: Write `web/src/pages/IndexPage.tsx`**

```tsx
import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { listReplays } from "../api";
import type { ReplayListItem } from "../types";
import { PlayerChip } from "../components/PlayerChip";

function formatDate(unix: number): string {
  if (!unix) return "—";
  return new Date(unix * 1000).toLocaleString("ru-RU", {
    day: "2-digit", month: "short", hour: "2-digit", minute: "2-digit",
  });
}

function formatDuration(ms: number): string {
  const min = Math.floor(ms / 60000);
  const sec = Math.floor((ms % 60000) / 1000);
  return `${min}:${sec.toString().padStart(2, "0")}`;
}

export function IndexPage() {
  const [items, setItems] = useState<ReplayListItem[] | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    listReplays().then(setItems).catch((e) => setErr(String(e)));
  }, []);

  if (err) return <p className="text-red-600">Error: {err}</p>;
  if (items === null) return <p>Loading…</p>;
  if (items.length === 0) return <p>No replays yet.</p>;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Партии ({items.length})</h1>
      <ul className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
        {items.map((it) => (
          <li key={it.match_id} className="bg-white rounded-lg border border-gray-200 hover:border-gray-400 transition">
            <Link to={`/replay/${it.match_id}`} className="block p-4">
              <div className="font-semibold">{it.map || "—"}</div>
              <div className="text-xs text-gray-500 mt-1">
                {formatDate(it.timestamp)} · {formatDuration(it.duration_ms)}
              </div>
              <div className="flex flex-wrap gap-1 my-2">
                {it.players.map((p) => <PlayerChip key={p.id} player={p} compact />)}
              </div>
              <div className="text-xs text-gray-600">{it.outcome}</div>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
```

- [ ] **Step 3.3: Rewrite `web/src/App.tsx` to use router**

```tsx
import { BrowserRouter, Routes, Route, Link } from "react-router-dom";
import { IndexPage } from "./pages/IndexPage";

export default function App() {
  return (
    <BrowserRouter>
      <header className="bg-gray-900 text-white px-6 py-3">
        <Link to="/" className="font-semibold">replayreport</Link>
      </header>
      <main className="max-w-6xl mx-auto p-6">
        <Routes>
          <Route path="/" element={<IndexPage />} />
          <Route path="/replay/:matchID" element={<ReplayPlaceholder />} />
        </Routes>
      </main>
    </BrowserRouter>
  );
}

function ReplayPlaceholder() {
  return <p>Replay detail page — Task 4.</p>;
}
```

- [ ] **Step 3.4: Smoke-test**

```bash
# Go binary already running from Task 2 smoke, or:
make replayreport
./bin/replayreport &
REPLAY_PID=$!
cd web && npm run dev &
VITE_PID=$!
sleep 3
curl -s http://localhost:5173/ -o /tmp/index.html
grep -c 'Партии' /tmp/index.html  # this is server-rendered as empty — React renders client-side
# Better: just check the bundle includes IndexPage
grep -c 'IndexPage' web/dist/assets/*.js 2>/dev/null || echo "rebuild needed"
kill $VITE_PID $REPLAY_PID
```

Manually open `http://localhost:5173/` in the browser if testing locally. Cards should appear with civ-color chips.

- [ ] **Step 3.5: Commit**

```bash
git add web/src/
git commit -m "web: index page with replay cards + router skeleton"
```

---

## Task 4: Replay detail page (header + final state + build order + anomalies)

**Files:**
- Create: `web/src/pages/ReplayPage.tsx`
- Create: `web/src/components/BuildOrderList.tsx`
- Create: `web/src/components/AnomalyList.tsx`
- Modify: `web/src/App.tsx` (mount ReplayPage on `/replay/:matchID`)

- [ ] **Step 4.1: Write `web/src/components/BuildOrderList.tsx`**

```tsx
import type { Analysis, ReplayEvent } from "../types";

const KEEP_EVENT_TYPES = new Set(["research", "construct", "resign"]);

const KEEP_TEMPLATE_PREFIXES = [
  "phase_",
  "structures/", // any construct is kept
];

function eventLabel(e: ReplayEvent): string | null {
  if (!KEEP_EVENT_TYPES.has(e.type)) return null;
  if (e.type === "resign") return "RESIGN";
  const data = e.data as { template?: string } | undefined;
  const tmpl = data?.template ?? "";
  if (e.type === "research") {
    // keep only phase_* + hero/champion/wonder gates if recognizable by name
    if (tmpl.startsWith("phase_")) return `research ${tmpl}`;
    return null;
  }
  if (e.type === "construct") {
    // keep significant structures only
    const keep = [
      "wonder", "civic_centre", "fortress", "barracks", "stable",
      "elephant_stable", "kennel", "embassy", "temple", "tower", "wallset",
    ];
    if (keep.some((k) => tmpl.includes(k))) return `construct ${tmpl}`;
    return null;
  }
  return null;
}

function fmtTime(ms: number): string {
  const m = Math.floor(ms / 60000);
  const s = Math.floor((ms % 60000) / 1000);
  return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
}

interface Props { analysis: Analysis }

export function BuildOrderList({ analysis }: Props) {
  const playerName: Record<number, string> = {};
  for (const p of analysis.players) playerName[p.id] = p.name;

  const rows: { time: string; player: string; event: string }[] = [];
  for (const e of analysis.events) {
    const label = eventLabel(e);
    if (!label) continue;
    rows.push({ time: fmtTime(e.t), player: playerName[e.player] ?? `p${e.player}`, event: label });
  }

  return (
    <table className="w-full text-sm">
      <thead>
        <tr className="text-left text-gray-500">
          <th className="font-normal py-1">t</th>
          <th className="font-normal py-1">player</th>
          <th className="font-normal py-1">event</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((r, i) => (
          <tr key={i} className="border-t border-gray-100">
            <td className="py-1 pr-2 font-mono">{r.time}</td>
            <td className="py-1 pr-2">{r.player}</td>
            <td className="py-1">{r.event}</td>
          </tr>
        ))}
        {rows.length === 0 && (
          <tr><td colSpan={3} className="text-gray-500 py-2">no significant events</td></tr>
        )}
      </tbody>
    </table>
  );
}
```

- [ ] **Step 4.2: Write `web/src/components/AnomalyList.tsx`**

```tsx
import type { Analysis } from "../types";

interface Props { analysis: Analysis }

export function AnomalyList({ analysis }: Props) {
  const rows: { type: string; severity: string; t: string; detail: string }[] = [];
  for (const m of Object.values(analysis.metrics.players)) {
    for (const an of m.anomalies) {
      let detail = "";
      const d = an.details as Record<string, unknown> | undefined;
      if (d?.target !== undefined) detail = `target=${d.target}`;
      rows.push({
        type: an.type,
        severity: an.severity,
        t: `${an.t_start_sec}..${an.t_end_sec}s`,
        detail,
      });
    }
  }

  if (rows.length === 0) return <p className="text-gray-500 italic">Чисто</p>;

  return (
    <ul className="space-y-1">
      {rows.map((r, i) => (
        <li key={i}>
          <span className="font-semibold">{r.type}</span>
          <span className="text-gray-500"> · {r.severity} · {r.t} · {r.detail}</span>
        </li>
      ))}
    </ul>
  );
}
```

- [ ] **Step 4.3: Write `web/src/pages/ReplayPage.tsx`**

```tsx
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { getReplay } from "../api";
import type { Analysis } from "../types";
import { PlayerChip } from "../components/PlayerChip";
import { BuildOrderList } from "../components/BuildOrderList";
import { AnomalyList } from "../components/AnomalyList";

export function ReplayPage() {
  const { matchID = "" } = useParams();
  const [a, setA] = useState<Analysis | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    getReplay(matchID).then(setA).catch((e) => setErr(String(e)));
  }, [matchID]);

  if (err) return <p className="text-red-600">Error: {err}</p>;
  if (!a) return <p>Loading…</p>;

  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-2xl font-bold">{a.game.map}</h1>
        <div className="text-sm text-gray-500 flex gap-4 mt-1">
          <span>matchID: {a.game.match_id}</span>
          <span>длительность: {Math.floor(a.game.duration_ms / 60000)} мин</span>
          <span>движок: {a.game.engine_version}</span>
        </div>
      </header>

      <section className="bg-white p-4 rounded-lg border border-gray-200">
        <h2 className="font-semibold mb-2">Игроки</h2>
        <div className="space-y-2">
          {a.players.map((p) => {
            const fs = a.final_state.players[String(p.id)];
            return (
              <div key={p.id} className="flex items-center gap-3 text-sm">
                <PlayerChip player={p} />
                {fs && (
                  <>
                    <span className="font-semibold">{fs.outcome}</span>
                    <span className="text-gray-500">фаза: {fs.phase || "—"}</span>
                    <span className="text-gray-500">поп: {fs.pop_count}/{fs.pop_limit}</span>
                  </>
                )}
              </div>
            );
          })}
        </div>
      </section>

      <section className="bg-white p-4 rounded-lg border border-gray-200">
        <h2 className="font-semibold mb-2">Build order (значимые события)</h2>
        <BuildOrderList analysis={a} />
      </section>

      <section className="bg-white p-4 rounded-lg border border-gray-200">
        <h2 className="font-semibold mb-2">Аномалии</h2>
        <AnomalyList analysis={a} />
      </section>
    </div>
  );
}
```

- [ ] **Step 4.4: Update `web/src/App.tsx` to mount ReplayPage**

Replace the `ReplayPlaceholder` function with the import, then change the route. Final App.tsx:

```tsx
import { BrowserRouter, Routes, Route, Link } from "react-router-dom";
import { IndexPage } from "./pages/IndexPage";
import { ReplayPage } from "./pages/ReplayPage";

export default function App() {
  return (
    <BrowserRouter>
      <header className="bg-gray-900 text-white px-6 py-3">
        <Link to="/" className="font-semibold">replayreport</Link>
      </header>
      <main className="max-w-6xl mx-auto p-6">
        <Routes>
          <Route path="/" element={<IndexPage />} />
          <Route path="/replay/:matchID" element={<ReplayPage />} />
        </Routes>
      </main>
    </BrowserRouter>
  );
}
```

- [ ] **Step 4.5: Smoke-test**

```bash
make replayreport
./bin/replayreport testdata/replays/short-germ-vs-3p
./bin/replayreport &
REPLAY_PID=$!
cd web && npm run dev &
VITE_PID=$!
sleep 3
# Get a matchID
MATCH=$(curl -s http://localhost:5173/api/replays | python3 -c "import sys,json; d=json.load(sys.stdin); print(d[0]['match_id'])")
echo "MATCH=$MATCH"
# Verify single-replay endpoint returns full analysis
curl -s "http://localhost:5173/api/replays/$MATCH" | python3 -c "import sys,json; d=json.load(sys.stdin); print('schema:', d['schema_version'], 'events:', len(d['events']))"
kill $VITE_PID $REPLAY_PID
```

Expected: schema 1, non-zero events.

Open `http://localhost:5173/replay/{MATCH}` in browser to visually verify.

- [ ] **Step 4.6: Commit**

```bash
git add web/src/
git commit -m "web: replay detail page (header + players + build order + anomalies)"
```

---

## Task 5: Action density chart in React (react-plotly.js)

**Files:**
- Create: `web/src/components/DensityChart.tsx`
- Modify: `web/src/pages/ReplayPage.tsx` (mount the chart)

- [ ] **Step 5.1: Write `web/src/components/DensityChart.tsx`**

```tsx
import Plot from "react-plotly.js";
import type { Analysis } from "../types";

interface Props { analysis: Analysis }

const CATEGORIES = ["military", "build", "research", "economy", "other"] as const;

export function DensityChart({ analysis }: Props) {
  const bins = analysis.metrics.action_density;
  if (!bins || bins.length === 0) {
    return <p className="text-gray-500 italic">no command bins</p>;
  }

  const x = bins.map((b) => b.t_sec);
  const traces = CATEGORIES.map((cat) => ({
    name: cat,
    type: "bar" as const,
    x,
    y: bins.map((b) => b.counts[cat] ?? 0),
  }));

  const phaseShapes: Plotly.Shape[] = [];
  const phaseAnnotations: Partial<Plotly.Annotations>[] = [];
  for (const m of Object.values(analysis.metrics.players)) {
    for (const [name, t] of Object.entries(m.phase_timings)) {
      phaseShapes.push({
        type: "line", xref: "x", yref: "paper",
        x0: t, x1: t, y0: 0, y1: 1,
        line: { dash: "dash", width: 1, color: "#555" },
      });
      phaseAnnotations.push({
        x: t, y: 1, yref: "paper",
        text: name, showarrow: false,
        font: { size: 10, color: "#555" },
      });
    }
  }

  const engShapes: Plotly.Shape[] = [];
  for (const m of Object.values(analysis.metrics.players)) {
    for (const e of m.engagements) {
      if (e.peak_units < 5) continue;
      engShapes.push({
        type: "line", xref: "x", yref: "paper",
        x0: e.t_start_sec, x1: e.t_start_sec, y0: 0, y1: 1,
        line: { dash: "solid", width: Math.min(1 + Math.log2(e.peak_units), 4), color: "rgba(220,0,0,0.5)" },
      });
    }
  }

  return (
    <Plot
      data={traces as unknown as Plotly.Data[]}
      layout={{
        barmode: "stack",
        margin: { t: 24, r: 16, l: 40, b: 32 },
        xaxis: { title: { text: "сек" } },
        yaxis: { title: { text: "команд / 30 сек" } },
        shapes: [...phaseShapes, ...engShapes],
        annotations: phaseAnnotations,
        legend: { orientation: "h", y: -0.2 },
        autosize: true,
      }}
      style={{ width: "100%", height: 360 }}
      useResizeHandler
      config={{ displayModeBar: false }}
    />
  );
}
```

Note on TS types: `react-plotly.js` types are sometimes imprecise. The `as unknown as Plotly.Data[]` cast in the data prop is a known workaround; if your tsconfig is strict the alternative is `data={traces as any}` — equally ugly. Pick one and move on.

- [ ] **Step 5.2: Add the chart to ReplayPage**

Edit `web/src/pages/ReplayPage.tsx`. Add the import:

```tsx
import { DensityChart } from "../components/DensityChart";
```

Insert a new section AFTER the "Игроки" section and BEFORE "Build order":

```tsx
      <section className="bg-white p-4 rounded-lg border border-gray-200">
        <h2 className="font-semibold mb-2">Плотность действий (30 сек)</h2>
        <DensityChart analysis={a} />
      </section>
```

- [ ] **Step 5.3: Smoke-test**

```bash
cd web
npm run build 2>&1 | tail -5
```

Expected: `vite build` succeeds without TS errors.

If it errors on Plotly types, change the cast as noted in 5.1.

- [ ] **Step 5.4: Manual browser check**

```bash
./bin/replayreport &
cd web && npm run dev &
sleep 3
# Open http://localhost:5173/, click a replay
# Verify: chart appears with stacked bars + dashed lines (phases) + red solid lines (engagements)
kill %1 %2 || true
```

- [ ] **Step 5.5: Commit**

```bash
git add web/
git commit -m "web: action density chart with phase + engagement markers"
```

---

## Task 6: Production embed pipeline (Makefile + build flow)

**Files:**
- Modify: `Makefile` (add `web-install`, `web-build`, `web-dev`; update `replayreport` deps)
- Verify: `internal/replay/webui/server.go` embed works after build
- Modify: `internal/replay/webui/dist/index.html` (stays as gitignored after first build, only the stub committed)

- [ ] **Step 6.1: Update Makefile**

Append to `Makefile`:

```make
.PHONY: web-install web-build web-dev
web-install:
	cd web && npm install

web-build: web-install
	cd web && npm run build
	rm -rf internal/replay/webui/dist
	mkdir -p internal/replay/webui/dist
	cp -R web/dist/. internal/replay/webui/dist/

web-dev:
	cd web && npm run dev
```

Update the existing `replayreport` target to depend on `web-build`:

Find:
```make
replayreport:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/replayreport ./cmd/replayreport
```

Replace with:
```make
replayreport: web-build
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/replayreport ./cmd/replayreport
```

Also add a `replayreport-fast` target that skips web build (for backend-only iterations):

```make
.PHONY: replayreport-fast
replayreport-fast:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/replayreport ./cmd/replayreport
```

- [ ] **Step 6.2: Smoke full pipeline**

```bash
# Clean state
rm -rf internal/replay/webui/dist
mkdir -p internal/replay/webui/dist
echo '<!doctype html><meta charset="utf-8"><body>stub' > internal/replay/webui/dist/index.html

# Build
make replayreport

# Verify dist has real React files now
ls internal/replay/webui/dist/
ls internal/replay/webui/dist/assets/

# Run
./bin/replayreport &
sleep 1
curl -s http://localhost:8080/ | head -c 500; echo
curl -s http://localhost:8080/replay/anything | grep -c 'replayreport' || true
curl -s http://localhost:8080/assets/ | head -1 || true
kill %1
```

Expected: `/` returns the real React index.html (includes `<script type="module"`...). `/replay/anything` returns the same index.html (SPA fallback). `/assets/*.js` returns the JS bundle.

- [ ] **Step 6.3: Verify `make replayreport-fast` still works (no web rebuild)**

```bash
touch internal/replay/webui/dist/.gitkeep  # ensure embed path exists
make replayreport-fast
./bin/replayreport &
sleep 1
curl -s http://localhost:8080/api/replays | head -c 100; echo
kill %1
```

Expected: API works. Frontend may be stale (from previous `make replayreport`) but that's intentional for fast iterations.

- [ ] **Step 6.4: Run all tests**

```bash
go test ./...
```

- [ ] **Step 6.5: Commit**

```bash
git add Makefile
git commit -m "build: web-build + web-dev targets, replayreport depends on web bundle"
```

---

## Task 7: Docs + final corpus smoke

**Files:**
- Modify: `README.md`
- Modify: `CLAUDE.md`

- [ ] **Step 7.1: Update README replayreport section**

Edit `README.md`. Replace the existing `## replayreport` section (added in the previous plan's Task 10) with:

```markdown
## replayreport

Бэк на Go + фронт на React (отдельная папка `web/`). Бинарь `replayreport`
парсит реплеи, держит REST API на :8080, и отдаёт собранный React-бандл из
embedded `dist/` на остальные роуты.

### Запуск

```bash
make replayreport                # build web/ + go → bin/replayreport
./bin/replayreport               # → http://localhost:8080
./bin/replayreport <replay-dir>  # парсит один replay (без HTTP)
./bin/replayreport --check --all # CI: exit 2 если что-то упало
```

### Разработка (hot-reload)

```bash
# Терминал 1: бэк
./bin/replayreport                # порт 8080

# Терминал 2: фронт с HMR
make web-dev                      # vite на 5173, проксирует /api/* на :8080
```

Открой `http://localhost:5173/` — изменения в `web/src/` подхватываются мгновенно.
Изменения в Go перебилди вручную (`make replayreport-fast` если фронт не менялся).

### Структура

- **Backend**: `cmd/replayreport/`, `internal/replay/` — парсер + REST API + SPA-fallback
- **Frontend**: `web/` — React 18 + TypeScript + Vite + Tailwind + react-plotly.js + react-router
- **API contract**: `GET /api/replays` (list), `GET /api/replays/{matchID}` (full Analysis)
- **Embed**: prod-сборка кладёт `web/dist/` в `internal/replay/webui/dist/` и embed-ит через `//go:embed`

### Что не в v1

Time-series графики population/resources/active-gatherers — `metadata.json`
реальных пользовательских реплеев не содержит `sequences`. Текущий график —
плотность команд (30-сек бины × 5 категорий) + маркеры фаз и крупных боёв.
Полноценные кривые требуют reducer-симуляции (фаза 3).

Также вне scope v1: APM/EAPM, кросс-партийная аналитика, real-time mode,
markdown-отчёт по партии, idle worker detection.
```

- [ ] **Step 7.2: Update CLAUDE.md Replay analyzer section**

Find the existing `## Replay analyzer` section and replace its **Пакеты** subsection with:

```markdown
**Пакеты:**
- `internal/replay/commands/` — streaming line-reader `commands.txt`
- `internal/replay/metadata/` — loader `metadata.json` (PlayerState scalars)
- `internal/replay/events/` — типизированный декодер cmd-строк
- `internal/replay/analytics/` — phase timings, engagements, panic_garrison, action density
- `internal/replay/output/` — schema + atomic JSON writer
- `internal/replay/pipeline.go` — `Run(replayDir)` оркестратор + mtime-кеш
- `internal/replay/webui/` — REST API (`/api/replays`, `/api/replays/{id}`) + SPA fallback (embedded React build)
- `cmd/replayreport/main.go` — CLI
- `web/` — React 18 + TypeScript + Vite + Tailwind + react-plotly.js + react-router. Build → `web/dist/` → copied to `internal/replay/webui/dist/` for embed.
```

Add right after the "Пакеты" subsection:

```markdown
**UI стек:** React + Vite + Tailwind + Plotly через react-plotly.js. Dev: vite на :5173 проксирует `/api` на Go :8080. Prod: Go embed-ит `web/dist/` и отдаёт SPA на любой не-`/api/*` роут. `make replayreport` собирает web → копирует в `internal/replay/webui/dist/` → собирает Go-бинарь. `make replayreport-fast` пропускает web-сборку.

**REST contract:**
- `GET /api/replays` → `[]{match_id, map, timestamp, duration_ms, players, outcome}`
- `GET /api/replays/{matchID}` → full `output.Analysis` (schema v1)
- 404 → `{"error": "not found"}`
```

- [ ] **Step 7.3: Final corpus smoke**

```bash
make replayreport
./bin/replayreport --check --all
```

Expected: `scan: N ok, M skipped (no metadata), 0 failed`.

- [ ] **Step 7.4: Run all tests**

```bash
go test ./...
```

- [ ] **Step 7.5: Manual browser smoke**

```bash
./bin/replayreport &
sleep 1
open http://localhost:8080/ || echo "open http://localhost:8080/ manually"
# Click a replay card; verify:
#   - page loads
#   - players show with civ-color chips
#   - action density chart renders
#   - phase markers (dashed) visible
#   - build order (filtered) populates
#   - anomalies show or "Чисто"
# Close server with kill %1
```

- [ ] **Step 7.6: Commit**

```bash
git add README.md CLAUDE.md
git commit -m "docs: react + rest migration — usage, dev workflow, stack"
```

---

## Self-Review

**Spec coverage:**

| Requirement | Task(s) |
|---|---|
| Remove templ + htmx | Task 1 |
| REST API: `GET /api/replays`, `GET /api/replays/{id}` | Task 1 |
| Go embed React build via `//go:embed all:dist` | Task 1 (stub) + Task 6 (real bundle copied) |
| SPA fallback (/, /replay/* → index.html; /assets/* → static) | Task 1 |
| API unit tests | Task 1 |
| React 18 + TypeScript + Vite | Task 2 |
| Tailwind CSS | Task 2 |
| react-router-dom v6 | Tasks 2–3 |
| react-plotly.js | Tasks 2, 5 |
| Vite proxy `/api/*` → :8080 | Task 2 |
| API client (typed) | Task 2 |
| TS mirror of `output.Analysis` schema | Task 2 |
| Index page with cards | Task 3 |
| Replay detail page | Task 4 |
| Filtered build order (no farmstead spam) | Task 4 (BuildOrderList — `eventLabel` filter) |
| Action density chart with phase + engagement markers | Task 5 |
| Production build pipeline (Makefile) | Task 6 |
| README + CLAUDE.md update | Task 7 |
| Final corpus smoke | Task 7 |

**Placeholder scan:** None of "TBD", "add appropriate", "fill in details". Every step has concrete code or a concrete command.

**Type consistency:**
- TS `Analysis.final_state.players` is `Record<string, PlayerFinalState>` (JSON keys are strings even though Go map keys are int). Verified consistent in ReplayPage (`a.final_state.players[String(p.id)]`).
- `Engagement.t_start_sec` / `t_end_sec` — used identically in TS types (Task 2), BuildOrderList (Task 4 — actually no, it doesn't read these), DensityChart (Task 5). Consistent.
- `Anomaly.t_start_sec` / `t_end_sec` — same.
- `DensityBin.counts[cat]` returns `number | undefined` in TS strict mode — DensityChart uses `?? 0` (Task 5). Consistent.
- `Color.r/g/b` lowercase keys — matches Go's `output.Color` JSON tags from the previous plan's Task 1 nit fix. Consistent.
- Engineer reading Task 5 only: types are imported from `../types`, which is fully defined in Task 2.

**Known carry-forward risks the engineer should know:**
1. `//go:embed all:dist` requires the embedded path to exist at compile time. Task 1.5 creates a stub. If someone runs `go build` without first running `make web-build`, the binary works but serves the stub. Acceptable.
2. The Makefile `web-build` copies `web/dist/.` (note the trailing slash + dot) — Linux/macOS behavior of `cp -R src/. dst/` is "copy contents into dst, not src itself". Verified.
3. `react-plotly.js` peer-depends on `plotly.js` — both in `dependencies` (Task 2.2). If npm complains about peer-deps, retry with `npm install --legacy-peer-deps`.
4. `web/dist` is gitignored — `npm run build` is needed before the bundle ships. CI/Makefile handles it; local dev too.
5. Removing the `github.com/a-h/templ` dep via `go mod tidy` (Task 1.7) — if Go reports the dep is still referenced somewhere, grep for stray `templ.` imports and delete them. Verify with `grep -r "a-h/templ" internal/ cmd/`.
6. The previous plan's `analysis.json` files in `~/Library/Application Support/0ad/replays/...` are still on disk. They're schema v1, fully compatible with the new code. No regen needed.
