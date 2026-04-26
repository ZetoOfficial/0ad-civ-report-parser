package render

import (
	"fmt"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

func (g *Generator) renderSummary(sb *strings.Builder, buildings []civdata.Entity) {
	fmt.Fprintln(sb, "## Приложение: Сводная таблица строимых зданий")
	fmt.Fprintln(sb)
	fmt.Fprintln(sb, "| Строение | Дерево | Камень | Металл | Еда | Время | Фаза |")
	fmt.Fprintln(sb, "|----------|--------|--------|--------|-----|-------|------|")
	for _, ph := range []civdata.Phase{civdata.PhaseVillage, civdata.PhaseTown, civdata.PhaseCity} {
		for _, b := range buildings {
			if civdata.BuildingPhase(b) != ph {
				continue
			}
			name := FormatGenericName(b.Element)
			if name == "" {
				name = b.Basename()
			}
			res := b.Element.Get("Cost/Resources")
			fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s | %s |\n",
				escapeTable(name),
				resOrDash(res, "wood"),
				resOrDash(res, "stone"),
				resOrDash(res, "metal"),
				resOrDash(res, "food"),
				FormatBuildTime(b.Element),
				ph.RU(),
			)
		}
	}
	fmt.Fprintln(sb)
}

func resOrDash(res *tmpl.Element, key string) string {
	if res == nil {
		return "—"
	}
	c := res.Child(key)
	if c == nil {
		return "—"
	}
	v, ok := tmpl.ParseInt(strings.TrimSpace(c.Text))
	if !ok {
		return "—"
	}
	if v == 0 {
		return "—"
	}
	return fmt.Sprintf("%d", v)
}
