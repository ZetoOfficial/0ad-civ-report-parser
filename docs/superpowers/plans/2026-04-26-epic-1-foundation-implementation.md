# Epic 1 — Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace single-file output (`<civ>_buildings_report.md`) with two-file pair (`<civ>_overview.md` + `<civ>_structree.md`) plus shared `common.md`, introducing `internal/config/` (JSON-based config), `internal/render/skeleton/` (text/template wrappers via `//go:embed`), and renaming `CivCode.OutputFile` to `BaseName`. No new parsers or content — pure infrastructure shift.

**Architecture:** Generator builds two markdown bodies per civ (overview + structree) plus one shared common body. Each body is wrapped by a thin skeleton template (info-block + `{{.Body}}` slot) and written to `<out-dir>/`. Config layered as defaults → optional `config.json` → CLI flags. Five-commit migration in this PR; each commit compiles and `go test ./...` passes.

**Tech Stack:** Go 1.25.5, stdlib only (`encoding/json`, `text/template`, `embed`, `flag`). No third-party deps. Tests via standard `testing` package.

**Spec:** `docs/superpowers/specs/2026-04-26-epic-1-foundation-design.md`

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `internal/config/config.go` | create | `Config` struct + `Defaults()` + `Load(path)` |
| `internal/config/config_test.go` | create | Unit tests for Defaults/Load |
| `internal/render/skeleton/skeleton.go` | create | `//go:embed FS` + `Render(name, data) (string, error)` |
| `internal/render/skeleton/skeleton_test.go` | create | Info-block substitution per template |
| `internal/render/skeleton/overview.tmpl` | create | Header + dateline + options + `{{.Body}}` |
| `internal/render/skeleton/structree.tmpl` | create | Same shape, structree heading |
| `internal/render/skeleton/common.tmpl` | create | Same shape, common heading |
| `internal/render/overview.go` | create | `renderOverview()` extracted from `report.go` |
| `internal/render/structree.go` | create | `renderStructree()` calling phases + units detail + summary |
| `internal/render/common.go` | create | `renderCommon()` — placeholder body for epic 1 |
| `internal/render/report.go` | modify | `Output` struct + `Generate()` orchestrator returning `Output` + `RenderCommon()` |
| `internal/civdata/codes.go` | modify | Rename `OutputFile` → `BaseName`; add `OverviewFile()`/`StructreeFile()` |
| `internal/civdata/codes_test.go` | modify | Add coverage for `BaseName` and file-name methods |
| `cmd/civreport/main.go` | modify | Add `--out-dir`, `--config`, `--print-basename`; remove `--out`; integrate `Config` + skeleton |
| `Makefile` | modify | `clean` targets new file names; new `golden-diff` target; remove `OUT` |
| `internal/render/golden_test.go` | modify (commits 3 & 5) | Adapt to new return type (commit 3); rewrite as structural test (commit 5) |
| `testdata/golden/germans_buildings_report.md` | delete (commit 5) | Legacy single-file golden, no longer applicable |
| `README.md` | modify | New CLI surface, new output file names |
| `CLAUDE.md` | modify | `config.yaml` → `config.json`; update file-name references |
| `docs/analysis.md` | modify | Same — `config.yaml` → `config.json` |

---

## Pre-flight check

- [ ] **Step 0.1: Confirm branch and clean tree**

Run:
```
git status -sb
```
Expected: branch `epic-1-foundation`. Untracked files `testdata/golden/germans_overview.md` and `germans_structree.md` are user-supplied references and stay untracked through this epic — do not commit them.

- [ ] **Step 0.2: Baseline tests green**

Run:
```
go test ./...
```
Expected: PASS (with `numeric_test` and `golden_test` skipping if gamedata unavailable; on Zeto's machine they run for real).

---

## Task 1: Config package

**Goal:** Introduce `internal/config/` with `Config` struct, `Defaults()`, and `Load(path string)`. JSON loader (no third-party YAML lib). Unused by anyone yet.

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1.1: Write failing test for `Defaults()` and `Load()`**

Create `internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

func TestDefaults(t *testing.T) {
	c := Defaults()
	if c.Gamedata != paths.DefaultGameDataRoot {
		t.Errorf("Gamedata = %q; want %q", c.Gamedata, paths.DefaultGameDataRoot)
	}
	if c.OutDir != "." {
		t.Errorf("OutDir = %q; want %q", c.OutDir, ".")
	}
	if c.Lang != "" {
		t.Errorf("Lang = %q; want empty", c.Lang)
	}
	if c.IncludeHistory {
		t.Errorf("IncludeHistory = true; want false")
	}
	if c.IncludeIcons {
		t.Errorf("IncludeIcons = true; want false")
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	c, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\") error: %v", err)
	}
	if c.Gamedata != paths.DefaultGameDataRoot {
		t.Errorf("expected defaults; Gamedata = %q", c.Gamedata)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/no-such-config.json")
	if err == nil {
		t.Fatal("expected error for missing file path")
	}
}

func TestLoad_PartialFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	body := `{"out_dir": "/tmp/out", "include_history": true}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.OutDir != "/tmp/out" {
		t.Errorf("OutDir = %q; want /tmp/out", c.OutDir)
	}
	if !c.IncludeHistory {
		t.Error("IncludeHistory = false; want true")
	}
	if c.Gamedata != paths.DefaultGameDataRoot {
		t.Errorf("Gamedata = %q; want default %q", c.Gamedata, paths.DefaultGameDataRoot)
	}
}

func TestLoad_BadJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
```

- [ ] **Step 1.2: Run tests — expect compile failure**

Run:
```
go test ./internal/config/...
```
Expected: FAIL with `undefined: Defaults` or similar (package does not exist).

- [ ] **Step 1.3: Implement `internal/config/config.go`**

Create the file:

```go
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

type Config struct {
	Gamedata       string `json:"gamedata"`
	OutDir         string `json:"out_dir"`
	Lang           string `json:"lang"`
	IncludeHistory bool   `json:"include_history"`
	IncludeIcons   bool   `json:"include_icons"`
}

func Defaults() Config {
	return Config{
		Gamedata: paths.DefaultGameDataRoot,
		OutDir:   ".",
	}
}

// Load reads a JSON config file at path and overlays it on Defaults().
// If path is empty, returns Defaults() with no error.
// If the file is missing or malformed, returns an error.
func Load(path string) (*Config, error) {
	c := Defaults()
	if path == "" {
		return &c, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}
	return &c, nil
}
```

- [ ] **Step 1.4: Run tests — expect PASS**

Run:
```
go test ./internal/config/...
```
Expected: PASS, all five tests green.

- [ ] **Step 1.5: Commit**

```
git add internal/config/config.go internal/config/config_test.go
git commit -m "$(cat <<'EOF'
config: add JSON loader and Config struct

Final field set (Gamedata, OutDir, Lang, IncludeHistory, IncludeIcons)
introduced in epic 1; Lang/IncludeHistory/IncludeIcons accepted from
JSON now, wired to CLI flags in later epics.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Skeleton package

**Goal:** Introduce `internal/render/skeleton/` with three thin `text/template` wrappers (`overview.tmpl`, `structree.tmpl`, `common.tmpl`) embedded via `//go:embed`. Single function `Render(name string, data Data) (string, error)`. Unused by anyone yet.

**Files:**
- Create: `internal/render/skeleton/skeleton.go`
- Create: `internal/render/skeleton/skeleton_test.go`
- Create: `internal/render/skeleton/overview.tmpl`
- Create: `internal/render/skeleton/structree.tmpl`
- Create: `internal/render/skeleton/common.tmpl`

- [ ] **Step 2.1: Write the three `.tmpl` files**

`internal/render/skeleton/overview.tmpl`:

```
# {{.CivName}} ({{.CivCodeUpper}}) — Civilization Overview

> Сгенерировано {{.Date}} утилитой civreport.
> Опции: include_history={{.IncludeHistory}}, include_icons={{.IncludeIcons}}, lang={{if .Lang}}{{.Lang}}{{else}}—{{end}}.
>
> Числовые значения — базовые из XML-шаблонов с применённым
> наследованием (`parent`-цепочки, миксины, `op="mul"`/`op="add"`,
> подстановки `{civ}`/`{native}`). Эффекты технологий и аур к статам
> не применены.

---

{{.Body}}
```

`internal/render/skeleton/structree.tmpl`:

```
# {{.CivName}} ({{.CivCodeUpper}}) — Structure Tree

> Сгенерировано {{.Date}} утилитой civreport.
> Опции: include_history={{.IncludeHistory}}, include_icons={{.IncludeIcons}}, lang={{if .Lang}}{{.Lang}}{{else}}—{{end}}.
>
> Числовые значения — базовые из XML-шаблонов с применённым наследованием.
> Эффекты технологий и аур к статам не применены.
> Глобальные модификаторы рангов и базовые дефолты — см. `common.md`.

---

{{.Body}}
```

`internal/render/skeleton/common.tmpl`:

```
# Common Reference (всеобщие данные для всех цив)

> Сгенерировано {{.Date}} утилитой civreport.
>
> Этот файл содержит данные, не зависящие от конкретной цивилизации:
> модификаторы Advanced/Elite, типы урона, типы ресурсов,
> статус-эффекты, дефолты базовых шаблонов.
> Цив-отчёты ссылаются якорями сюда вместо повторения.

---

{{.Body}}
```

- [ ] **Step 2.2: Write failing test for skeleton.Render**

Create `internal/render/skeleton/skeleton_test.go`:

```go
package skeleton

import (
	"strings"
	"testing"
)

func TestRender_Overview_Substitutions(t *testing.T) {
	d := Data{
		CivName:        "Spartans",
		CivCodeUpper:   "Spart",
		Date:           "2026-04-26",
		Lang:           "",
		IncludeHistory: false,
		IncludeIcons:   false,
		Body:           "BODY-MARKER",
	}
	out, err := Render("overview", d)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	musts := []string{
		"# Spartans (Spart) — Civilization Overview",
		"Сгенерировано 2026-04-26",
		"include_history=false",
		"lang=—",
		"BODY-MARKER",
	}
	for _, m := range musts {
		if !strings.Contains(out, m) {
			t.Errorf("Render(overview) missing %q in:\n%s", m, out)
		}
	}
}

func TestRender_Structree_LangSubstitution(t *testing.T) {
	d := Data{CivName: "Han Chinese", CivCodeUpper: "Han", Date: "2026-04-26", Lang: "ru", Body: "X"}
	out, err := Render("structree", d)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(out, "lang=ru") {
		t.Errorf("expected lang=ru in:\n%s", out)
	}
	if !strings.Contains(out, "# Han Chinese (Han) — Structure Tree") {
		t.Errorf("expected structree header; got:\n%s", out)
	}
}

func TestRender_Common_BodySlot(t *testing.T) {
	d := Data{Date: "2026-04-26", Body: "COMMON-MARKER"}
	out, err := Render("common", d)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(out, "# Common Reference") {
		t.Errorf("missing common header in:\n%s", out)
	}
	if !strings.Contains(out, "COMMON-MARKER") {
		t.Errorf("missing body in:\n%s", out)
	}
}

func TestRender_UnknownTemplate(t *testing.T) {
	_, err := Render("nope", Data{})
	if err == nil {
		t.Fatal("expected error for unknown template name")
	}
}
```

- [ ] **Step 2.3: Run tests — expect compile failure**

Run:
```
go test ./internal/render/skeleton/...
```
Expected: FAIL — package does not exist.

- [ ] **Step 2.4: Implement `skeleton.go`**

Create `internal/render/skeleton/skeleton.go`:

```go
package skeleton

import (
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed *.tmpl
var fsys embed.FS

type Data struct {
	CivName        string
	CivCodeUpper   string
	Date           string
	Lang           string
	IncludeHistory bool
	IncludeIcons   bool
	Body           string
}

var templates = map[string]*template.Template{}

func init() {
	for _, name := range []string{"overview", "structree", "common"} {
		raw, err := fsys.ReadFile(name + ".tmpl")
		if err != nil {
			panic(fmt.Sprintf("skeleton: embed read %s.tmpl: %v", name, err))
		}
		tpl, err := template.New(name).Parse(string(raw))
		if err != nil {
			panic(fmt.Sprintf("skeleton: parse %s.tmpl: %v", name, err))
		}
		templates[name] = tpl
	}
}

// Render evaluates the named template (overview/structree/common) with
// the given data and returns the rendered markdown.
func Render(name string, data Data) (string, error) {
	tpl, ok := templates[name]
	if !ok {
		return "", fmt.Errorf("skeleton: unknown template %q", name)
	}
	var sb strings.Builder
	if err := tpl.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("skeleton: execute %s: %w", name, err)
	}
	return sb.String(), nil
}
```

- [ ] **Step 2.5: Run tests — expect PASS**

Run:
```
go test ./internal/render/skeleton/...
```
Expected: PASS, four tests green.

- [ ] **Step 2.6: Commit**

```
git add internal/render/skeleton/
git commit -m "$(cat <<'EOF'
render: add skeleton package with text/template wrappers

Three thin wrappers (overview/structree/common) embedded via go:embed.
Each substitutes a {{.Body}} slot inside an info-block header. No
caller yet — wired up in commit 4 of this epic.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Split `report.go` into overview/structree/common; introduce `Output` type

**Goal:** Move existing `renderOverview` and `renderHeader` into `overview.go`; move `renderPhases`, `renderUnitsDetail`, `renderSummary` into `structree.go`; create `common.go` with placeholder. `Generator.Generate(civInfo)` returns `Output{Overview, Structree}` instead of one string. Add `Generator.RenderCommon()`. **Zero content changes** — only relocations and signature change.

**Files:**
- Create: `internal/render/overview.go`
- Create: `internal/render/structree.go`
- Create: `internal/render/common.go`
- Modify: `internal/render/report.go`
- Modify: `internal/render/golden_test.go` (minimal compile-fix)
- Modify: `cmd/civreport/main.go` (minimal compile-fix; full integration in Task 4)

- [ ] **Step 3.1: Define `Output` and refactor `report.go`**

Replace the contents of `internal/render/report.go` with:

```go
package render

import (
	"path/filepath"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/aura"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

type Generator struct {
	Layout   paths.Layout
	Resolver *tmpl.Resolver
	Catalog  *tech.Catalog
}

// Output holds the rendered markdown bodies for one civilization.
// They do not yet contain the skeleton header — main.go wraps them
// via internal/render/skeleton.
type Output struct {
	Overview  string
	Structree string
}

func NewGenerator(layout paths.Layout, resolver *tmpl.Resolver) *Generator {
	return &Generator{
		Layout:   layout,
		Resolver: resolver,
		Catalog:  tech.NewCatalog(layout.Technologies()),
	}
}

func (g *Generator) Generate(civInfo civdata.CivCode) (Output, error) {
	civ, err := civdata.LoadCiv(g.Layout.CivJSON(civInfo.Code))
	if err != nil {
		return Output{}, err
	}
	buildings, err := civdata.Buildings(g.Layout.StructuresOf(civInfo.Code), civInfo.Code, g.Resolver)
	if err != nil {
		return Output{}, err
	}
	units, err := civdata.Units(g.Layout.UnitsOf(civInfo.Code), civInfo.Code, g.Resolver)
	if err != nil {
		return Output{}, err
	}
	bonuses, err := g.Catalog.AllCivBonuses(civInfo.Code)
	if err != nil {
		return Output{}, err
	}
	notciv, err := g.Catalog.AllNotCiv(civInfo.Code)
	if err != nil {
		return Output{}, err
	}
	heroAuras, _ := aura.ListInDir(g.Layout.HeroAuras(), civInfo.Code+"_hero_")
	catafalqueAuras, _ := aura.ListInDir(g.Layout.CatafalqueAuras(), civInfo.Code+"_")

	return Output{
		Overview:  g.renderOverview(civInfo, civ, bonuses, notciv),
		Structree: g.renderStructree(civInfo.Code, buildings, units, heroAuras, catafalqueAuras),
	}, nil
}

// RenderCommon returns the body of the shared common.md.
// In epic 1 this is a placeholder — populated in epic 2.
func (g *Generator) RenderCommon() (string, error) {
	return renderCommonBody(), nil
}

func formatStartEntities(entities []civdata.StartEntity) string {
	parts := []string{}
	for _, e := range entities {
		base := filepath.Base(e.Template)
		name := strings.TrimPrefix(base, "structures/")
		name = strings.TrimPrefix(name, "units/")
		count := e.Count
		if count == 0 {
			count = 1
		}
		parts = append(parts, name+" ×"+itoa(count))
	}
	return strings.Join(parts, ", ")
}

func itoa(n int) string {
	// kept local to avoid an extra strconv import for this small helper
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits [20]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		digits[i] = '-'
	}
	return string(digits[i:])
}

func escapeTable(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
```

Note: `formatStartEntities` and `escapeTable` move here from the old `report.go`. The old file contained `fmt.Sprintf("%s ×%d", ...)`; we keep behaviour identical via a local `itoa`. (If you prefer, import `strconv` and use `strconv.Itoa(count)` — equivalent.)

- [ ] **Step 3.2: Create `internal/render/overview.go`**

```go
package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/i18n"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
)

// renderOverview returns the markdown body for the Civilization Overview tab.
// In epic 1 this preserves the previously-rendered "general info" section
// from the old report.go verbatim. New sections (Identity/Heroes/CivSpecific)
// arrive in epic 2.
func (g *Generator) renderOverview(info civdata.CivCode, civ *civdata.Civ, bonuses, notciv []*tech.Technology) string {
	var sb strings.Builder
	fmt.Fprintln(&sb, "## Общая информация о цивилизации")
	fmt.Fprintln(&sb)
	fmt.Fprintf(&sb, "- **Код:** `%s`\n", civ.Code)
	fmt.Fprintf(&sb, "- **Культура:** %s\n", civ.Culture())
	fmt.Fprintf(&sb, "- **Стартовые юниты:** %s\n", formatStartEntities(civ.StartEntities))
	if len(civ.TeamBonuses) > 0 {
		tb := civ.TeamBonuses[0]
		fmt.Fprintf(&sb, "- **Командный бонус (%s):** %s\n", tb.Name, tb.Description)
	}
	fmt.Fprintln(&sb)

	fmt.Fprintln(&sb, "### Цивилизационные бонусы")
	fmt.Fprintln(&sb)
	if len(civ.CivBonuses) == 0 && len(bonuses) == 0 {
		fmt.Fprintln(&sb, "*Особых цивилизационных бонусов не зафиксировано.*")
	} else {
		fmt.Fprintln(&sb, "| Бонус | Источник | Эффект |")
		fmt.Fprintln(&sb, "|-------|----------|--------|")
		for _, b := range civ.CivBonuses {
			fmt.Fprintf(&sb, "| %s | civ JSON | %s |\n",
				escapeTable(b.Name), escapeTable(b.Description))
		}
		for _, t := range bonuses {
			auto := ""
			if t.AutoResearch {
				auto = " (авто)"
			}
			tip := t.Tooltip
			if tip == "" {
				tip = i18n.DescribeModifications(t.Modifications)
			}
			if tip == "" {
				tip = t.Description
			}
			fmt.Fprintf(&sb, "| %s%s | %s | %s |\n",
				escapeTable(t.GenericName), auto, t.Name, escapeTable(tip))
		}
	}
	fmt.Fprintln(&sb)

	fmt.Fprintf(&sb, "### Технологии, НЕДОСТУПНЫЕ %s\n\n", info.NameRU)
	if len(notciv) == 0 {
		fmt.Fprintln(&sb, "Явных запретов через `notciv` для этой цивилизации не найдено.")
	} else {
		sort.Slice(notciv, func(i, j int) bool { return notciv[i].Name < notciv[j].Name })
		for _, t := range notciv {
			tip := t.Tooltip
			if tip == "" {
				tip = i18n.DescribeModifications(t.Modifications)
			}
			fmt.Fprintf(&sb, "- **%s** — %s\n", t.GenericName, tip)
		}
	}
	fmt.Fprintln(&sb)
	return sb.String()
}
```

- [ ] **Step 3.3: Create `internal/render/structree.go`**

This file inherits the existing phase/building/research/units/summary code from the old `report.go`, `units.go`, and `summary.go`. **Do not duplicate logic — just relocate.** Concretely:

1. Move `renderPhases`, `indexUnits`, `renderBuilding`, `renderTrains`, `collectTrainTokens`, `renderResearches`, `formatTechCost`, `requirementPhase` from old `report.go` into a new `structree.go`.
2. Re-target them as helpers used by the new `renderStructree` entry point.
3. Keep `internal/render/units.go` and `internal/render/summary.go` as they are (they already live in separate files; the entry point just calls into them).

The new `internal/render/structree.go`:

```go
package render

import (
	"fmt"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/aura"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/i18n"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

// renderStructree returns the markdown body for the Structure Tree tab.
// Epic 1: identical content to what the old report.go produced for
// phases + units detail + summary. New sections in epic 4.
func (g *Generator) renderStructree(civCode string, buildings, units []civdata.Entity, heroAuras, catafalqueAuras []*aura.Aura) string {
	var sb strings.Builder
	g.renderPhases(&sb, civCode, buildings, units)
	g.renderUnitsDetail(&sb, units, heroAuras, catafalqueAuras)
	g.renderSummary(&sb, buildings)
	return sb.String()
}

func (g *Generator) renderPhases(sb *strings.Builder, civCode string, buildings, units []civdata.Entity) {
	groups := civdata.GroupByPhase(buildings)
	phases := []struct {
		p     civdata.Phase
		title string
	}{
		{civdata.PhaseVillage, "VILLAGE PHASE"},
		{civdata.PhaseTown, "TOWN PHASE"},
		{civdata.PhaseCity, "CITY PHASE"},
	}
	unitByID := indexUnits(civCode, units)
	for _, ph := range phases {
		fmt.Fprintf(sb, "## %s\n\n", ph.title)
		list := groups[ph.p]
		if len(list) == 0 {
			fmt.Fprintln(sb, "*В этой фазе нет уникальных построек.*")
			fmt.Fprintln(sb)
			fmt.Fprintln(sb, "---")
			fmt.Fprintln(sb)
			continue
		}
		for _, b := range list {
			g.renderBuilding(sb, civCode, b, unitByID)
			fmt.Fprintln(sb, "---")
			fmt.Fprintln(sb)
		}
	}
}

func indexUnits(civCode string, units []civdata.Entity) map[string]civdata.Entity {
	m := make(map[string]civdata.Entity, len(units))
	for _, u := range units {
		base := u.Basename()
		m["units/"+civCode+"/"+base] = u
		m[base] = u
	}
	return m
}

func (g *Generator) renderBuilding(sb *strings.Builder, civCode string, b civdata.Entity, unitByID map[string]civdata.Entity) {
	name := FormatGenericName(b.Element)
	if name == "" {
		name = b.Basename()
	}
	fmt.Fprintf(sb, "### %s\n\n", name)
	fmt.Fprintln(sb, "| Параметр | Значение |")
	fmt.Fprintln(sb, "|----------|----------|")
	fmt.Fprintf(sb, "| Стоимость | %s |\n", FormatCost(b.Element))
	fmt.Fprintf(sb, "| Время постройки | %s |\n", FormatBuildTime(b.Element))
	fmt.Fprintf(sb, "| ОЗ | %s |\n", FormatHP(b.Element))
	if a := FormatArmor(b.Element); a != "—" {
		fmt.Fprintf(sb, "| Броня | %s |\n", a)
	}
	if pop := FormatPopulationBonus(b.Element); pop != "" {
		fmt.Fprintf(sb, "| Население | %s |\n", pop)
	}
	if t := FormatTerritory(b.Element); t != "" {
		fmt.Fprintf(sb, "| Территория | %s |\n", t)
	}
	if g := FormatGarrison(b.Element); g != "" {
		fmt.Fprintf(sb, "| Гарнизон | %s |\n", g)
	}
	if v := FormatVision(b.Element); v != "—" {
		fmt.Fprintf(sb, "| Обзор | %s |\n", v)
	}
	fmt.Fprintln(sb)

	g.renderTrains(sb, civCode, b, unitByID)
	g.renderResearches(sb, b)
}

func (g *Generator) renderTrains(sb *strings.Builder, civCode string, b civdata.Entity, unitByID map[string]civdata.Entity) {
	tokens := collectTrainTokens(b.Element)
	if len(tokens) == 0 {
		return
	}
	rows := []string{}
	for _, tok := range tokens {
		expanded := tmpl.SubstCiv(tok, civCode)
		u, ok := unitByID[expanded]
		if !ok {
			continue
		}
		row := fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s | %s |",
			FormatGenericName(u.Element),
			FormatCost(u.Element),
			FormatBuildTime(u.Element),
			FormatHP(u.Element),
			FormatAttackShort(u.Element),
			FormatArmorHPC(u.Element),
			FormatWalkSpeed(u.Element),
			FormatPopulation(u.Element),
		)
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return
	}
	fmt.Fprintln(sb, "#### Тренирует")
	fmt.Fprintln(sb)
	fmt.Fprintln(sb, "| Юнит | Стоимость | Время | ОЗ | Атака | Броня (H/P/C) | Скорость | Население |")
	fmt.Fprintln(sb, "|------|-----------|-------|-----|-------|---------------|----------|-----------|")
	for _, row := range rows {
		fmt.Fprintln(sb, row)
	}
	fmt.Fprintln(sb)
}

func collectTrainTokens(e *tmpl.Element) []string {
	tokens := []string{}
	if t := e.GetTokens("Trainer/Entities"); len(t) > 0 {
		tokens = append(tokens, t...)
	}
	if t := e.GetTokens("ProductionQueue/Entities"); len(t) > 0 {
		tokens = append(tokens, t...)
	}
	out := tokens[:0]
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if t == "" || strings.HasPrefix(t, "-") {
			continue
		}
		out = append(out, t)
	}
	return out
}

func (g *Generator) renderResearches(sb *strings.Builder, b civdata.Entity) {
	tokens := []string{}
	if t := b.Element.GetTokens("Trainer/Technologies"); len(t) > 0 {
		tokens = append(tokens, t...)
	}
	if t := b.Element.GetTokens("ProductionQueue/Technologies"); len(t) > 0 {
		tokens = append(tokens, t...)
	}
	cleaned := []string{}
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if t == "" || strings.HasPrefix(t, "-") {
			continue
		}
		cleaned = append(cleaned, t)
	}
	if len(cleaned) == 0 {
		return
	}
	fmt.Fprintln(sb, "#### Исследует")
	fmt.Fprintln(sb)
	fmt.Fprintln(sb, "| Технология | Стоимость | Время | Фаза | Эффект |")
	fmt.Fprintln(sb, "|-----------|-----------|-------|------|--------|")
	for _, name := range cleaned {
		t, err := g.Catalog.ByName(name)
		if err != nil {
			fmt.Fprintf(sb, "| %s | — | — | — | (не найдено) |\n", name)
			continue
		}
		cost := formatTechCost(t.Cost)
		time := "—"
		if t.ResearchTime > 0 {
			time = fmt.Sprintf("%s сек", i18n.FormatNumber(t.ResearchTime))
		}
		phase := requirementPhase(t.Requirements)
		eff := t.Tooltip
		if eff == "" {
			eff = i18n.DescribeModifications(t.Modifications)
		}
		gen := t.GenericName
		if gen == "" {
			gen = t.Name
		}
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s |\n", escapeTable(gen), cost, time, phase, escapeTable(eff))
	}
	fmt.Fprintln(sb)
}

func formatTechCost(c tech.Cost) string {
	parts := []string{}
	if c.Food != 0 {
		parts = append(parts, fmt.Sprintf("%d %s", c.Food, i18n.ResourceName("food")))
	}
	if c.Wood != 0 {
		parts = append(parts, fmt.Sprintf("%d %s", c.Wood, i18n.ResourceName("wood")))
	}
	if c.Stone != 0 {
		parts = append(parts, fmt.Sprintf("%d %s", c.Stone, i18n.ResourceName("stone")))
	}
	if c.Metal != 0 {
		parts = append(parts, fmt.Sprintf("%d %s", c.Metal, i18n.ResourceName("metal")))
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, ", ")
}

func requirementPhase(req tech.Requirements) string {
	if req == nil {
		return "—"
	}
	if v, ok := req["tech"]; ok {
		if s, ok := v.(string); ok {
			return i18n.PhaseRequirement(s)
		}
	}
	if all, ok := req["all"]; ok {
		if list, ok := all.([]any); ok {
			for _, item := range list {
				if m, ok := item.(map[string]any); ok {
					if v, ok := m["tech"].(string); ok {
						if p := i18n.PhaseRequirement(v); p != "" {
							return p
						}
					}
				}
			}
		}
	}
	return "—"
}
```

- [ ] **Step 3.4: Create `internal/render/common.go`**

```go
package render

// renderCommonBody returns the body of the shared common.md.
// Epic 1: placeholder. Epic 2 fills global modifiers, damage types,
// resource types, status effects, and base template defaults.
func renderCommonBody() string {
	return "## TODO\n\n" +
		"Этот файл будет наполнен в эпике 2: модификаторы Advanced/Elite,\n" +
		"типы урона, типы ресурсов, статус-эффекты, дефолты базовых шаблонов.\n\n" +
		"Цив-отчёты ссылаются якорями на разделы этого файла.\n"
}
```

- [ ] **Step 3.5: Compile-fix `cmd/civreport/main.go` (minimal)**

`Generator.Generate` now returns `Output`, not `string`. Update only the three call sites in `main.go` to keep the program compiling. Full integration of skeleton + config happens in Task 4. Replace each:

```go
body, err := gen.Generate(civInfo)
// ...
os.WriteFile(out, []byte(body), 0o600)
```

with:

```go
out, err := gen.Generate(civInfo)
// ...
combined := out.Overview + "\n" + out.Structree
os.WriteFile(outPath, []byte(combined), 0o600)
```

(temporary stitching — replaced in Task 4 with two-file write).

Specifically the patches:

In `runOne`:

```go
out, err := gen.Generate(info)
if err != nil {
    fail("generate %s: %v", info.Code, err)
}
body := out.Overview + "\n" + out.Structree
outPath := outFlag
if outPath == "" {
    outPath = info.OutputFile
}
if err := os.WriteFile(outPath, []byte(body), 0o600); err != nil {
    fail("write %s: %v", outPath, err)
}
abs, _ := filepath.Abs(outPath)
lines := strings.Count(body, "\n") + 1
fmt.Printf("OK %s → %s (%d lines)\n", info.Code, abs, lines)
```

In `runAll`:

```go
out, err := gen.Generate(civInfo)
if err != nil {
    fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", civInfo.Code, err)
    continue
}
body := out.Overview + "\n" + out.Structree
outPath := civInfo.OutputFile
if outFlag != "" {
    outPath = filepath.Join(outFlag, civInfo.OutputFile)
}
if err := os.WriteFile(outPath, []byte(body), 0o600); err != nil {
    fmt.Fprintf(os.Stderr, "WRITE %s: %v\n", civInfo.Code, err)
    continue
}
lines := strings.Count(body, "\n") + 1
fmt.Printf("OK %s → %s (%d lines)\n", civInfo.Code, outPath, lines)
```

In `runCheck`:

```go
out, err := gen.Generate(civInfo)
if err != nil {
    fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", civInfo.Code, err)
    failed++
    continue
}
body := out.Overview + "\n" + out.Structree
lines := strings.Count(body, "\n") + 1
ok := lines >= 100
mark := "OK"
if !ok {
    mark = "WARN"
    failed++
}
fmt.Printf("%s %s (%d lines)\n", mark, civInfo.Code, lines)
```

- [ ] **Step 3.6: Compile-fix `internal/render/golden_test.go` (minimal)**

Adapt the call site of `Generate` so the file compiles. Full rewrite happens in Task 5. Replace:

```go
body, err := g.Generate(info)
```

with:

```go
out, err := g.Generate(info)
```

and:

```go
lines := strings.Count(body, "\n") + 1
// ... if !strings.Contains(body, m) { ...
```

with:

```go
body := out.Overview + "\n" + out.Structree
lines := strings.Count(body, "\n") + 1
// ... if !strings.Contains(body, m) { ...
```

- [ ] **Step 3.7: Delete the old monolithic body of `report.go`**

The old `report.go` contained `renderHeader`, `renderOverview`, `renderPhases`, `indexUnits`, `renderBuilding`, `renderTrains`, `collectTrainTokens`, `renderResearches`, `formatTechCost`, `requirementPhase`, `escapeTable`, `formatStartEntities`. After Steps 3.1–3.3 these all live in their new files. Verify nothing duplicated:

```
grep -n "func.*renderHeader\|func.*renderOverview\|func.*renderPhases\|func.*renderBuilding\|func.*renderTrains\|func.*collectTrainTokens\|func.*renderResearches\|func.*formatTechCost\|func.*requirementPhase\|func.*formatStartEntities\|func.*escapeTable" internal/render/*.go
```

Each function name should appear exactly once. `renderHeader` is no longer needed (the skeleton template provides the heading) — it should appear zero times. If duplicates appear, delete from `report.go`.

- [ ] **Step 3.8: Build and run tests**

Run:
```
go build ./...
go test ./...
```

Expected: PASS. Numeric tests pass. Adapted golden_test passes (it now reads the stitched body).

- [ ] **Step 3.9: Commit**

```
git add internal/render/ cmd/civreport/main.go
git commit -m "$(cat <<'EOF'
render: split report into overview/structree/common; introduce Output

Generate() now returns Output{Overview, Structree}; RenderCommon()
added as a stub for epic 2. Existing rendering logic relocated
verbatim — no behavioural change. Old monolithic header (renderHeader)
removed; the skeleton template provides the heading from epic 4 of
this epic onwards.

main.go and golden_test.go are compile-patched to stitch the two
fields back into one body; full two-file integration arrives in the
next commit of this PR.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Rename `OutputFile` → `BaseName`; integrate Config and skeleton in CLI

**Goal:** Update `CivCode` field name and add file-name methods. Wire `internal/config` and `internal/render/skeleton` through `cmd/civreport/main.go`. Remove `--out`, add `--out-dir`, `--config`, `--print-basename`. After this task, `civreport <civ>` writes three files in the configured directory.

**Files:**
- Modify: `internal/civdata/codes.go`
- Modify: `internal/civdata/codes_test.go`
- Modify: `cmd/civreport/main.go`

- [ ] **Step 4.1: Add failing test for `BaseName` and file-name methods**

Append to `internal/civdata/codes_test.go`:

```go
func TestCivCode_FileNames(t *testing.T) {
	c, ok := ResolveCivInput("germ")
	if !ok {
		t.Fatal("germ not resolved")
	}
	if c.BaseName != "germans" {
		t.Errorf("BaseName = %q; want %q", c.BaseName, "germans")
	}
	if got := c.OverviewFile(); got != "germans_overview.md" {
		t.Errorf("OverviewFile() = %q", got)
	}
	if got := c.StructreeFile(); got != "germans_structree.md" {
		t.Errorf("StructreeFile() = %q", got)
	}
}

func TestCivCode_AllHaveBaseName(t *testing.T) {
	for _, c := range Civilizations {
		if c.BaseName == "" {
			t.Errorf("civ %q: BaseName empty", c.Code)
		}
	}
}
```

- [ ] **Step 4.2: Run tests — expect compile failure**

Run:
```
go test ./internal/civdata/...
```
Expected: FAIL — `BaseName`, `OverviewFile`, `StructreeFile` undefined.

- [ ] **Step 4.3: Update `internal/civdata/codes.go`**

Replace the file with:

```go
package civdata

import "strings"

type CivCode struct {
	Code     string
	NameEN   string
	NameRU   string
	BaseName string
	Aliases  []string
}

func (c CivCode) OverviewFile() string  { return c.BaseName + "_overview.md" }
func (c CivCode) StructreeFile() string { return c.BaseName + "_structree.md" }

var Civilizations = []CivCode{
	{"athen", "Athenians", "Афиняне", "athenians", []string{"афин", "athens"}},
	{"brit", "Britons", "Бритты", "britons", []string{"брит", "britons"}},
	{"cart", "Carthaginians", "Карфагеняне", "carthaginians", []string{"карфаг"}},
	{"gaul", "Gauls", "Галлы", "gauls", []string{"галл"}},
	{"germ", "Germans", "Германцы", "germans", []string{"герман", "german"}},
	{"han", "Han Chinese", "Хань", "han", []string{"хан", "китай", "chinese"}},
	{"iber", "Iberians", "Иберы", "iberians", []string{"ибер", "iberians"}},
	{"kush", "Kushites", "Кушиты", "kushites", []string{"куш", "kushite"}},
	{"mace", "Macedonians", "Македоняне", "macedonians", []string{"македон", "macedon"}},
	{"maur", "Mauryas", "Маурьи", "mauryas", []string{"маурь", "маур", "индиец", "mauryas"}},
	{"pers", "Achaemenid Persians", "Персы", "persians", []string{"перс", "persian", "ахеменид"}},
	{"ptol", "Ptolemies", "Птолемеи", "ptolemies", []string{"птолеме", "ptolemy", "египт"}},
	{"rome", "Romans", "Римляне", "romans", []string{"рим", "romans"}},
	{"sele", "Seleucids", "Селевкиды", "seleucids", []string{"селевкид", "seleucid"}},
	{"spart", "Spartans", "Спартанцы", "spartans", []string{"спарт", "spartan"}},
}

func ResolveCivInput(input string) (CivCode, bool) {
	if input == "" {
		return CivCode{}, false
	}
	low := strings.ToLower(strings.TrimSpace(input))
	for _, c := range Civilizations {
		if c.Code == low {
			return c, true
		}
	}
	for _, c := range Civilizations {
		for _, alias := range c.Aliases {
			if strings.Contains(low, strings.ToLower(alias)) {
				return c, true
			}
		}
	}
	return CivCode{}, false
}

func ByCode(code string) (CivCode, bool) {
	for _, c := range Civilizations {
		if c.Code == code {
			return c, true
		}
	}
	return CivCode{}, false
}
```

- [ ] **Step 4.4: Run civdata tests — expect PASS**

Run:
```
go test ./internal/civdata/...
```
Expected: PASS — three tests (existing `TestResolveCivInput` plus the two new ones).

- [ ] **Step 4.5: Verify build still passes overall**

Run:
```
go build ./...
```
Expected: FAIL with `info.OutputFile undefined` — `cmd/civreport/main.go` still references the old field name. Proceed to next step to fix.

- [ ] **Step 4.6: Rewrite `cmd/civreport/main.go`**

Replace the file entirely:

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/config"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/render"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/render/skeleton"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

func main() {
	var (
		gamedataFlag  string
		outDirFlag    string
		configFlag    string
		printBaseName bool
		all           bool
		check         bool
	)
	flag.StringVar(&gamedataFlag, "gamedata", "", "path to 0 A.D. mods/public root (overrides OAD_GAMEDATA_ROOT and config)")
	flag.StringVar(&outDirFlag, "out-dir", "", "output directory for generated files (default: from config or '.')")
	flag.StringVar(&configFlag, "config", "", "path to JSON config file")
	flag.BoolVar(&printBaseName, "print-basename", false, "print BaseName for the given civ and exit (used by Makefile)")
	flag.BoolVar(&all, "all", false, "generate reports for all 15 civilizations")
	flag.BoolVar(&check, "check", false, "smoke-check: parse all civs without writing files")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: civreport [flags] <civ>\n\n")
		fmt.Fprintf(os.Stderr, "Generate Russian-language overview + structure-tree reports\n")
		fmt.Fprintf(os.Stderr, "for one or more 0 A.D. civilizations.\n\n")
		fmt.Fprintf(os.Stderr, "Output: <civ>_overview.md, <civ>_structree.md, common.md.\n\n")
		fmt.Fprintf(os.Stderr, "Civ codes: athen, brit, cart, gaul, germ, han, iber, kush,\n")
		fmt.Fprintf(os.Stderr, "           mace, maur, pers, ptol, rome, sele, spart\n")
		fmt.Fprintf(os.Stderr, "Russian aliases also supported (спарт, афин, германцы, ...)\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	cfg, err := config.Load(configFlag)
	if err != nil {
		fail("config: %v", err)
	}
	// CLI overlays config: only flags actually present override.
	if gamedataFlag != "" {
		cfg.Gamedata = gamedataFlag
	} else if env := os.Getenv(paths.EnvGameDataRoot); env != "" {
		cfg.Gamedata = env
	}
	if outDirFlag != "" {
		cfg.OutDir = outDirFlag
	}

	if printBaseName {
		args := flag.Args()
		if len(args) != 1 {
			fail("--print-basename requires exactly one civ argument")
		}
		info, ok := civdata.ResolveCivInput(args[0])
		if !ok {
			fail("could not resolve civilization %q", args[0])
		}
		fmt.Println(info.BaseName)
		return
	}

	if _, err := os.Stat(filepath.Join(cfg.Gamedata, "simulation", "templates")); err != nil {
		fail("gamedata templates not found at %s/simulation/templates: %v", cfg.Gamedata, err)
	}

	idx, err := tmpl.NewIndex(filepath.Join(cfg.Gamedata, "simulation", "templates"))
	if err != nil {
		fail("template index: %v", err)
	}
	resolver := tmpl.NewResolver(idx)
	gen := render.NewGenerator(paths.Layout{Root: cfg.Gamedata}, resolver)

	switch {
	case check:
		runCheck(gen, cfg)
	case all:
		runAll(gen, cfg)
	default:
		args := flag.Args()
		if len(args) != 1 {
			flag.Usage()
			os.Exit(2)
		}
		runOne(gen, cfg, args[0])
	}
}

func runOne(gen *render.Generator, cfg *config.Config, input string) {
	info, ok := civdata.ResolveCivInput(input)
	if !ok {
		fail("could not resolve civilization %q. Try one of: athen, spart, germ, ...", input)
	}
	out, err := gen.Generate(info)
	if err != nil {
		fail("generate %s: %v", info.Code, err)
	}
	if err := writeCivFiles(cfg, info, out); err != nil {
		fail("write %s: %v", info.Code, err)
	}
	if err := writeCommon(cfg, gen); err != nil {
		fail("write common.md: %v", err)
	}
	abs, _ := filepath.Abs(cfg.OutDir)
	fmt.Printf("OK %s → %s/{%s,%s} + common.md\n", info.Code, abs, info.OverviewFile(), info.StructreeFile())
}

func runAll(gen *render.Generator, cfg *config.Config) {
	for _, info := range civdata.Civilizations {
		out, err := gen.Generate(info)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", info.Code, err)
			continue
		}
		if err := writeCivFiles(cfg, info, out); err != nil {
			fmt.Fprintf(os.Stderr, "WRITE %s: %v\n", info.Code, err)
			continue
		}
		fmt.Printf("OK %s → %s, %s\n", info.Code, info.OverviewFile(), info.StructreeFile())
	}
	if err := writeCommon(cfg, gen); err != nil {
		fmt.Fprintf(os.Stderr, "WRITE common.md: %v\n", err)
	} else {
		fmt.Println("OK common.md")
	}
}

func runCheck(gen *render.Generator, cfg *config.Config) {
	failed := 0
	for _, info := range civdata.Civilizations {
		out, err := gen.Generate(info)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", info.Code, err)
			failed++
			continue
		}
		ovLines := strings.Count(out.Overview, "\n") + 1
		stLines := strings.Count(out.Structree, "\n") + 1
		mark := "OK"
		if ovLines < 30 || stLines < 100 {
			mark = "WARN"
			failed++
		}
		fmt.Printf("%s %s (overview=%d, structree=%d)\n", mark, info.Code, ovLines, stLines)
	}
	if _, err := gen.RenderCommon(); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL common: %v\n", err)
		failed++
	}
	_ = cfg
	if failed > 0 {
		os.Exit(1)
	}
}

func writeCivFiles(cfg *config.Config, info civdata.CivCode, out render.Output) error {
	if err := ensureOutDir(cfg); err != nil {
		return err
	}
	date := time.Now().Format("2006-01-02") // Go reference layout = YYYY-MM-DD
	codeUpper := strings.ToUpper(info.Code[:1]) + info.Code[1:]
	overview, err := skeleton.Render("overview", skeleton.Data{
		CivName:        info.NameEN,
		CivCodeUpper:   codeUpper,
		Date:           date,
		Lang:           cfg.Lang,
		IncludeHistory: cfg.IncludeHistory,
		IncludeIcons:   cfg.IncludeIcons,
		Body:           out.Overview,
	})
	if err != nil {
		return fmt.Errorf("render overview skeleton: %w", err)
	}
	structree, err := skeleton.Render("structree", skeleton.Data{
		CivName:        info.NameEN,
		CivCodeUpper:   codeUpper,
		Date:           date,
		Lang:           cfg.Lang,
		IncludeHistory: cfg.IncludeHistory,
		IncludeIcons:   cfg.IncludeIcons,
		Body:           out.Structree,
	})
	if err != nil {
		return fmt.Errorf("render structree skeleton: %w", err)
	}
	if err := os.WriteFile(filepath.Join(cfg.OutDir, info.OverviewFile()), []byte(overview), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cfg.OutDir, info.StructreeFile()), []byte(structree), 0o600); err != nil {
		return err
	}
	return nil
}

func writeCommon(cfg *config.Config, gen *render.Generator) error {
	if err := ensureOutDir(cfg); err != nil {
		return err
	}
	body, err := gen.RenderCommon()
	if err != nil {
		return err
	}
	wrapped, err := skeleton.Render("common", skeleton.Data{
		Date: time.Now().Format("2006-01-02"),
		Body: body,
	})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(cfg.OutDir, "common.md"), []byte(wrapped), 0o600)
}

func ensureOutDir(cfg *config.Config) error {
	if cfg.OutDir == "" {
		cfg.OutDir = "."
	}
	return os.MkdirAll(cfg.OutDir, 0o755)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "civreport: "+format+"\n", args...)
	os.Exit(1)
}
```

- [ ] **Step 4.7: Compile-fix `internal/render/golden_test.go` to remove stale stitching**

Since `main.go` no longer concatenates Overview+Structree, the golden test still does for now. That's fine — it remains as the compile-stitch from Task 3 until Task 5 rewrites it.

- [ ] **Step 4.8: Build and smoke**

Run:
```
go build ./...
go test ./...
```
Expected: PASS.

- [ ] **Step 4.9: Local end-to-end smoke**

Run:
```
go run ./cmd/civreport --out-dir /tmp/civreport-smoke spart
ls -la /tmp/civreport-smoke/
```

Expected output:
```
spartans_overview.md
spartans_structree.md
common.md
```

Quickly inspect the first 12 lines of each:
```
head -12 /tmp/civreport-smoke/spartans_overview.md
head -12 /tmp/civreport-smoke/spartans_structree.md
head -12 /tmp/civreport-smoke/common.md
```

Each must show the skeleton header (`# Spartans (Spart) — ...`, `# Common Reference ...`) followed by `Сгенерировано <date>` and option line.

- [ ] **Step 4.10: Commit**

```
git add internal/civdata/codes.go internal/civdata/codes_test.go cmd/civreport/main.go
git commit -m "$(cat <<'EOF'
civdata,cmd: rename OutputFile to BaseName, write three files via skeleton

CivCode.OutputFile -> BaseName with OverviewFile()/StructreeFile()
helpers. main.go now loads Config (JSON), wraps Generate() output via
the skeleton package, and writes <civ>_overview.md, <civ>_structree.md,
and common.md to --out-dir. Removed --out (single-file); added
--out-dir, --config, --print-basename. --check reports per-file line
counts; thresholds 30 (overview) and 100 (structree).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Makefile + golden_test rewrite + docs sync

**Goal:** Update Makefile (remove `OUT`, add `--out-dir`, add `golden-diff` target, fix `clean`). Rewrite `golden_test.go` as a structural test against the new format. Delete the legacy single-file golden. Sync `README.md`, `CLAUDE.md`, `docs/analysis.md` to mention `config.json` and the new file pair.

**Files:**
- Modify: `Makefile`
- Modify: `internal/render/golden_test.go`
- Delete: `testdata/golden/germans_buildings_report.md`
- Modify: `README.md`
- Modify: `CLAUDE.md`
- Modify: `docs/analysis.md`

- [ ] **Step 5.1: Replace `Makefile`**

```
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

.PHONY: help build all-civs check test clean civ golden-diff $(CIVS)

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
	rm -f *_overview.md *_structree.md common.md
	rm -f *_buildings_report.md
	rm -rf $(BIN_DIR)
```

- [ ] **Step 5.2: Smoke Makefile targets**

Run:
```
make build
make spart
ls spartans_overview.md spartans_structree.md common.md
make clean
```
Expected: Targets succeed; the three files exist after `make spart`; `make clean` removes them.

Run:
```
make golden-diff CIV=germ 2>&1 | head -80
```
Expected: Two diff sections printed (large diff, since epic 1 doesn't fill content). Exit code 0.

- [ ] **Step 5.3: Rewrite `internal/render/golden_test.go` as structural**

Replace the file:

```go
package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

func TestGoldenGermStructure(t *testing.T) {
	skipIfNoGamedata(t)

	layout := paths.Layout{Root: gamedataRoot()}
	resolver := newResolver(t)
	g := NewGenerator(layout, resolver)

	info, ok := civdata.ResolveCivInput("germ")
	if !ok {
		t.Fatalf("germ resolution failed")
	}

	out, err := g.Generate(info)
	if err != nil {
		t.Fatalf("generate germ: %v", err)
	}

	overviewLines := strings.Count(out.Overview, "\n") + 1
	if overviewLines < 30 {
		t.Errorf("overview too short: %d lines (want >= 30)", overviewLines)
	}
	structreeLines := strings.Count(out.Structree, "\n") + 1
	if structreeLines < 100 {
		t.Errorf("structree too short: %d lines (want >= 100)", structreeLines)
	}

	overviewMust := []string{
		"## Общая информация о цивилизации",
		"- **Код:** `germ`",
		"### Цивилизационные бонусы",
		"### Технологии, НЕДОСТУПНЫЕ Германцы",
	}
	for _, m := range overviewMust {
		if !strings.Contains(out.Overview, m) {
			t.Errorf("overview missing %q", m)
		}
	}

	structreeMust := []string{
		"## VILLAGE PHASE",
		"## TOWN PHASE",
		"## CITY PHASE",
		"## Приложение: Детальная информация по типам юнитов",
		"## Приложение: Сводная таблица строимых зданий",
	}
	for _, m := range structreeMust {
		if !strings.Contains(out.Structree, m) {
			t.Errorf("structree missing %q", m)
		}
	}

	commonBody, err := g.RenderCommon()
	if err != nil {
		t.Fatalf("RenderCommon: %v", err)
	}
	if !strings.Contains(commonBody, "TODO") {
		t.Errorf("common body should mention TODO placeholder in epic 1")
	}

	wd, _ := os.Getwd()
	for _, f := range []string{"germans_overview.md", "germans_structree.md"} {
		path := filepath.Join(wd, "..", "..", "testdata", "golden", f)
		if _, err := os.Stat(path); err != nil {
			t.Logf("golden reference %s not present (ok in epic 1)", f)
			continue
		}
		if _, err := os.ReadFile(path); err != nil {
			t.Errorf("read golden %s: %v", f, err)
		}
	}
}
```

- [ ] **Step 5.4: Delete legacy golden**

Run:
```
git rm testdata/golden/germans_buildings_report.md
```
Expected: file removed and staged.

- [ ] **Step 5.5: Run all tests**

Run:
```
go test ./...
```
Expected: PASS.

- [ ] **Step 5.6: Update `README.md`**

Replace `README.md` with content describing the new CLI surface. Use this template:

```markdown
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
```

- [ ] **Step 5.7: Update `CLAUDE.md`**

Locate references to `config.yaml` and the old single-file output, replace:

Run:
```
grep -n "config.yaml\|buildings_report\|<civ>_overview.md.*<civ>_structree.md\|--out " CLAUDE.md
```

For each match, edit by hand:
- `config.yaml` → `config.json`
- `<civ>_buildings_report.md` references → describe new pair
- `--out` (single file) → `--out-dir`

The "Команды" section ought to mention `civreport --out-dir <dir>` instead of `--out <file>`. The "Опции в `config.yaml`" section becomes "Опции в `config.json`" with the same fields and valid JSON syntax instead of YAML.

- [ ] **Step 5.8: Update `docs/analysis.md`**

Run:
```
grep -n "config.yaml" docs/analysis.md
```

For each match, replace `config.yaml` with `config.json`. There is one explicit decision block referring to YAML — append a note: «Формат конфига: JSON (stdlib-only зависимости в проекте). См. `docs/superpowers/specs/2026-04-26-epic-1-foundation-design.md`.»

- [ ] **Step 5.9: Final test pass**

Run:
```
go test ./...
make build
make check
make golden-diff CIV=germ 2>&1 | tail -20
```
Expected: tests pass; build succeeds; check passes for all 15 civs (with WARN possibly due to short overview — that is OK, warning is informational); golden-diff prints expected large diff.

- [ ] **Step 5.10: Commit**

```
git add Makefile internal/render/golden_test.go README.md CLAUDE.md docs/analysis.md
git rm --cached testdata/golden/germans_buildings_report.md 2>/dev/null || true
git commit -m "$(cat <<'EOF'
makefile,test,docs: update for two-file output

- Makefile: --out-dir / --config / golden-diff target; clean covers
  new file names; --out removed.
- golden_test.go: structural test against the new Output struct;
  byte-diff deferred to epic 4.
- testdata/golden/germans_buildings_report.md: removed (legacy).
- README, CLAUDE.md, docs/analysis.md: config.yaml -> config.json,
  updated CLI surface and output file names.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Post-implementation verification

- [ ] **Step P.1: PR-ready check**

Run:
```
git log --oneline master..HEAD
```
Expected: 5 commits on top of master (`docs: epic 1 foundation design`, `config: ...`, `render: skeleton ...`, `render: split report ...`, `civdata,cmd: rename ...`, `makefile,test,docs: ...`). Note: `docs: epic 1 foundation design` was committed before this plan; total commits on the branch = 6.

- [ ] **Step P.2: Definition of Done re-check (from spec)**

For each item in the spec's Definition of Done, verify:

1. `civreport spart` writes three files to current dir → `make spart && ls spartans_*.md common.md`
2. `civreport --out-dir /tmp/out spart` writes to `/tmp/out/`
3. `civreport --all` produces 30 + 1 files
4. `civreport --check` smoke passes
5. `civreport --config ./config.json spart` reads config (write a test config with `out_dir=/tmp/cfg-test` and verify files land there)
6. All `make` targets work
7. `make golden-diff CIV=germ` runs and prints diff
8. `CivCode.OutputFile` removed, `BaseName` exists
9. `testdata/golden/germans_buildings_report.md` deleted; `germans_overview.md` and `germans_structree.md` remain on disk (untracked OK)
10. README, CLAUDE.md, analysis.md updated

- [ ] **Step P.3: Hand off**

Branch `epic-1-foundation` is ready for PR to `main`. Suggested PR title:

```
Epic 1 — Foundation: two-file output + config + skeleton
```

Body should reference the spec at `docs/superpowers/specs/2026-04-26-epic-1-foundation-design.md` and list the five commits with one-line descriptions.

---

## Self-review notes

- All five spec sections (skeleton, config, BaseName, output split, golden strategy) have explicit tasks: 2, 1, 4, 3, 5 respectively.
- No placeholders inside steps — every code block is self-contained and runnable.
- Type consistency: `Output{Overview, Structree string}`, `skeleton.Data{...}`, `config.Config{...}` — names referenced consistently across Tasks 1–5.
- Date format trick in Step 4.6 explicitly corrected to `time.Now().Format("2006-01-02")`.
- Common.md placeholder body is text-only ("TODO" + description) — no fake structure that would confuse epic 2 work.
- The compile-stitch (`out.Overview + "\n" + out.Structree`) introduced in Task 3 is removed by Task 4 (main.go) and Task 5 (golden_test.go) — no stitching survives the PR.
