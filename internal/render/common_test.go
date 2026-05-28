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
	// Per-status sub-block headings.
	mustHaveAll(t, body, []string{
		"### Poisoned",
		"### Burning",
	})
	// Quoted applier/receiver tooltips from actual poisoned.json.
	mustHaveAll(t, body, []string{
		"> **Применяющему:** This unit causes poison damage.",
		"> **Пострадавшему:** This unit is poisoned.",
	})
}
