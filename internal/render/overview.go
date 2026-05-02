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
	bonuses, civSpecific, notciv []*tech.Technology,
	units, buildings []civdata.Entity,
	heroAuras []*aura.Aura,
) string {
	var sb strings.Builder
	g.overviewIdentity(&sb, civ, player)
	if g.IncludeHistory && player != nil && player.History != "" {
		g.overviewHistory(&sb, player)
	}
	g.overviewHeroes(&sb, units, heroAuras)
	g.overviewCivSpecificStructures(&sb, buildings)
	g.overviewSpecificTechnologies(&sb, info.Code, civSpecific)
	g.overviewCivBonuses(&sb, info.Code, civ, bonuses)
	g.overviewGlobalAutoResearchFooter(&sb)
	g.overviewTeamBonus(&sb, civ, teamBonus)
	g.overviewNotCiv(&sb, info, notciv)
	return sb.String()
}

// 1. Идентичность

func (g *Generator) overviewIdentity(sb *strings.Builder,
	civ *civdata.Civ, player *civdata.PlayerTemplate) {

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

// 5. Уникальные технологии (короткий список)

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
