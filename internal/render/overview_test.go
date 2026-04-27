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
		"## Историческая справка",
		"## Общая информация о цивилизации",
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

func TestOverview_AffectsRendering(t *testing.T) {
	skipIfNoGamedata(t)
	layout := paths.Layout{Root: gamedataRoot()}
	g := NewGenerator(layout, newResolver(t))
	info, _ := civdata.ResolveCivInput("spart")
	out, err := g.Generate(info)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	// The team-bonus aura's top-level Affects field surfaces as
	// "Цель: `Hero`" in the team bonus block.
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
