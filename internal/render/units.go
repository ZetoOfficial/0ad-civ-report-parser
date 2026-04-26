package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/aura"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/i18n"
)

func (g *Generator) renderUnitsDetail(sb *strings.Builder, units []civdata.Entity, heroAuras, catafalqueAuras []*aura.Aura) {
	fmt.Fprintln(sb, "## Приложение: Детальная информация по типам юнитов")
	fmt.Fprintln(sb)

	groups := classifyUnits(units)
	order := []struct {
		key   string
		title string
	}{
		{"support", "Поддержка и работники"},
		{"infantry", "Пехота"},
		{"cavalry", "Конница"},
		{"champion", "Чемпионы"},
		{"hero", "Герои"},
		{"siege", "Осадные орудия"},
		{"healer", "Лекари"},
		{"ship", "Корабли"},
		{"catafalque", "Катафалк (Capture-the-Relic)"},
		{"other", "Прочее"},
	}
	for _, sec := range order {
		list := groups[sec.key]
		if len(list) == 0 {
			continue
		}
		sort.Slice(list, func(i, j int) bool { return list[i].Basename() < list[j].Basename() })
		fmt.Fprintf(sb, "### %s\n\n", sec.title)
		for _, u := range list {
			g.renderUnitBlock(sb, u)
			if sec.key == "hero" {
				g.renderHeroAuras(sb, u, heroAuras)
			}
			if sec.key == "catafalque" {
				g.renderCatafalqueAuras(sb, catafalqueAuras)
			}
			fmt.Fprintln(sb)
		}
	}
}

func classifyUnits(units []civdata.Entity) map[string][]civdata.Entity {
	groups := map[string][]civdata.Entity{}
	for _, u := range units {
		key := classifyUnit(u)
		groups[key] = append(groups[key], u)
	}
	return groups
}

func classifyUnit(u civdata.Entity) string {
	switch {
	case civdata.IsCatafalque(u):
		return "catafalque"
	case civdata.IsHero(u):
		return "hero"
	case civdata.IsChampion(u):
		return "champion"
	case civdata.IsHealer(u):
		return "healer"
	case civdata.IsSupport(u):
		return "support"
	case civdata.IsShip(u):
		return "ship"
	case civdata.IsSiege(u):
		return "siege"
	case strings.HasPrefix(u.Basename(), "infantry_"):
		return "infantry"
	case strings.HasPrefix(u.Basename(), "cavalry_"):
		return "cavalry"
	}
	return "other"
}

func (g *Generator) renderUnitBlock(sb *strings.Builder, u civdata.Entity) {
	name := FormatGenericName(u.Element)
	if name == "" {
		name = u.Basename()
	}
	classes := u.Element.GetTokens("Identity/VisibleClasses")
	classBadge := ""
	if len(classes) > 0 {
		classBadge = " — " + strings.Join(classes, ", ")
	}
	fmt.Fprintf(sb, "#### %s%s\n\n", name, classBadge)
	fmt.Fprintln(sb, "| Параметр | Значение |")
	fmt.Fprintln(sb, "|----------|----------|")
	fmt.Fprintf(sb, "| Стоимость | %s |\n", FormatCost(u.Element))
	fmt.Fprintf(sb, "| Время тренировки | %s |\n", FormatBuildTime(u.Element))
	fmt.Fprintf(sb, "| ОЗ | %s |\n", FormatHP(u.Element))
	fmt.Fprintf(sb, "| Броня (H/P/C) | %s |\n", FormatArmorHPC(u.Element))
	fmt.Fprintf(sb, "| Скорость | %s |\n", FormatWalkSpeed(u.Element))
	fmt.Fprintf(sb, "| Обзор | %s |\n", FormatVision(u.Element))
	fmt.Fprintf(sb, "| Население | %s |\n", FormatPopulation(u.Element))

	if a := FormatMeleeAttack(u.Element); a != "" {
		fmt.Fprintf(sb, "| Атака (ближ.) | %s |\n", a)
	}
	if a := FormatRangedAttack(u.Element); a != "" {
		fmt.Fprintf(sb, "| Атака (стрельба) | %s |\n", a)
	}

	prom := u.Element.Get("Promotion")
	if prom != nil {
		entity := prom.GetText("Entity")
		xp := prom.GetText("RequiredXp")
		if entity != "" || xp != "" {
			fmt.Fprintf(sb, "| Промоушн | %s @ %s XP |\n", entity, xp)
		}
	}

	gathering := u.Element.Get("ResourceGatherer/Rates")
	if gathering != nil && len(gathering.Children) > 0 {
		parts := []string{}
		for _, c := range gathering.Children {
			parts = append(parts, fmt.Sprintf("%s=%s", c.Name, strings.TrimSpace(c.Text)))
		}
		fmt.Fprintf(sb, "| Сбор ресурсов | %s |\n", strings.Join(parts, ", "))
	}

	fmt.Fprintln(sb)
}

func (g *Generator) renderHeroAuras(sb *strings.Builder, u civdata.Entity, heroAuras []*aura.Aura) {
	matched := []*aura.Aura{}
	heroName := strings.TrimPrefix(u.Basename(), "hero_")
	for _, a := range heroAuras {
		if strings.Contains(a.Name, heroName) {
			matched = append(matched, a)
		}
	}
	auraTokens := u.Element.GetTokens("Auras")
	for _, tok := range auraTokens {
		base := strings.TrimPrefix(tok, "units/heroes/")
		for _, a := range heroAuras {
			if a.Name == base {
				if !containsAura(matched, a) {
					matched = append(matched, a)
				}
			}
		}
	}
	if len(matched) == 0 {
		return
	}
	fmt.Fprintln(sb, "**Ауры:**")
	fmt.Fprintln(sb)
	fmt.Fprintln(sb, "| Файл | Тип | Радиус | Цель | Эффект |")
	fmt.Fprintln(sb, "|------|-----|--------|------|--------|")
	for _, a := range matched {
		radius := "—"
		if a.Radius > 0 {
			radius = i18n.FormatNumber(a.Radius)
		}
		target := strings.Join(a.AffectsHumanReadable(), ", ")
		desc := a.AuraDescription
		if desc == "" {
			desc = i18n.DescribeModifications(a.Modifications)
		}
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s |\n",
			escapeTable(a.Name), escapeTable(a.Type), radius, escapeTable(target), escapeTable(desc))
	}
	fmt.Fprintln(sb)
}

func containsAura(list []*aura.Aura, a *aura.Aura) bool {
	for _, x := range list {
		if x.Path == a.Path {
			return true
		}
	}
	return false
}

func (g *Generator) renderCatafalqueAuras(sb *strings.Builder, auras []*aura.Aura) {
	if len(auras) == 0 {
		return
	}
	fmt.Fprintln(sb, "**Ауры катафалка (применяются при захвате реликвии):**")
	fmt.Fprintln(sb)
	fmt.Fprintln(sb, "| Файл | Тип | Радиус | Цель | Эффект |")
	fmt.Fprintln(sb, "|------|-----|--------|------|--------|")
	for _, a := range auras {
		radius := "—"
		if a.Radius > 0 {
			radius = i18n.FormatNumber(a.Radius)
		}
		target := strings.Join(a.AffectsHumanReadable(), ", ")
		desc := a.AuraDescription
		if desc == "" {
			desc = i18n.DescribeModifications(a.Modifications)
		}
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s |\n",
			escapeTable(a.Name), escapeTable(a.Type), radius, escapeTable(target), escapeTable(desc))
	}
	fmt.Fprintln(sb)
}
