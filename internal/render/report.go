package render

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/aura"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/i18n"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

type Generator struct {
	Layout   paths.Layout
	Resolver *tmpl.Resolver
	Catalog  *tech.Catalog
}

func NewGenerator(layout paths.Layout, resolver *tmpl.Resolver) *Generator {
	return &Generator{
		Layout:   layout,
		Resolver: resolver,
		Catalog:  tech.NewCatalog(layout.Technologies()),
	}
}

func (g *Generator) Generate(civInfo civdata.CivCode) (string, error) {
	civ, err := civdata.LoadCiv(g.Layout.CivJSON(civInfo.Code))
	if err != nil {
		return "", err
	}
	buildings, err := civdata.Buildings(g.Layout.StructuresOf(civInfo.Code), civInfo.Code, g.Resolver)
	if err != nil {
		return "", err
	}
	units, err := civdata.Units(g.Layout.UnitsOf(civInfo.Code), civInfo.Code, g.Resolver)
	if err != nil {
		return "", err
	}
	bonuses, err := g.Catalog.AllCivBonuses(civInfo.Code)
	if err != nil {
		return "", err
	}
	notciv, err := g.Catalog.AllNotCiv(civInfo.Code)
	if err != nil {
		return "", err
	}

	heroAuras, _ := aura.ListInDir(g.Layout.HeroAuras(), civInfo.Code+"_hero_")
	catafalqueAuras, _ := aura.ListInDir(g.Layout.CatafalqueAuras(), civInfo.Code+"_")

	var sb strings.Builder
	g.renderHeader(&sb, civInfo)
	g.renderOverview(&sb, civInfo, civ, bonuses, notciv)
	g.renderPhases(&sb, civInfo.Code, buildings, units)
	g.renderUnitsDetail(&sb, units, heroAuras, catafalqueAuras)
	g.renderSummary(&sb, buildings)
	return sb.String(), nil
}

func (g *Generator) renderHeader(sb *strings.Builder, info civdata.CivCode) {
	caps := strings.ToUpper(info.Code[:1]) + info.Code[1:]
	fmt.Fprintf(sb, "# %s (%s) — Полный отчёт по строениям, юнитам и технологиям\n\n", info.NameEN, caps)
	fmt.Fprintf(sb, "> **Важно:** Данные сгенерированы автоматически из XML/JSON шаблонов в\n")
	fmt.Fprintf(sb, "> `binaries/data/mods/public/simulation/`. Числовые значения соответствуют\n")
	fmt.Fprintf(sb, "> базовым значениям шаблонов с применённым наследованием (parent chain,\n")
	fmt.Fprintf(sb, "> mixins, op=mul/op=add). Эффекты технологий и аур не применены к статам.\n\n")
}

func (g *Generator) renderOverview(sb *strings.Builder, info civdata.CivCode, civ *civdata.Civ, bonuses, notciv []*tech.Technology) {
	fmt.Fprintln(sb, "## Общая информация о цивилизации")
	fmt.Fprintln(sb)
	fmt.Fprintf(sb, "- **Код:** `%s`\n", civ.Code)
	fmt.Fprintf(sb, "- **Культура:** %s\n", civ.Culture())
	fmt.Fprintf(sb, "- **Стартовые юниты:** %s\n", formatStartEntities(civ.StartEntities))
	if len(civ.TeamBonuses) > 0 {
		tb := civ.TeamBonuses[0]
		fmt.Fprintf(sb, "- **Командный бонус (%s):** %s\n", tb.Name, tb.Description)
	}
	fmt.Fprintln(sb)

	fmt.Fprintln(sb, "### Цивилизационные бонусы")
	fmt.Fprintln(sb)
	if len(civ.CivBonuses) == 0 && len(bonuses) == 0 {
		fmt.Fprintln(sb, "*Особых цивилизационных бонусов не зафиксировано.*")
	} else {
		fmt.Fprintln(sb, "| Бонус | Источник | Эффект |")
		fmt.Fprintln(sb, "|-------|----------|--------|")
		for _, b := range civ.CivBonuses {
			fmt.Fprintf(sb, "| %s | civ JSON | %s |\n",
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
			fmt.Fprintf(sb, "| %s%s | %s | %s |\n",
				escapeTable(t.GenericName), auto, t.Name, escapeTable(tip))
		}
	}
	fmt.Fprintln(sb)

	fmt.Fprintf(sb, "### Технологии, НЕДОСТУПНЫЕ %s\n\n", info.NameRU)
	if len(notciv) == 0 {
		fmt.Fprintln(sb, "Явных запретов через `notciv` для этой цивилизации не найдено.")
	} else {
		sort.Slice(notciv, func(i, j int) bool { return notciv[i].Name < notciv[j].Name })
		for _, t := range notciv {
			tip := t.Tooltip
			if tip == "" {
				tip = i18n.DescribeModifications(t.Modifications)
			}
			fmt.Fprintf(sb, "- **%s** — %s\n", t.GenericName, tip)
		}
	}
	fmt.Fprintln(sb)
	fmt.Fprintln(sb, "---")
	fmt.Fprintln(sb)
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
		parts = append(parts, fmt.Sprintf("%s ×%d", name, count))
	}
	return strings.Join(parts, ", ")
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

func escapeTable(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
