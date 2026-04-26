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
