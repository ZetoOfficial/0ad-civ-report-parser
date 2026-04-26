package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

func TestGoldenGermSmoke(t *testing.T) {
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
	body := out.Overview + "\n" + out.Structree
	lines := strings.Count(body, "\n") + 1
	if lines < 700 {
		t.Errorf("germ report too short: %d lines (want ≥ 700)", lines)
	}

	must := []string{
		"## VILLAGE PHASE",
		"## TOWN PHASE",
		"## CITY PHASE",
		"## Приложение: Детальная информация по типам юнитов",
		"## Приложение: Сводная таблица строимых зданий",
		"### Цивилизационные бонусы",
	}
	for _, m := range must {
		if !strings.Contains(body, m) {
			t.Errorf("germ report missing required section: %q", m)
		}
	}

	wd, _ := os.Getwd()
	goldenPath := filepath.Join(wd, "..", "..", "testdata", "golden", "germans_buildings_report.md")
	if _, err := os.Stat(goldenPath); err != nil {
		t.Skipf("golden file not found at %s", goldenPath)
	}

	gold, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Skipf("read golden: %v", err)
	}
	checks := []string{
		"## VILLAGE PHASE",
		"## TOWN PHASE",
		"## CITY PHASE",
	}
	for _, c := range checks {
		if !strings.Contains(string(gold), c) {
			t.Errorf("golden file missing %q (golden may be outdated)", c)
		}
	}
}
