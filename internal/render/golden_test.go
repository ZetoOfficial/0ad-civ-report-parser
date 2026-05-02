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
		"### Стены",
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
