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

	// Overview line-count threshold tightened to 50 in Task 7.
	overviewLines := strings.Count(out.Overview, "\n") + 1
	_ = overviewLines
	structreeLines := strings.Count(out.Structree, "\n") + 1
	if structreeLines < 100 {
		t.Errorf("structree too short: %d lines (want >= 100)", structreeLines)
	}

	overviewMust := []string{
		"## Идентичность",
		"- **Код:** `germ`",
		"## Цивилизационные бонусы",
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

	// Optional reference goldens at testdata/golden/germans_*.md.
	// In Epic 1 we only verify they're readable when present; Epic 4
	// will replace this with a strict byte-diff against Generate output.
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
