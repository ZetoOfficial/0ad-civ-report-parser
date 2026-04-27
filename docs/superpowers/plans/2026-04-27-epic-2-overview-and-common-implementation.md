# Epic 2 — Overview + common.md Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Заполнить `<civ>_overview.md` контентом, эквивалентным игровой Civilization Overview tab, и наполнить `common.md` глобальной справочной информацией. Закрывает gaps 2/3/4/5/6/20/21/22 из `docs/analysis.md`.

**Architecture:** Расширяем существующие загрузчики (Player template через `tmpl.Resolver`, team-bonus aura через `aura.Load`) и tech-модель (per-mod `Affects`, `RequirementsTooltip`, `Replaces`, новый `GlobalAutoResearch`, `DescribeRequirements`). Переписываем `render/overview.go` с 8 обязательными секциями + опциональной History. Наполняем `render/common.go` 6 секциями (Advanced/Elite/Прочие auto/Damage/Resources/Status). Структура — последовательность 7 коммитов в одном PR на ветке `epic-2-overview-common`.

**Tech Stack:** Go 1.25.5, stdlib only. Существующая инфраструктура: `tmpl.Resolver` (XML с наследованием), `tech.Catalog` (lazy JSON loader), `aura.Load` (JSON-парсер аур), `i18n.translatePath` (~50 ключей перевода).

**Spec:** [docs/superpowers/specs/2026-04-27-epic-2-overview-and-common-design.md](../specs/2026-04-27-epic-2-overview-and-common-design.md)

**Branch:** `epic-2-overview-common` (создать перед стартом).

**Source data root:** `/Users/zeto/Projects/study/0ad/binaries/data/mods/public` (read-only). Все тесты, требующие данных, используют `skipIfNoGamedata(t)` или эквивалент.

---

## File Structure

| Path | Action | Responsibility |
|---|---|---|
| `internal/paths/paths.go` | modify | + `PlayerTemplate`, `TeamBonus`, `DamageTypes`, `Resources`, `StatusEffects` accessors |
| `internal/civdata/player.go` | create | `PlayerTemplate` struct + `LoadPlayerTemplate` + history-newline helper |
| `internal/civdata/player_test.go` | create | spart Identity loading |
| `internal/aura/teambonus.go` | create | `LoadTeamBonus(layout, civ)` тонкая обёртка над `aura.Load` |
| `internal/aura/teambonus_test.go` | create | spart team bonus AuraName + modifications |
| `internal/tech/tech.go` | modify | + `Modification.AffectsRaw json.RawMessage` + `AffectsList()`, + `Technology.RequirementsTooltip`, + `Technology.Replaces` |
| `internal/tech/tech_test.go` | create | Modification.Affects: string-форма и []string-форма |
| `internal/tech/auto.go` | create | `Catalog.GlobalAutoResearch()` — autoResearch без `requirements.civ` |
| `internal/tech/auto_test.go` | create | проверка списка |
| `internal/tech/requirements.go` | create | `DescribeRequirements(req)` — entity/all/any/notciv/civ/tech на русском |
| `internal/tech/requirements_test.go` | create | разные ветки requirements |
| `internal/i18n/modifier.go` | modify | + суффикс `(только {affects})` в `DescribeModification` |
| `internal/i18n/tech_name.go` | create | `TechDisplayName(t, civCode)` с lookup specificName |
| `internal/i18n/modifier_test.go` | create | оба новых поведения |
| `internal/render/report.go` | modify | + поле `Generator.IncludeHistory`, `Generate` подгружает Player+TeamBonus, `Output.Overview` получает новые входы через `renderOverview` |
| `internal/render/overview.go` | rewrite | 8 секций + опциональная History |
| `internal/render/overview_test.go` | create | структурный тест для spart |
| `internal/render/common_data.go` | create | загрузчики `loadDamageTypes/Resources/StatusEffects` |
| `internal/render/common.go` | rewrite | 6 секций |
| `internal/render/common_test.go` | create | структурный тест common |
| `internal/render/golden_test.go` | modify | удалить epic-1 проверки; добавить новые заголовки; порог 25→50 |
| `cmd/civreport/main.go` | modify | флаг `--include-history` → `cfg.IncludeHistory` → `Generator.IncludeHistory` |
| `README.md` | modify | секция «CLI options»: новый флаг; описание содержимого overview/common |
| `CLAUDE.md` | modify | актуализировать «Текущее»/«Целевое» (теперь почти совпадают) |

---

## Pre-flight

- [ ] **Step 0.1: Create epic-2 branch**

```bash
cd /Users/zeto/go/src/github.com/ZetoOfficial/0ad-civ-report-parser
git checkout master
git pull --ff-only
git checkout -b epic-2-overview-common
```

- [ ] **Step 0.2: Confirm baseline is green**

Run: `go build ./... && go test ./...`
Expected: PASS (numeric tests skip if gamedata absent — это ок).

- [ ] **Step 0.3: Confirm spec is committed on master**

Run: `git log --oneline master -- docs/superpowers/specs/2026-04-27-epic-2-overview-and-common-design.md`
Expected: одна строка с коммитом `docs: spec for epic 2 (overview + common.md)` (это `5a43ec3`).

---

## Task 1 — civdata: load Identity from `special/players/<civ>.xml`

**Closes:** gap 20.

**Files:**
- Modify: `internal/paths/paths.go` (add `PlayerTemplate`)
- Create: `internal/civdata/player.go`
- Create: `internal/civdata/player_test.go`

- [ ] **Step 1.1: Add `PlayerTemplate(civ)` to `paths.Layout`**

Edit `internal/paths/paths.go`. After the existing `CivJSON`/`TechJSON`/`AuraJSON` lines, append:

```go
func (l Layout) PlayerTemplate(civ string) string {
	return filepath.Join(l.Templates(), "special", "players", civ+".xml")
}
```

- [ ] **Step 1.2: Write failing test in `internal/civdata/player_test.go`**

```go
package civdata

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

func gamedataRoot() string {
	if env := os.Getenv("OAD_GAMEDATA_ROOT"); env != "" {
		return env
	}
	return "/Users/zeto/Projects/study/0ad/binaries/data/mods/public"
}

func skipIfNoGamedata(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(gamedataRoot(), "simulation/templates")); err != nil {
		t.Skipf("gamedata unavailable: %v", err)
	}
}

func TestLoadPlayerTemplate_Spart(t *testing.T) {
	skipIfNoGamedata(t)

	layout := paths.Layout{Root: gamedataRoot()}
	idx, err := tmpl.NewIndex(filepath.Join(gamedataRoot(), "simulation/templates"))
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	resolver := tmpl.NewResolver(idx)

	pt, err := LoadPlayerTemplate(layout, "spart", resolver)
	if err != nil {
		t.Fatalf("LoadPlayerTemplate: %v", err)
	}
	if pt.GenericName != "Spartans" {
		t.Errorf("GenericName = %q; want %q", pt.GenericName, "Spartans")
	}
	if !strings.HasPrefix(pt.History, "Sparta was a prominent city-state") {
		t.Errorf("History prefix mismatch: %q", pt.History[:min(60, len(pt.History))])
	}
	if pt.IconPath != "emblems/emblem_spartans.png" {
		t.Errorf("IconPath = %q", pt.IconPath)
	}
	// History contains literal "\n" sequences in source; LoadPlayerTemplate
	// should convert them to real "\n\n" paragraph breaks.
	if strings.Contains(pt.History, `\n`) {
		t.Errorf("History still contains literal \\n: %q", pt.History)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
```

- [ ] **Step 1.3: Run test, expect failure**

Run: `go test ./internal/civdata/ -run TestLoadPlayerTemplate_Spart -v`
Expected: FAIL with "undefined: LoadPlayerTemplate".

- [ ] **Step 1.4: Create `internal/civdata/player.go`**

```go
package civdata

import (
	"fmt"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

type PlayerTemplate struct {
	GenericName string
	History     string
	IconPath    string
}

// LoadPlayerTemplate resolves templates/special/players/<civ>.xml through
// the given resolver (so parent="template_player" inheritance is applied),
// then extracts Identity/{GenericName,History,Icon}.
//
// History fields in the source files store paragraph breaks as the
// literal two-character sequence "\n". We convert that to real markdown
// paragraph breaks ("\n\n") so the rendered overview reads naturally.
func LoadPlayerTemplate(layout paths.Layout, civCode string,
	resolver *tmpl.Resolver) (*PlayerTemplate, error) {
	path := layout.PlayerTemplate(civCode)
	el, err := resolver.Resolve(path)
	if err != nil {
		return nil, fmt.Errorf("resolve player template %s: %w", path, err)
	}
	return &PlayerTemplate{
		GenericName: el.GetText("Identity/GenericName"),
		History:     unescapeNewlines(el.GetText("Identity/History")),
		IconPath:    el.GetText("Identity/Icon"),
	}, nil
}

// unescapeNewlines converts the literal two-character sequence \n into a
// markdown paragraph break (real \n\n). It only touches that exact pattern;
// real newlines from the XML stay untouched.
func unescapeNewlines(s string) string {
	return strings.ReplaceAll(s, `\n`, "\n\n")
}
```

- [ ] **Step 1.5: Run test, expect success**

Run: `go test ./internal/civdata/ -run TestLoadPlayerTemplate_Spart -v`
Expected: PASS (or SKIP if gamedata unavailable — also ok in CI).

- [ ] **Step 1.6: Run the full package tests to confirm no regressions**

Run: `go test ./internal/civdata/... -v`
Expected: PASS.

- [ ] **Step 1.7: Commit**

```bash
git add internal/paths/paths.go internal/civdata/player.go internal/civdata/player_test.go
git commit -m "civdata: load Identity from special/players/<civ>.xml"
```

---

## Task 2 — aura: add `LoadTeamBonus` single-file loader

**Closes:** gap 2 (data loader half).

**Files:**
- Modify: `internal/paths/paths.go` (add `TeamBonus`)
- Create: `internal/aura/teambonus.go`
- Create: `internal/aura/teambonus_test.go`

- [ ] **Step 2.1: Add `TeamBonus(civ)` to `paths.Layout`**

Edit `internal/paths/paths.go`. After `PlayerTemplate`, append:

```go
func (l Layout) TeamBonus(civ string) string {
	return filepath.Join(l.Auras(), "teambonuses", civ+"_player_teambonus.json")
}
```

- [ ] **Step 2.2: Write failing test in `internal/aura/teambonus_test.go`**

```go
package aura

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

func gamedataRoot() string {
	if env := os.Getenv("OAD_GAMEDATA_ROOT"); env != "" {
		return env
	}
	return "/Users/zeto/Projects/study/0ad/binaries/data/mods/public"
}

func skipIfNoGamedata(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(gamedataRoot(), "simulation/templates")); err != nil {
		t.Skipf("gamedata unavailable: %v", err)
	}
}

func TestLoadTeamBonus_Spart(t *testing.T) {
	skipIfNoGamedata(t)

	layout := paths.Layout{Root: gamedataRoot()}
	a, err := LoadTeamBonus(layout, "spart")
	if err != nil {
		t.Fatalf("LoadTeamBonus: %v", err)
	}
	if a.AuraName != "Peloponnesian League" {
		t.Errorf("AuraName = %q; want %q", a.AuraName, "Peloponnesian League")
	}
	if a.Type != "global" {
		t.Errorf("Type = %q; want %q", a.Type, "global")
	}
	// Spartans team bonus = heroes free across 4 resources.
	if len(a.Modifications) != 4 {
		t.Errorf("Modifications len = %d; want 4", len(a.Modifications))
	}
	for _, m := range a.Modifications {
		if m.Replace == nil {
			t.Errorf("modification %s: expected Replace to be set", m.Value)
		}
	}
}
```

- [ ] **Step 2.3: Run test, expect failure**

Run: `go test ./internal/aura/ -run TestLoadTeamBonus_Spart -v`
Expected: FAIL with "undefined: LoadTeamBonus".

- [ ] **Step 2.4: Create `internal/aura/teambonus.go`**

```go
package aura

import (
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

// LoadTeamBonus reads data/auras/teambonuses/<civ>_player_teambonus.json
// and returns it as an *Aura. This is the structured source for the
// "Командный бонус" block in the civilization overview.
//
// Returns an os.PathError if the file is missing — callers may treat
// missing as "civ has no team bonus" if appropriate. In R28 every civ
// has exactly one team bonus.
func LoadTeamBonus(layout paths.Layout, civCode string) (*Aura, error) {
	return Load(layout.TeamBonus(civCode))
}
```

- [ ] **Step 2.5: Run test, expect success**

Run: `go test ./internal/aura/ -run TestLoadTeamBonus_Spart -v`
Expected: PASS.

- [ ] **Step 2.6: Full package tests**

Run: `go test ./internal/aura/... -v`
Expected: PASS.

- [ ] **Step 2.7: Commit**

```bash
git add internal/paths/paths.go internal/aura/teambonus.go internal/aura/teambonus_test.go
git commit -m "aura: add LoadTeamBonus single-file loader"
```

---

## Task 3 — tech: per-mod Affects, RequirementsTooltip, GlobalAutoResearch, DescribeRequirements

**Closes:** gaps 3, 4, 6 (tech-model side).

**Files:**
- Modify: `internal/tech/tech.go`
- Create: `internal/tech/tech_test.go`
- Create: `internal/tech/auto.go`
- Create: `internal/tech/auto_test.go`
- Create: `internal/tech/requirements.go`
- Create: `internal/tech/requirements_test.go`

- [ ] **Step 3.1: Extend `Modification` and `Technology`**

Edit `internal/tech/tech.go`.

Replace the `Modification` struct (lines 12-17) with:

```go
type Modification struct {
	Value      string          `json:"value"`
	Multiply   float64         `json:"multiply,omitempty"`
	Add        float64         `json:"add,omitempty"`
	Replace    any             `json:"replace,omitempty"`
	AffectsRaw json.RawMessage `json:"affects,omitempty"`
}

// AffectsList parses the per-modification "affects" field, which the JSON
// source may store either as a single string ("Melee") or an array
// (["Melee", "Cavalry"]). Returns nil if the field is absent.
func (m Modification) AffectsList() []string {
	if len(m.AffectsRaw) == 0 {
		return nil
	}
	var s string
	if err := json.Unmarshal(m.AffectsRaw, &s); err == nil {
		if s == "" {
			return nil
		}
		return []string{s}
	}
	var arr []string
	if err := json.Unmarshal(m.AffectsRaw, &arr); err == nil {
		return arr
	}
	return nil
}
```

In the `Technology` struct, after the existing `Pair`/`Top`/`Bottom`/`Icon`/`ReplacedBy` fields, add:

```go
	RequirementsTooltip string   `json:"requirementsTooltip"`
	Replaces            []string `json:"replaces"`
```

Final field block of `Technology` should read (only the additions are new — keep existing fields in their original order):

```go
type Technology struct {
	Name                string
	Path                string
	GenericName         string         `json:"genericName"`
	Description         string         `json:"description"`
	SpecificName        map[string]any `json:"specificName,omitempty"`
	AutoResearch        bool           `json:"autoResearch"`
	Cost                Cost           `json:"cost"`
	ResearchTime        float64        `json:"researchTime"`
	Tooltip             string         `json:"tooltip"`
	Modifications       []Modification `json:"modifications"`
	Affects             []string       `json:"affects"`
	Requirements        Requirements   `json:"requirements"`
	Supersedes          string         `json:"supersedes"`
	Pair                string         `json:"pair"`
	Top                 string         `json:"top"`
	Bottom              string         `json:"bottom"`
	Icon                string         `json:"icon"`
	ReplacedBy          string         `json:"replacedBy"`
	RequirementsTooltip string         `json:"requirementsTooltip"`
	Replaces            []string       `json:"replaces"`
}
```

- [ ] **Step 3.2: Write failing test for Modification.AffectsList**

Create `internal/tech/tech_test.go`:

```go
package tech

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestModification_AffectsList_String(t *testing.T) {
	raw := `{"value":"Attack/Melee/Damage/Hack","multiply":1.1,"affects":"Melee"}`
	var m Modification
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := m.AffectsList()
	want := []string{"Melee"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("AffectsList = %v; want %v", got, want)
	}
}

func TestModification_AffectsList_Array(t *testing.T) {
	raw := `{"value":"Health/Max","multiply":1.25,"affects":["Melee","Cavalry"]}`
	var m Modification
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := m.AffectsList()
	want := []string{"Melee", "Cavalry"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("AffectsList = %v; want %v", got, want)
	}
}

func TestModification_AffectsList_Absent(t *testing.T) {
	raw := `{"value":"Health/Max","add":10}`
	var m Modification
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := m.AffectsList(); got != nil {
		t.Errorf("AffectsList = %v; want nil", got)
	}
}
```

- [ ] **Step 3.3: Run tests**

Run: `go test ./internal/tech/ -run TestModification_AffectsList -v`
Expected: PASS (the struct edit from Step 3.1 already enables this).

- [ ] **Step 3.4: Create `internal/tech/auto.go`**

```go
package tech

import (
	"path/filepath"
	"sort"
	"strings"
)

// GlobalAutoResearch returns autoResearch technologies from the root
// technologies/ directory whose requirements have no civ filter — i.e.
// they apply automatically to every civ. In R28 this includes
// unit_advanced, unit_elite, phase_village, soldier_ranged_experience,
// upgrade_rank_advanced_mercenary.
//
// Civ-specific bonuses live in technologies/civbonuses/ and are returned
// by AllCivBonuses(civ); they are not included here even though they
// also have autoResearch=true.
func (c *Catalog) GlobalAutoResearch() ([]*Technology, error) {
	matches, err := filepath.Glob(filepath.Join(c.dir, "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	out := []*Technology{}
	for _, p := range matches {
		// Skip civbonuses — they live in a subdirectory but Glob with
		// "*.json" already excludes subdirs; defensive check kept for
		// readability.
		if strings.Contains(p, string(filepath.Separator)+"civbonuses"+string(filepath.Separator)) {
			continue
		}
		t, err := Load(p)
		if err != nil {
			return nil, err
		}
		if !t.AutoResearch {
			continue
		}
		if RequiresCiv(t.Requirements) != "" {
			continue
		}
		out = append(out, t)
	}
	return out, nil
}
```

- [ ] **Step 3.5: Write failing test for GlobalAutoResearch**

Create `internal/tech/auto_test.go`:

```go
package tech

import (
	"os"
	"path/filepath"
	"testing"
)

func gamedataRoot() string {
	if env := os.Getenv("OAD_GAMEDATA_ROOT"); env != "" {
		return env
	}
	return "/Users/zeto/Projects/study/0ad/binaries/data/mods/public"
}

func skipIfNoGamedata(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(gamedataRoot(), "simulation/templates")); err != nil {
		t.Skipf("gamedata unavailable: %v", err)
	}
}

func TestGlobalAutoResearch_R28Set(t *testing.T) {
	skipIfNoGamedata(t)

	c := NewCatalog(filepath.Join(gamedataRoot(), "simulation", "data", "technologies"))
	got, err := c.GlobalAutoResearch()
	if err != nil {
		t.Fatalf("GlobalAutoResearch: %v", err)
	}
	gotNames := map[string]bool{}
	for _, t := range got {
		gotNames[t.Name] = true
	}
	mustContain := []string{
		"unit_advanced",
		"unit_elite",
		"phase_village",
		"soldier_ranged_experience",
		"upgrade_rank_advanced_mercenary",
	}
	for _, name := range mustContain {
		if !gotNames[name] {
			t.Errorf("GlobalAutoResearch missing %q (got %v)", name, mapKeys(gotNames))
		}
	}
	mustExclude := []string{
		"spart_population",  // civbonuses subdir, RequiresCiv != ""
		"maur_elephants",    // civbonuses subdir, RequiresCiv != ""
	}
	for _, name := range mustExclude {
		if gotNames[name] {
			t.Errorf("GlobalAutoResearch unexpectedly contains %q", name)
		}
	}
}

func mapKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
```

- [ ] **Step 3.6: Run test**

Run: `go test ./internal/tech/ -run TestGlobalAutoResearch -v`
Expected: PASS (or SKIP if no gamedata).

- [ ] **Step 3.7: Create `internal/tech/requirements.go`**

```go
package tech

import (
	"fmt"
	"sort"
	"strings"
)

// DescribeRequirements renders a Requirements tree as human-readable
// Russian text. Used as the fallback when a tech does not provide its
// own requirementsTooltip. Handles entity/all/any/notciv/civ/tech.
//
// Examples:
//   {"tech":"phase_town"}
//     → "технология: phase_town"
//   {"entity":{"class":"Village","number":5}}
//     → "5+ зданий класса Village"
//   {"all":[{"tech":"phase_town"},{"entity":{"class":"Village","number":5}}]}
//     → "phase_town И 5+ зданий класса Village"
func DescribeRequirements(req Requirements) string {
	if req == nil {
		return ""
	}
	parts := []string{}
	for _, key := range sortedKeys(req) {
		switch key {
		case "civ":
			if v, ok := req["civ"].(string); ok {
				parts = append(parts, "цивилизация: "+v)
			}
		case "notciv":
			if list := stringOrList(req["notciv"]); len(list) > 0 {
				parts = append(parts, "не для цив: "+strings.Join(list, ", "))
			}
		case "tech":
			if v, ok := req["tech"].(string); ok {
				parts = append(parts, "технология: "+v)
			}
		case "entity":
			parts = append(parts, describeEntity(req["entity"]))
		case "all":
			parts = append(parts, describeJoin(req["all"], " И "))
		case "any":
			parts = append(parts, describeJoin(req["any"], " ИЛИ "))
		}
	}
	out := []string{}
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, "; ")
}

func describeEntity(v any) string {
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	class, _ := m["class"].(string)
	number := 0
	switch n := m["number"].(type) {
	case float64:
		number = int(n)
	case int:
		number = n
	}
	switch {
	case class != "" && number > 0:
		return fmt.Sprintf("%d+ зданий класса %s", number, class)
	case class != "":
		return "класс: " + class
	}
	return ""
}

func describeJoin(v any, sep string) string {
	list, ok := v.([]any)
	if !ok {
		return ""
	}
	parts := []string{}
	for _, item := range list {
		if m, ok := item.(map[string]any); ok {
			if s := DescribeRequirements(Requirements(m)); s != "" {
				parts = append(parts, s)
			}
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "(" + strings.Join(parts, sep) + ")"
}

func stringOrList(v any) []string {
	switch x := v.(type) {
	case string:
		if x == "" {
			return nil
		}
		return []string{x}
	case []any:
		out := []string{}
		for _, e := range x {
			if s, ok := e.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func sortedKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
```

- [ ] **Step 3.8: Write failing test for DescribeRequirements**

Create `internal/tech/requirements_test.go`:

```go
package tech

import (
	"strings"
	"testing"
)

func TestDescribeRequirements_Tech(t *testing.T) {
	got := DescribeRequirements(Requirements{"tech": "phase_town"})
	if got != "технология: phase_town" {
		t.Errorf("got %q", got)
	}
}

func TestDescribeRequirements_Entity(t *testing.T) {
	req := Requirements{
		"entity": map[string]any{"class": "Village", "number": float64(5)},
	}
	got := DescribeRequirements(req)
	want := "5+ зданий класса Village"
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestDescribeRequirements_All(t *testing.T) {
	req := Requirements{
		"all": []any{
			map[string]any{"tech": "phase_town"},
			map[string]any{"entity": map[string]any{"class": "Village", "number": float64(5)}},
		},
	}
	got := DescribeRequirements(req)
	if !strings.Contains(got, "phase_town") || !strings.Contains(got, "Village") {
		t.Errorf("missing branches in %q", got)
	}
	if !strings.Contains(got, " И ") {
		t.Errorf("missing AND separator in %q", got)
	}
}

func TestDescribeRequirements_Any(t *testing.T) {
	req := Requirements{
		"any": []any{
			map[string]any{"tech": "tech_a"},
			map[string]any{"tech": "tech_b"},
		},
	}
	got := DescribeRequirements(req)
	if !strings.Contains(got, " ИЛИ ") {
		t.Errorf("missing OR separator in %q", got)
	}
}

func TestDescribeRequirements_NotCiv(t *testing.T) {
	req := Requirements{"notciv": []any{"spart", "athen"}}
	got := DescribeRequirements(req)
	if !strings.Contains(got, "не для цив") {
		t.Errorf("missing notciv prefix in %q", got)
	}
	if !strings.Contains(got, "spart") || !strings.Contains(got, "athen") {
		t.Errorf("missing civs in %q", got)
	}
}

func TestDescribeRequirements_Empty(t *testing.T) {
	if got := DescribeRequirements(nil); got != "" {
		t.Errorf("nil → %q; want empty", got)
	}
	if got := DescribeRequirements(Requirements{}); got != "" {
		t.Errorf("empty map → %q; want empty", got)
	}
}
```

- [ ] **Step 3.9: Run tests**

Run: `go test ./internal/tech/... -v`
Expected: PASS (auto_test may SKIP if no gamedata; rest run unconditionally).

- [ ] **Step 3.10: Build all to confirm no upstream regressions**

Run: `go build ./...`
Expected: success.

- [ ] **Step 3.11: Run full test suite**

Run: `go test ./...`
Expected: PASS / SKIP.

- [ ] **Step 3.12: Commit**

```bash
git add internal/tech/
git commit -m "tech: per-mod Affects, RequirementsTooltip, GlobalAutoResearch, DescribeRequirements"
```

---

## Task 4 — i18n: per-mod affects suffix; TechDisplayName(specificName)

**Closes:** gap 5; finishes gap 6 user-facing rendering.

**Files:**
- Modify: `internal/i18n/modifier.go`
- Create: `internal/i18n/tech_name.go`
- Create: `internal/i18n/modifier_test.go`

- [ ] **Step 4.1: Update `DescribeModification` in `internal/i18n/modifier.go`**

Replace the current `DescribeModification` function (lines 10-27) with:

```go
func DescribeModification(m tech.Modification) string {
	target := translatePath(m.Value)
	var body string
	switch {
	case m.Multiply != 0:
		body = fmt.Sprintf("%s %s", target, FormatPercent(m.Multiply))
	case m.Add != 0:
		sign := "+"
		val := m.Add
		if val < 0 {
			sign = "−"
			val = -val
		}
		body = fmt.Sprintf("%s %s%s", target, sign, FormatNumber(val))
	case m.Replace != nil:
		body = fmt.Sprintf("%s = %v", target, m.Replace)
	default:
		body = target
	}
	if affects := m.AffectsList(); len(affects) > 0 {
		body += fmt.Sprintf(" (только %s)", strings.Join(affects, "+"))
	}
	return body
}
```

Note: keep `DescribeModifications` plural function unchanged; it just calls the singular.

- [ ] **Step 4.2: Create `internal/i18n/tech_name.go`**

```go
package i18n

import (
	"fmt"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
)

// TechDisplayName returns the human-readable tech name for a given civ.
// If specificName has an entry for the civ code, it is appended in
// parentheses ("Town Phase (локально: Astiteia)"). Falls back to
// GenericName, then Name.
func TechDisplayName(t *tech.Technology, civCode string) string {
	if t == nil {
		return ""
	}
	name := t.GenericName
	if name == "" {
		name = t.Name
	}
	if t.SpecificName == nil {
		return name
	}
	v, ok := t.SpecificName[civCode]
	if !ok {
		return name
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return name
	}
	return fmt.Sprintf("%s (локально: %s)", name, s)
}
```

- [ ] **Step 4.3: Write failing tests in `internal/i18n/modifier_test.go`**

```go
package i18n

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
)

func TestDescribeModification_AffectsSuffix_String(t *testing.T) {
	raw := `{"value":"Attack/Melee/Damage/Hack","multiply":1.1,"affects":"Melee"}`
	var m tech.Modification
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := DescribeModification(m)
	if !strings.Contains(got, "(только Melee)") {
		t.Errorf("missing suffix in %q", got)
	}
	if !strings.Contains(got, "+10%") {
		t.Errorf("missing percent in %q", got)
	}
}

func TestDescribeModification_AffectsSuffix_Array(t *testing.T) {
	raw := `{"value":"Health/Max","multiply":1.25,"affects":["Melee","Cavalry"]}`
	var m tech.Modification
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := DescribeModification(m)
	if !strings.Contains(got, "(только Melee+Cavalry)") {
		t.Errorf("missing combined suffix in %q", got)
	}
}

func TestDescribeModification_NoAffects_NoSuffix(t *testing.T) {
	raw := `{"value":"Health/Max","multiply":1.25}`
	var m tech.Modification
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := DescribeModification(m)
	if strings.Contains(got, "только") {
		t.Errorf("unexpected suffix in %q", got)
	}
}

func TestTechDisplayName_NoSpecificName(t *testing.T) {
	te := &tech.Technology{Name: "x", GenericName: "Phase Town"}
	if got := TechDisplayName(te, "spart"); got != "Phase Town" {
		t.Errorf("got %q", got)
	}
}

func TestTechDisplayName_WithSpecificName(t *testing.T) {
	te := &tech.Technology{
		Name:        "phase_town",
		GenericName: "Town Phase",
		SpecificName: map[string]any{
			"spart": "Astiteia",
			"athen": "Astuteia",
		},
	}
	got := TechDisplayName(te, "spart")
	if got != "Town Phase (локально: Astiteia)" {
		t.Errorf("got %q", got)
	}
}

func TestTechDisplayName_OtherCivFallsBack(t *testing.T) {
	te := &tech.Technology{
		Name:         "phase_town",
		GenericName:  "Town Phase",
		SpecificName: map[string]any{"spart": "Astiteia"},
	}
	got := TechDisplayName(te, "germ") // not in specificName
	if got != "Town Phase" {
		t.Errorf("got %q", got)
	}
}
```

- [ ] **Step 4.4: Run tests**

Run: `go test ./internal/i18n/ -v`
Expected: PASS.

- [ ] **Step 4.5: Run full suite to confirm no regressions**

Run: `go test ./...`
Expected: PASS / SKIP.

- [ ] **Step 4.6: Commit**

```bash
git add internal/i18n/
git commit -m "i18n: per-mod affects suffix; TechDisplayName(specificName)"
```

---

## Task 5 — render: rewrite `renderOverview` for epic-2 sections

**Closes:** gaps 2 (rendering), 3 (overview footer link), 5 (specificName lookup), 6 (affects display in overview), 20 (Identity), 21 (CivSpecific structures+technologies), 22 (Heroes block).

**Files:**
- Modify: `internal/render/report.go` (Generator field, Generate signature, helpers)
- Rewrite: `internal/render/overview.go`
- Create: `internal/render/overview_test.go`

- [ ] **Step 5.1: Add `IncludeHistory` field to `Generator`**

Edit `internal/render/report.go`. Replace the `Generator` struct (lines 14-18) with:

```go
type Generator struct {
	Layout         paths.Layout
	Resolver       *tmpl.Resolver
	Catalog        *tech.Catalog
	IncludeHistory bool
}
```

`NewGenerator` keeps its current signature (caller sets `IncludeHistory` separately after construction).

- [ ] **Step 5.2: Update `Generator.Generate` to load Player template + TeamBonus and pass to `renderOverview`**

In `internal/render/report.go`, replace the body of `Generate` (lines 36-64) with:

```go
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

	// New in epic 2: Player template (Identity), team bonus aura.
	// LoadPlayerTemplate failure is fatal — every civ has a Player template.
	player, err := civdata.LoadPlayerTemplate(g.Layout, civInfo.Code, g.Resolver)
	if err != nil {
		return Output{}, err
	}
	// LoadTeamBonus may legitimately be missing for civs without one in
	// future R-versions, but in R28 every civ has one. Treat ENOENT as
	// "no team bonus" rather than failing.
	teamBonus, err := aura.LoadTeamBonus(g.Layout, civInfo.Code)
	if err != nil && !os.IsNotExist(err) {
		return Output{}, err
	}

	return Output{
		Overview:  g.renderOverview(civInfo, civ, player, teamBonus, bonuses, notciv, units, buildings, heroAuras),
		Structree: g.renderStructree(civInfo.Code, buildings, units, heroAuras, catafalqueAuras),
	}, nil
}
```

Add `"os"` to the import block at the top of `report.go`.

- [ ] **Step 5.3: Rewrite `internal/render/overview.go`**

Replace the entire file with:

```go
package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/aura"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/i18n"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
)

// renderOverview returns the markdown body of <civ>_overview.md (without
// the skeleton wrapper). All top-level sections are rendered at "## "
// level for symmetry with structree.
//
// Section order (per spec 2026-04-27):
//  1. Идентичность
//  2. Историческая справка        (only if g.IncludeHistory and non-empty)
//  3. Герои
//  4. Уникальные строения
//  5. Уникальные технологии
//  6. Цивилизационные бонусы
//  7. Footer-ссылка на common.md (auto-research effects)
//  8. Командный бонус
//  9. Технологии, недоступные …
func (g *Generator) renderOverview(
	info civdata.CivCode,
	civ *civdata.Civ,
	player *civdata.PlayerTemplate,
	teamBonus *aura.Aura,
	bonuses, notciv []*tech.Technology,
	units, buildings []civdata.Entity,
	heroAuras []*aura.Aura,
) string {
	var sb strings.Builder
	g.overviewIdentity(&sb, info, civ, player)
	if g.IncludeHistory && player != nil && player.History != "" {
		g.overviewHistory(&sb, player)
	}
	g.overviewHeroes(&sb, units, heroAuras)
	g.overviewCivSpecificStructures(&sb, buildings)
	g.overviewSpecificTechnologies(&sb, info.Code, bonuses)
	g.overviewCivBonuses(&sb, info.Code, civ, bonuses)
	g.overviewGlobalAutoResearchFooter(&sb)
	g.overviewTeamBonus(&sb, civ, teamBonus)
	g.overviewNotCiv(&sb, info, notciv)
	return sb.String()
}

// 1. Идентичность

func (g *Generator) overviewIdentity(sb *strings.Builder,
	info civdata.CivCode, civ *civdata.Civ, player *civdata.PlayerTemplate) {

	fmt.Fprintln(sb, "## Идентичность")
	fmt.Fprintln(sb)
	fmt.Fprintf(sb, "- **Код:** `%s`\n", civ.Code)
	fmt.Fprintf(sb, "- **Культура:** %s\n", civ.Culture())
	if player != nil && player.GenericName != "" {
		fmt.Fprintf(sb, "- **Имя в данных:** %s\n", player.GenericName)
	}
	if player != nil && player.IconPath != "" {
		fmt.Fprintf(sb, "- **Эмблема:** `%s`\n", player.IconPath)
	}
	if se := formatStartEntities(civ.StartEntities); se != "" {
		fmt.Fprintf(sb, "- **Стартовые юниты:** %s\n", se)
	}
	fmt.Fprintln(sb)
}

// 2. Историческая справка (опционально)

func (g *Generator) overviewHistory(sb *strings.Builder, player *civdata.PlayerTemplate) {
	fmt.Fprintln(sb, "## Историческая справка")
	fmt.Fprintln(sb)
	for _, line := range strings.Split(player.History, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			fmt.Fprintln(sb, ">")
		} else {
			fmt.Fprintf(sb, "> %s\n", line)
		}
	}
	fmt.Fprintln(sb)
}

// 3. Герои

func (g *Generator) overviewHeroes(sb *strings.Builder,
	units []civdata.Entity, heroAuras []*aura.Aura) {

	heroes := []civdata.Entity{}
	for _, u := range units {
		if civdata.IsHero(u) {
			heroes = append(heroes, u)
		}
	}
	sort.Slice(heroes, func(i, j int) bool { return heroes[i].Basename() < heroes[j].Basename() })

	fmt.Fprintln(sb, "## Герои")
	fmt.Fprintln(sb)
	if len(heroes) == 0 {
		fmt.Fprintln(sb, "*У цивы нет уникальных героев.*")
		fmt.Fprintln(sb)
		return
	}
	for _, h := range heroes {
		name := FormatGenericName(h.Element)
		if name == "" {
			name = h.Basename()
		}
		classes := h.Element.GetTokens("Identity/VisibleClasses")
		classBadge := ""
		if len(classes) > 0 {
			classBadge = " — " + strings.Join(classes, ", ")
		}
		desc := pickHeroAuraDescription(h, heroAuras)
		if desc != "" {
			fmt.Fprintf(sb, "- **%s**%s. %s\n", name, classBadge, desc)
		} else {
			fmt.Fprintf(sb, "- **%s**%s\n", name, classBadge)
		}
	}
	fmt.Fprintln(sb)
}

// pickHeroAuraDescription finds the first matching aura for a hero and
// returns its auraDescription (or, falling back, the first modification
// rendered as a one-liner).
func pickHeroAuraDescription(h civdata.Entity, heroAuras []*aura.Aura) string {
	heroName := strings.TrimPrefix(h.Basename(), "hero_")
	auraTokens := h.Element.GetTokens("Auras")
	for _, tok := range auraTokens {
		base := strings.TrimPrefix(tok, "units/heroes/")
		for _, a := range heroAuras {
			if a.Name == base {
				return firstAuraDescription(a)
			}
		}
	}
	for _, a := range heroAuras {
		if strings.Contains(a.Name, heroName) {
			return firstAuraDescription(a)
		}
	}
	return ""
}

func firstAuraDescription(a *aura.Aura) string {
	if a.AuraDescription != "" {
		return a.AuraDescription
	}
	if len(a.Modifications) > 0 {
		return i18n.DescribeModification(a.Modifications[0])
	}
	return ""
}

// 4. Уникальные строения

func (g *Generator) overviewCivSpecificStructures(sb *strings.Builder,
	buildings []civdata.Entity) {

	specials := []civdata.Entity{}
	for _, b := range buildings {
		classes := b.Element.GetTokens("Identity/Classes")
		hasCivSpecific := false
		hasStructure := false
		for _, c := range classes {
			switch c {
			case "CivSpecific":
				hasCivSpecific = true
			case "Structure":
				hasStructure = true
			}
		}
		if hasCivSpecific && hasStructure {
			specials = append(specials, b)
		}
	}
	sort.Slice(specials, func(i, j int) bool { return specials[i].Basename() < specials[j].Basename() })

	fmt.Fprintln(sb, "## Уникальные строения")
	fmt.Fprintln(sb)
	if len(specials) == 0 {
		fmt.Fprintln(sb, "*У цивы нет уникальных строений.*")
		fmt.Fprintln(sb)
		return
	}
	for _, b := range specials {
		name := FormatGenericName(b.Element)
		if name == "" {
			name = b.Basename()
		}
		tooltip := b.Element.GetText("Identity/Tooltip")
		if tooltip != "" {
			fmt.Fprintf(sb, "- **%s** — %s\n", name, tooltip)
		} else {
			fmt.Fprintf(sb, "- **%s**\n", name)
		}
	}
	fmt.Fprintln(sb)
}

// 5. Уникальные технологии (краткий список)

func (g *Generator) overviewSpecificTechnologies(sb *strings.Builder,
	civCode string, bonuses []*tech.Technology) {

	fmt.Fprintln(sb, "## Уникальные технологии")
	fmt.Fprintln(sb)
	if len(bonuses) == 0 {
		fmt.Fprintln(sb, "*У цивы нет уникальных технологий.*")
		fmt.Fprintln(sb)
		return
	}
	sorted := append([]*tech.Technology(nil), bonuses...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	for _, t := range sorted {
		name := i18n.TechDisplayName(t, civCode)
		short := t.Tooltip
		if short == "" {
			short = t.Description
		}
		if short != "" {
			fmt.Fprintf(sb, "- **%s** — %s\n", name, short)
		} else {
			fmt.Fprintf(sb, "- **%s**\n", name)
		}
	}
	fmt.Fprintln(sb)
}

// 6. Цивилизационные бонусы (расширенная таблица)

func (g *Generator) overviewCivBonuses(sb *strings.Builder,
	civCode string, civ *civdata.Civ, bonuses []*tech.Technology) {

	fmt.Fprintln(sb, "## Цивилизационные бонусы")
	fmt.Fprintln(sb)
	if len(civ.CivBonuses) == 0 && len(bonuses) == 0 {
		fmt.Fprintln(sb, "*Особых цивилизационных бонусов не зафиксировано.*")
		fmt.Fprintln(sb)
		return
	}
	fmt.Fprintln(sb, "| Бонус | Источник | Требования | Эффект |")
	fmt.Fprintln(sb, "|-------|----------|------------|--------|")
	for _, b := range civ.CivBonuses {
		fmt.Fprintf(sb, "| %s | civ JSON | — | %s |\n",
			escapeTable(b.Name), escapeTable(b.Description))
	}
	for _, t := range bonuses {
		name := i18n.TechDisplayName(t, civCode)
		auto := ""
		if t.AutoResearch {
			auto = " (авто)"
		}
		req := t.RequirementsTooltip
		if req == "" {
			req = tech.DescribeRequirements(t.Requirements)
		}
		if req == "" {
			req = "—"
		}
		eff := t.Tooltip
		if eff == "" {
			eff = i18n.DescribeModifications(t.Modifications)
		}
		fmt.Fprintf(sb, "| %s%s | civbonuses/%s | %s | %s |\n",
			escapeTable(name), auto, t.Name, escapeTable(req), escapeTable(eff))
	}
	fmt.Fprintln(sb)
}

// 7. Footer-ссылка на common.md

func (g *Generator) overviewGlobalAutoResearchFooter(sb *strings.Builder) {
	fmt.Fprintln(sb, "> Глобальные авто-эффекты при повышении ранга применяются ко всем")
	fmt.Fprintln(sb, "> цивам — см. [common.md#модификаторы-advanced](common.md#модификаторы-advanced)")
	fmt.Fprintln(sb, "> и [#модификаторы-elite](common.md#модификаторы-elite).")
	fmt.Fprintln(sb)
}

// 8. Командный бонус

func (g *Generator) overviewTeamBonus(sb *strings.Builder,
	civ *civdata.Civ, teamBonus *aura.Aura) {

	fmt.Fprintln(sb, "## Командный бонус")
	fmt.Fprintln(sb)
	jsonName := ""
	jsonDesc := ""
	if len(civ.TeamBonuses) > 0 {
		jsonName = civ.TeamBonuses[0].Name
		jsonDesc = civ.TeamBonuses[0].Description
	}
	if teamBonus != nil {
		title := jsonName
		if teamBonus.AuraName != "" {
			title = teamBonus.AuraName
		}
		desc := jsonDesc
		if teamBonus.AuraDescription != "" {
			desc = teamBonus.AuraDescription
		}
		if title != "" {
			fmt.Fprintf(sb, "**%s.** %s\n\n", title, desc)
		} else if desc != "" {
			fmt.Fprintf(sb, "%s\n\n", desc)
		}
		if teamBonus.Type != "" {
			fmt.Fprintf(sb, "- Тип ауры: `%s`\n", teamBonus.Type)
		}
		if affects := teamBonus.AffectsHumanReadable(); len(affects) > 0 {
			fmt.Fprintf(sb, "- Цель: `%s`\n", strings.Join(affects, ", "))
		}
		if len(teamBonus.AffectedPlayers) > 0 {
			fmt.Fprintf(sb, "- Игроки: `%s`\n", strings.Join(teamBonus.AffectedPlayers, ", "))
		}
		fmt.Fprintln(sb)
		if len(teamBonus.Modifications) > 0 {
			fmt.Fprintln(sb, "| Цель | Эффект |")
			fmt.Fprintln(sb, "|------|--------|")
			for _, m := range teamBonus.Modifications {
				fmt.Fprintf(sb, "| %s | %s |\n",
					escapeTable(m.Value),
					escapeTable(i18n.DescribeModification(m)))
			}
			fmt.Fprintln(sb)
		}
	} else if jsonDesc != "" {
		// Fallback: civ.json string only.
		if jsonName != "" {
			fmt.Fprintf(sb, "**%s.** %s\n\n", jsonName, jsonDesc)
		} else {
			fmt.Fprintf(sb, "%s\n\n", jsonDesc)
		}
	} else {
		fmt.Fprintln(sb, "*У цивы нет командного бонуса.*")
		fmt.Fprintln(sb)
	}
}

// 9. Технологии, недоступные …

func (g *Generator) overviewNotCiv(sb *strings.Builder,
	info civdata.CivCode, notciv []*tech.Technology) {

	fmt.Fprintf(sb, "## Технологии, недоступные %s\n\n", info.NameRU)
	if len(notciv) == 0 {
		fmt.Fprintln(sb, "Явных запретов через `notciv` для этой цивилизации не найдено.")
		fmt.Fprintln(sb)
		return
	}
	sort.Slice(notciv, func(i, j int) bool { return notciv[i].Name < notciv[j].Name })
	for _, t := range notciv {
		name := i18n.TechDisplayName(t, info.Code)
		tip := t.Tooltip
		if tip == "" {
			tip = i18n.DescribeModifications(t.Modifications)
		}
		if tip != "" {
			fmt.Fprintf(sb, "- **%s** — %s\n", name, tip)
		} else {
			fmt.Fprintf(sb, "- **%s**\n", name)
		}
	}
	fmt.Fprintln(sb)
}
```

- [ ] **Step 5.4: Build to confirm compiles**

Run: `go build ./...`
Expected: success.

- [ ] **Step 5.5: Write `internal/render/overview_test.go`**

```go
package render

import (
	"strings"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

func TestOverview_Spart_Sections(t *testing.T) {
	skipIfNoGamedata(t)

	layout := paths.Layout{Root: gamedataRoot()}
	resolver := newResolver(t)
	g := NewGenerator(layout, resolver)

	info, ok := civdata.ResolveCivInput("spart")
	if !ok {
		t.Fatal("spart not resolved")
	}
	out, err := g.Generate(info)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	mustHaveAll(t, out.Overview, []string{
		"## Идентичность",
		"## Герои",
		"## Уникальные строения",
		"## Уникальные технологии",
		"## Цивилизационные бонусы",
		"## Командный бонус",
		"## Технологии, недоступные Спартанцы",
		"common.md#модификаторы-advanced",
		"Peloponnesian League",
		"`spart`",
	})
	mustNotHave(t, out.Overview, []string{
		"## Историческая справка", // disabled by default
		"## Общая информация о цивилизации", // epic-1 wording removed
	})
}

func TestOverview_Spart_HistoryFlag(t *testing.T) {
	skipIfNoGamedata(t)

	layout := paths.Layout{Root: gamedataRoot()}
	resolver := newResolver(t)
	g := NewGenerator(layout, resolver)
	g.IncludeHistory = true

	info, _ := civdata.ResolveCivInput("spart")
	out, err := g.Generate(info)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(out.Overview, "## Историческая справка") {
		t.Errorf("history section missing when IncludeHistory=true")
	}
	if !strings.Contains(out.Overview, "Sparta was a prominent city-state") {
		t.Errorf("history body missing")
	}
}

func TestOverview_AffectsSuffix_Visible(t *testing.T) {
	// At least one civ-bonus or team-bonus row should render the (только …)
	// suffix added in epic 2. We pick spartans because their Peloponnesian
	// League team bonus has affects:["Hero"] in JSON.
	skipIfNoGamedata(t)
	layout := paths.Layout{Root: gamedataRoot()}
	g := NewGenerator(layout, newResolver(t))
	info, _ := civdata.ResolveCivInput("spart")
	out, err := g.Generate(info)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// The team bonus modifications themselves don't have per-mod affects,
	// but the team-bonus aura's top-level Affects field should reach the
	// "Цель" bullet ("Цель: `Hero`"). Verify that.
	if !strings.Contains(out.Overview, "Цель: `Hero`") {
		t.Errorf("expected `Цель: \\`Hero\\`` in spart team bonus block")
	}
}

func mustHaveAll(t *testing.T, body string, needles []string) {
	t.Helper()
	for _, n := range needles {
		if !strings.Contains(body, n) {
			t.Errorf("body missing %q", n)
		}
	}
}

func mustNotHave(t *testing.T, body string, forbidden []string) {
	t.Helper()
	for _, n := range forbidden {
		if strings.Contains(body, n) {
			t.Errorf("body unexpectedly contains %q", n)
		}
	}
}
```

- [ ] **Step 5.6: Run overview tests**

Run: `go test ./internal/render/ -run 'TestOverview' -v`
Expected: PASS.

- [ ] **Step 5.7: Patch `golden_test.go` to stay green during the transition**

The existing `TestGoldenGermStructure` asserts on epic-1 strings (`## Общая информация о цивилизации`, `### Цивилизационные бонусы`) that no longer appear after the renderOverview rewrite. To keep `go test ./...` green at this commit, do a minimal targeted edit (full rewrite happens in Task 7.2):

In `internal/render/golden_test.go`, replace the `overviewMust` slice with a relaxed list that matches the new structure but keeps coverage modest:

```go
	overviewMust := []string{
		"## Идентичность",
		"- **Код:** `germ`",
		"## Цивилизационные бонусы",
	}
```

Also remove the line-count threshold check that says `< 25` — keep the file's structural asserts only. Replace lines 30–36 (the comment and overviewLines block) with:

```go
	// Overview line-count threshold tightened to 50 in Task 7.2.
	overviewLines := strings.Count(out.Overview, "\n") + 1
	_ = overviewLines
```

Leave the `commonBody contains "TODO"` assertion as-is for now — common.md is still the epic-1 placeholder until Task 6.

- [ ] **Step 5.8: Run full suite**

Run: `go test ./...`
Expected: PASS / SKIP across the board.

- [ ] **Step 5.9: Commit**

```bash
git add internal/render/report.go internal/render/overview.go internal/render/overview_test.go internal/render/golden_test.go
git commit -m "render: rewrite renderOverview with epic-2 sections"
```

---

## Task 6 — render: populate `common.md`

**Closes:** finishes gap 3 (autoResearch text); adds shared-reference content.

**Files:**
- Create: `internal/render/common_data.go` (loaders for damage_types/resources/status_effects)
- Rewrite: `internal/render/common.go`
- Create: `internal/render/common_test.go`
- Modify: `internal/paths/paths.go` (DamageTypes/Resources/StatusEffects accessors)

- [ ] **Step 6.1: Add layout accessors for `data/damage_types/`, `data/resources/`, `data/status_effects/`**

Edit `internal/paths/paths.go`. After the existing accessors, append:

```go
func (l Layout) DamageTypes() string {
	return filepath.Join(l.Root, "simulation", "data", "damage_types")
}

func (l Layout) Resources() string {
	return filepath.Join(l.Root, "simulation", "data", "resources")
}

func (l Layout) StatusEffects() string {
	return filepath.Join(l.Root, "simulation", "data", "status_effects")
}
```

- [ ] **Step 6.2: Create `internal/render/common_data.go`**

```go
package render

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type damageType struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}

type resourceType struct {
	Code        string            `json:"code"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Order       int               `json:"order"`
	Subtypes    map[string]string `json:"subtypes"`
}

type statusEffect struct {
	Code            string `json:"code"`
	StatusName      string `json:"statusName"`
	ApplierTooltip  string `json:"applierTooltip"`
	ReceiverTooltip string `json:"receiverTooltip"`
}

func loadDamageTypes(dir string) ([]damageType, error) {
	out := []damageType{}
	if err := loadJSONDir(dir, func(raw []byte, _ string) error {
		var d damageType
		if err := json.Unmarshal(raw, &d); err != nil {
			return err
		}
		out = append(out, d)
		return nil
	}); err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].Code < out[j].Code
	})
	return out, nil
}

func loadResources(dir string) ([]resourceType, error) {
	out := []resourceType{}
	if err := loadJSONDir(dir, func(raw []byte, _ string) error {
		var r resourceType
		if err := json.Unmarshal(raw, &r); err != nil {
			return err
		}
		out = append(out, r)
		return nil
	}); err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].Code < out[j].Code
	})
	return out, nil
}

func loadStatusEffects(dir string) ([]statusEffect, error) {
	out := []statusEffect{}
	if err := loadJSONDir(dir, func(raw []byte, _ string) error {
		var s statusEffect
		if err := json.Unmarshal(raw, &s); err != nil {
			return err
		}
		out = append(out, s)
		return nil
	}); err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out, nil
}

func loadJSONDir(dir string, fn func(raw []byte, path string) error) error {
	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return err
	}
	sort.Strings(matches)
	for _, p := range matches {
		raw, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("read %s: %w", p, err)
		}
		if err := fn(raw, p); err != nil {
			return fmt.Errorf("parse %s: %w", p, err)
		}
	}
	return nil
}

// resourceSubtypeKeys returns subtypes ordered alphabetically for
// deterministic rendering.
func resourceSubtypeKeys(r resourceType) []string {
	keys := make([]string, 0, len(r.Subtypes))
	for k := range r.Subtypes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// joinSubtypes formats subtypes as "fish, fruit, grain, meat".
func joinSubtypes(r resourceType) string {
	keys := resourceSubtypeKeys(r)
	return strings.Join(keys, ", ")
}
```

- [ ] **Step 6.3: Rewrite `internal/render/common.go`**

```go
package render

import (
	"fmt"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/i18n"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
)

// renderCommonBody returns the body of common.md (without the skeleton
// wrapper). Sections, in order:
//   1. Модификаторы Advanced
//   2. Модификаторы Elite
//   3. Прочие глобальные авто-эффекты
//   4. Типы урона
//   5. Типы ресурсов
//   6. Статус-эффекты
//
// All section titles render at "## " level so anchors generated by
// GitHub-style markdown work for cross-file links from civ overviews.
func (g *Generator) renderCommonBody() (string, error) {
	var sb strings.Builder
	if err := g.commonAdvanced(&sb); err != nil {
		return "", err
	}
	if err := g.commonElite(&sb); err != nil {
		return "", err
	}
	if err := g.commonOtherAuto(&sb); err != nil {
		return "", err
	}
	if err := g.commonDamageTypes(&sb); err != nil {
		return "", err
	}
	if err := g.commonResources(&sb); err != nil {
		return "", err
	}
	if err := g.commonStatusEffects(&sb); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (g *Generator) commonAdvanced(sb *strings.Builder) error {
	t, err := g.Catalog.ByName("unit_advanced")
	if err != nil {
		return err
	}
	fmt.Fprintln(sb, "## Модификаторы Advanced")
	fmt.Fprintln(sb)
	if t.Tooltip != "" {
		fmt.Fprintf(sb, "> %s\n\n", t.Tooltip)
	}
	commonModTable(sb, t.Modifications)
	return nil
}

func (g *Generator) commonElite(sb *strings.Builder) error {
	t, err := g.Catalog.ByName("unit_elite")
	if err != nil {
		return err
	}
	fmt.Fprintln(sb, "## Модификаторы Elite")
	fmt.Fprintln(sb)
	if t.Tooltip != "" {
		fmt.Fprintf(sb, "> %s\n\n", t.Tooltip)
	}
	commonModTable(sb, t.Modifications)
	return nil
}

func (g *Generator) commonOtherAuto(sb *strings.Builder) error {
	all, err := g.Catalog.GlobalAutoResearch()
	if err != nil {
		return err
	}
	others := []*tech.Technology{}
	for _, t := range all {
		if t.Name == "unit_advanced" || t.Name == "unit_elite" {
			continue
		}
		others = append(others, t)
	}
	fmt.Fprintln(sb, "## Прочие глобальные авто-эффекты")
	fmt.Fprintln(sb)
	if len(others) == 0 {
		fmt.Fprintln(sb, "*В текущих данных только Advanced/Elite.*")
		fmt.Fprintln(sb)
		return nil
	}
	fmt.Fprintln(sb, "Сюда попадают autoResearch-техи без `requirements.civ`,")
	fmt.Fprintln(sb, "не относящиеся к ранговому апгрейду:")
	fmt.Fprintln(sb)
	for _, t := range others {
		name := t.GenericName
		if name == "" {
			name = t.Name
		}
		fmt.Fprintf(sb, "### %s (`%s`)\n\n", name, t.Name)
		if t.Tooltip != "" {
			fmt.Fprintf(sb, "> %s\n\n", t.Tooltip)
		} else if t.Description != "" {
			fmt.Fprintf(sb, "> %s\n\n", t.Description)
		}
		if len(t.Modifications) > 0 {
			commonModTable(sb, t.Modifications)
		}
	}
	return nil
}

func commonModTable(sb *strings.Builder, mods []tech.Modification) {
	if len(mods) == 0 {
		fmt.Fprintln(sb, "*Нет модификаций.*")
		fmt.Fprintln(sb)
		return
	}
	fmt.Fprintln(sb, "| Цель | Эффект |")
	fmt.Fprintln(sb, "|------|--------|")
	for _, m := range mods {
		fmt.Fprintf(sb, "| %s | %s |\n",
			escapeTable(m.Value),
			escapeTable(i18n.DescribeModification(m)))
	}
	fmt.Fprintln(sb)
}

func (g *Generator) commonDamageTypes(sb *strings.Builder) error {
	dts, err := loadDamageTypes(g.Layout.DamageTypes())
	if err != nil {
		return err
	}
	fmt.Fprintln(sb, "## Типы урона")
	fmt.Fprintln(sb)
	fmt.Fprintln(sb, "| Код | Имя | Описание |")
	fmt.Fprintln(sb, "|-----|-----|----------|")
	for _, d := range dts {
		fmt.Fprintf(sb, "| `%s` | %s | %s |\n",
			d.Code, escapeTable(d.Name), escapeTable(d.Description))
	}
	fmt.Fprintln(sb)
	return nil
}

func (g *Generator) commonResources(sb *strings.Builder) error {
	rs, err := loadResources(g.Layout.Resources())
	if err != nil {
		return err
	}
	fmt.Fprintln(sb, "## Типы ресурсов")
	fmt.Fprintln(sb)
	fmt.Fprintln(sb, "| Код | Имя | Описание |")
	fmt.Fprintln(sb, "|-----|-----|----------|")
	for _, r := range rs {
		fmt.Fprintf(sb, "| `%s` | %s | %s |\n",
			r.Code, escapeTable(r.Name), escapeTable(r.Description))
	}
	fmt.Fprintln(sb)
	any := false
	for _, r := range rs {
		if len(r.Subtypes) > 0 {
			any = true
			break
		}
	}
	if any {
		fmt.Fprintln(sb, "### Подтипы ресурсов")
		fmt.Fprintln(sb)
		for _, r := range rs {
			if len(r.Subtypes) == 0 {
				continue
			}
			fmt.Fprintf(sb, "- **%s**: %s\n", r.Code, joinSubtypes(r))
		}
		fmt.Fprintln(sb)
	}
	return nil
}

func (g *Generator) commonStatusEffects(sb *strings.Builder) error {
	es, err := loadStatusEffects(g.Layout.StatusEffects())
	if err != nil {
		return err
	}
	fmt.Fprintln(sb, "## Статус-эффекты")
	fmt.Fprintln(sb)
	fmt.Fprintln(sb, "| Код | Имя | На применяющего | На пострадавшего |")
	fmt.Fprintln(sb, "|-----|-----|-----------------|------------------|")
	for _, e := range es {
		fmt.Fprintf(sb, "| `%s` | %s | %s | %s |\n",
			e.Code, escapeTable(e.StatusName),
			escapeTable(e.ApplierTooltip), escapeTable(e.ReceiverTooltip))
	}
	fmt.Fprintln(sb)
	return nil
}
```

- [ ] **Step 6.4: Update `RenderCommon` in `report.go`**

In `internal/render/report.go`, replace:

```go
// RenderCommon returns the body of the shared common.md.
// In epic 1 this is a placeholder — populated in epic 2.
func (g *Generator) RenderCommon() (string, error) {
	return renderCommonBody(), nil
}
```

with:

```go
// RenderCommon returns the body of the shared common.md (without the
// skeleton wrapper). See internal/render/common.go for section layout.
func (g *Generator) RenderCommon() (string, error) {
	return g.renderCommonBody()
}
```

- [ ] **Step 6.5: Build to confirm compiles**

Run: `go build ./...`
Expected: success.

- [ ] **Step 6.6: Write `internal/render/common_test.go`**

```go
package render

import (
	"strings"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

func TestCommon_AllSections(t *testing.T) {
	skipIfNoGamedata(t)

	layout := paths.Layout{Root: gamedataRoot()}
	g := NewGenerator(layout, newResolver(t))
	body, err := g.RenderCommon()
	if err != nil {
		t.Fatalf("RenderCommon: %v", err)
	}
	mustHaveAll(t, body, []string{
		"## Модификаторы Advanced",
		"## Модификаторы Elite",
		"## Прочие глобальные авто-эффекты",
		"## Типы урона",
		"## Типы ресурсов",
		"## Статус-эффекты",
	})
}

func TestCommon_AdvancedContent(t *testing.T) {
	skipIfNoGamedata(t)

	layout := paths.Layout{Root: gamedataRoot()}
	g := NewGenerator(layout, newResolver(t))
	body, err := g.RenderCommon()
	if err != nil {
		t.Fatalf("RenderCommon: %v", err)
	}
	// Advanced gives +25% HP and per-mod (только Melee) suffix.
	if !strings.Contains(body, "Health/Max") {
		t.Errorf("missing Health/Max in advanced section")
	}
	if !strings.Contains(body, "(только Melee)") {
		t.Errorf("missing per-mod affects suffix")
	}
}

func TestCommon_DamageTypes(t *testing.T) {
	skipIfNoGamedata(t)

	layout := paths.Layout{Root: gamedataRoot()}
	g := NewGenerator(layout, newResolver(t))
	body, err := g.RenderCommon()
	if err != nil {
		t.Fatalf("RenderCommon: %v", err)
	}
	for _, code := range []string{"Hack", "Pierce", "Crush", "Fire", "Poison"} {
		if !strings.Contains(body, "| `"+code+"` |") {
			t.Errorf("missing damage code %q row", code)
		}
	}
}

func TestCommon_Resources_WithSubtypes(t *testing.T) {
	skipIfNoGamedata(t)

	layout := paths.Layout{Root: gamedataRoot()}
	g := NewGenerator(layout, newResolver(t))
	body, err := g.RenderCommon()
	if err != nil {
		t.Fatalf("RenderCommon: %v", err)
	}
	if !strings.Contains(body, "### Подтипы ресурсов") {
		t.Errorf("missing subtypes header")
	}
	if !strings.Contains(body, "- **food**:") {
		t.Errorf("missing food subtypes line")
	}
}

func TestCommon_StatusEffects(t *testing.T) {
	skipIfNoGamedata(t)

	layout := paths.Layout{Root: gamedataRoot()}
	g := NewGenerator(layout, newResolver(t))
	body, err := g.RenderCommon()
	if err != nil {
		t.Fatalf("RenderCommon: %v", err)
	}
	if !strings.Contains(body, "Burning") || !strings.Contains(body, "Poisoned") {
		t.Errorf("missing burning/poisoned in status effects")
	}
}
```

- [ ] **Step 6.7: Run common tests**

Run: `go test ./internal/render/ -run 'TestCommon' -v`
Expected: PASS.

- [ ] **Step 6.8: Patch `golden_test.go` to drop the epic-1 `TODO` substring assertion**

After this commit `common.md` no longer contains «TODO». In `internal/render/golden_test.go` remove the block:

```go
	if !strings.Contains(commonBody, "TODO") {
		t.Errorf("common body should mention TODO placeholder in epic 1")
	}
```

Replace it with a minimal positive assertion that won't conflict with Task 7.2's full rewrite:

```go
	if !strings.Contains(commonBody, "## Модификаторы Advanced") {
		t.Errorf("common missing Advanced section")
	}
```

- [ ] **Step 6.9: Run full suite**

Run: `go test ./...`
Expected: PASS / SKIP across the board.

- [ ] **Step 6.10: Commit**

```bash
git add internal/render/common.go internal/render/common_data.go internal/render/common_test.go internal/render/report.go internal/render/golden_test.go internal/paths/paths.go
git commit -m "render: populate common.md (advanced/elite/auto/damage/resources/status)"
```

---

## Task 7 — cmd,config: bind `--include-history`; update docs and golden_test

**Closes:** wire `--include-history` flag end-to-end; sync README/CLAUDE.md; replace epic-1 golden_test asserts with epic-2.

**Files:**
- Modify: `cmd/civreport/main.go`
- Modify: `internal/render/golden_test.go`
- Modify: `README.md`
- Modify: `CLAUDE.md`

- [ ] **Step 7.1: Bind `--include-history` flag in `cmd/civreport/main.go`**

In `cmd/civreport/main.go` `main()`, locate the existing flag block. After the `--check` flag declaration, add:

```go
	var includeHistory bool
	flag.BoolVar(&includeHistory, "include-history", false, "include the civ history paragraph in <civ>_overview.md (off by default)")
```

After the `if outDirFlag != "" { cfg.OutDir = outDirFlag }` line, add:

```go
	if includeHistory {
		cfg.IncludeHistory = true
	}
```

After the `gen := render.NewGenerator(...)` line, add:

```go
	gen.IncludeHistory = cfg.IncludeHistory
```

The full mutation surface in `main` is therefore:
1. `flag.BoolVar(&includeHistory, "include-history", ...)`
2. `if includeHistory { cfg.IncludeHistory = true }`
3. `gen.IncludeHistory = cfg.IncludeHistory`

The Usage block should also list the new flag implicitly via `flag.PrintDefaults()` — no manual edit needed.

- [ ] **Step 7.2: Rewrite `internal/render/golden_test.go`**

Replace the entire file with:

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

	// After epic 2, overview body grows from ~30 lines (epic 1) to
	// ~80–150. Threshold raised to 50 — leaves headroom but still flags
	// regressions where major sections collapse to placeholders.
	overviewLines := strings.Count(out.Overview, "\n") + 1
	if overviewLines < 50 {
		t.Errorf("overview too short: %d lines (want >= 50)", overviewLines)
	}
	structreeLines := strings.Count(out.Structree, "\n") + 1
	if structreeLines < 100 {
		t.Errorf("structree too short: %d lines (want >= 100)", structreeLines)
	}

	overviewMust := []string{
		"## Идентичность",
		"- **Код:** `germ`",
		"## Герои",
		"## Уникальные строения",
		"## Уникальные технологии",
		"## Цивилизационные бонусы",
		"## Командный бонус",
		"## Технологии, недоступные Германцы",
		"common.md#модификаторы-advanced",
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
	commonMust := []string{
		"## Модификаторы Advanced",
		"## Модификаторы Elite",
		"## Прочие глобальные авто-эффекты",
		"## Типы урона",
		"## Типы ресурсов",
		"## Статус-эффекты",
	}
	for _, m := range commonMust {
		if !strings.Contains(commonBody, m) {
			t.Errorf("common missing %q", m)
		}
	}

	// Optional reference goldens at testdata/golden/germans_*.md.
	// Strict-byte-diff arrives in epic 4.
	wd, _ := os.Getwd()
	for _, f := range []string{"germans_overview.md", "germans_structree.md"} {
		path := filepath.Join(wd, "..", "..", "testdata", "golden", f)
		if _, err := os.Stat(path); err != nil {
			t.Logf("golden reference %s not present (ok pre-epic-4)", f)
			continue
		}
		if _, err := os.ReadFile(path); err != nil {
			t.Errorf("read golden %s: %v", f, err)
		}
	}
}
```

- [ ] **Step 7.3: Build and run full test suite**

Run: `go build ./... && go test ./...`
Expected: PASS / SKIP across the board. No more `TestGoldenGermStructure` failure.

- [ ] **Step 7.4: Smoke-run `civreport` against real data**

Run: `go run ./cmd/civreport --out-dir /tmp/civreport-epic2 spart`
Expected output:
```
OK spart → /tmp/civreport-epic2/{spartans_overview.md,spartans_structree.md} + common.md
```

Inspect: `head -40 /tmp/civreport-epic2/spartans_overview.md`
Expected to see header + skeleton info-block + `## Идентичность` + bullets.

Run with history flag: `go run ./cmd/civreport --include-history --out-dir /tmp/civreport-epic2-h spart`
Inspect: `grep -c '## Историческая справка' /tmp/civreport-epic2-h/spartans_overview.md`
Expected: `1`.

- [ ] **Step 7.5: Smoke-run `--check` and `--all`**

Run: `go run ./cmd/civreport --check`
Expected: 15 lines `OK <civ> (overview=…, structree=…)`, exit 0.

Run: `go run ./cmd/civreport --out-dir /tmp/civreport-epic2-all --all`
Expected: 15 `OK` lines + `OK common.md`, 30 cv-files + 1 common.md.

- [ ] **Step 7.6: Update `README.md`**

Open `README.md`. In the «CLI / Опции» section (or wherever flags are listed), add the new flag:

```markdown
- `--include-history` — добавляет блок «Историческая справка» в `<civ>_overview.md` из `Identity/History` Player-шаблона. По умолчанию выключено.
```

In the «Содержимое отчётов» (or equivalent) section, replace any epic-1
description with the new layout:

```markdown
### `<civ>_overview.md`

Восемь обязательных секций (Identity, Heroes, Уникальные строения,
Уникальные технологии, Цивилизационные бонусы, Командный бонус,
Технологии-недоступные) + опциональная Историческая справка.
Цив-бонусы содержат колонку «Требования» и per-mod суффикс
`(только …)`. Footer ссылается на `common.md` для глобальных
auto-research эффектов.

### `common.md`

Общая для всех 15 цив справка: модификаторы Advanced/Elite, прочие
глобальные авто-эффекты, типы урона, типы ресурсов, статус-эффекты.
Цив-overview ссылается на якоря этого файла.
```

(Если соответствующих заголовков ещё нет — добавь их в логичном месте README, не дублируя.)

- [ ] **Step 7.7: Update `CLAUDE.md`**

Open `CLAUDE.md`. The «Текущее» / «Целевая» split was an epic-1 artifact. After epic 2 they're closer to converged; do the following minimal edits:

1. In the file structure block of «Целевая структура» (around the `internal/render/` entry), update:
   - `skeleton_overview.tmpl` etc. — already moved to `skeleton/`-subdir, эта часть была актуальна; ничего не меняем.
   - `overview.go`/`structree.go`/`common.go` — пометь как реализованные после эпика 2 (можно убрать слово «NEW»).

2. В блоке «Опции» (`./bin/civreport --lang`/`--include-history`/`--include-icons`):
   - оставь `--include-history` без пометки «не реализовано» — после эпика 2 он работает.
   - пометь `--lang ru` и `--include-icons` как `(не реализовано до эпика 5)`.

3. В «Структура пакетов» → «Текущее»: добавь упоминание новых файлов
   `civdata/player.go`, `aura/teambonus.go`, `tech/auto.go`,
   `tech/requirements.go`, `i18n/tech_name.go`,
   `render/common_data.go` рядом с уже перечисленными.

4. В «Не покрыто» / «Известные ограничения» (если такая секция есть в
   CLAUDE.md или в `docs/current-state.md`) — вычеркни закрытые gaps:
   2, 3, 4 (частично — только overview), 5, 6, 20, 21, 22.

Если время поджимает, минимум: добавить в верх «Конвенции» строчку:
```markdown
- **После эпика 2**: overview содержит 8 секций + опц. History;
  common.md наполнен (Advanced/Elite/Auto/Damage/Resources/Status).
```

- [ ] **Step 7.8: Final full test run**

Run: `go test ./...`
Expected: PASS / SKIP.

- [ ] **Step 7.9: Final build**

Run: `go build ./...`
Expected: success, no warnings.

- [ ] **Step 7.10: Commit**

```bash
git add cmd/civreport/main.go internal/render/golden_test.go README.md CLAUDE.md
git commit -m "cmd,config,docs: bind --include-history; update golden_test and docs"
```

---

## Post-flight

- [ ] **Step 8.1: Push branch**

```bash
git push -u origin epic-2-overview-common
```

- [ ] **Step 8.2: Confirm 7 commits on the branch**

Run: `git log master..HEAD --oneline`
Expected: 7 lines, in order:
1. civdata: load Identity from special/players/<civ>.xml
2. aura: add LoadTeamBonus single-file loader
3. tech: per-mod Affects, RequirementsTooltip, GlobalAutoResearch, DescribeRequirements
4. i18n: per-mod affects suffix; TechDisplayName(specificName)
5. render: rewrite renderOverview with epic-2 sections
6. render: populate common.md (advanced/elite/auto/damage/resources/status)
7. cmd,config,docs: bind --include-history; update golden_test and docs

- [ ] **Step 8.3: Confirm DoD checklist from spec**

Per spec § «Definition of Done»:

1. ✅ 7 коммитов на ветке, каждый компилируется + тесты проходят (Step 8.2 + run `go test` at each commit if doubt).
2. ✅ `civreport spart` пишет 3 файла, overview содержит 8 секций (Step 7.4).
3. ✅ `--include-history` добавляет «Историческая справка» (Step 7.4).
4. ✅ `common.md` содержит 6 секций (Step 6.6 test).
5. ✅ `civreport --all` пишет 30 + 1 (Step 7.5).
6. ✅ `civreport --check` зелёный (Step 7.5).
7. ✅ `go test ./...` зелёный (Step 7.8).
8. ✅ README + CLAUDE.md актуализированы (Step 7.6, 7.7).
9. ✅ Civ Bonuses таблица содержит «Требования» + per-mod `(только …)` (Step 5.5 + 5.7).
10. ✅ Якорные ссылки в overview (Step 5.5 test).

If any are red, fix before opening PR.

- [ ] **Step 8.4: Open PR (optional, can be deferred)**

Once user requests it, use `gh pr create` per the global commit/PR conventions in CLAUDE.md.

---

## Self-Review Notes

**Spec coverage check:**

| Spec section | Implemented in |
|---|---|
| Identity loader (gap 20) | Task 1 |
| Team-bonus loader (gap 2) | Task 2, used in Task 5 |
| Per-mod Affects (gap 6) | Task 3 (struct), Task 4 (rendering) |
| RequirementsTooltip + DescribeRequirements (gap 4) | Task 3 (struct + helpers), Task 5 (used in Civ Bonuses table) |
| specificName lookup (gap 5) | Task 4 (TechDisplayName), Task 5 (Civ Bonuses + notciv) |
| GlobalAutoResearch (gap 3) | Task 3 (helper), Task 6 (rendered in common.md) |
| Heroes block in overview (gap 22) | Task 5 (`overviewHeroes`) |
| CivSpecific structures block (gap 21) | Task 5 (`overviewCivSpecificStructures`) |
| Specific Technologies block (gap 21) | Task 5 (`overviewSpecificTechnologies`) |
| Common.md content | Task 6 |
| `--include-history` CLI | Task 7 |
| Golden test update | Task 7 |
| README/CLAUDE.md updates | Task 7 |

**No-placeholder check:** Every code step contains complete code. Every shell command is exact. No "TBD" or "implement later".

**Type consistency:** `Generator.IncludeHistory bool` declared in Task 5.1, set in Task 7.1, read in `overview.go` (Task 5.3). `LoadPlayerTemplate` signature `(layout, civ, resolver)` consistent across Task 1 (definition), Task 5.2 (call). `LoadTeamBonus` signature `(layout, civ)` consistent across Task 2 and Task 5.2. `Modification.AffectsRaw` field used by `AffectsList()` in Task 3, called from `DescribeModification` in Task 4 — same field name throughout.
