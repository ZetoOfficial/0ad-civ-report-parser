package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/i18n"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
)

// renderOverview returns the markdown body for the Civilization Overview tab.
// In epic 1 this preserves the previously-rendered "general info" section
// from the old report.go verbatim. New sections (Identity/Heroes/CivSpecific)
// arrive in epic 2.
func (g *Generator) renderOverview(info civdata.CivCode, civ *civdata.Civ, bonuses, notciv []*tech.Technology) string {
	var sb strings.Builder
	fmt.Fprintln(&sb, "## Общая информация о цивилизации")
	fmt.Fprintln(&sb)
	fmt.Fprintf(&sb, "- **Код:** `%s`\n", civ.Code)
	fmt.Fprintf(&sb, "- **Культура:** %s\n", civ.Culture())
	fmt.Fprintf(&sb, "- **Стартовые юниты:** %s\n", formatStartEntities(civ.StartEntities))
	if len(civ.TeamBonuses) > 0 {
		tb := civ.TeamBonuses[0]
		fmt.Fprintf(&sb, "- **Командный бонус (%s):** %s\n", tb.Name, tb.Description)
	}
	fmt.Fprintln(&sb)

	fmt.Fprintln(&sb, "### Цивилизационные бонусы")
	fmt.Fprintln(&sb)
	if len(civ.CivBonuses) == 0 && len(bonuses) == 0 {
		fmt.Fprintln(&sb, "*Особых цивилизационных бонусов не зафиксировано.*")
	} else {
		fmt.Fprintln(&sb, "| Бонус | Источник | Эффект |")
		fmt.Fprintln(&sb, "|-------|----------|--------|")
		for _, b := range civ.CivBonuses {
			fmt.Fprintf(&sb, "| %s | civ JSON | %s |\n",
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
			fmt.Fprintf(&sb, "| %s%s | %s | %s |\n",
				escapeTable(t.GenericName), auto, t.Name, escapeTable(tip))
		}
	}
	fmt.Fprintln(&sb)

	fmt.Fprintf(&sb, "### Технологии, НЕДОСТУПНЫЕ %s\n\n", info.NameRU)
	if len(notciv) == 0 {
		fmt.Fprintln(&sb, "Явных запретов через `notciv` для этой цивилизации не найдено.")
	} else {
		sort.Slice(notciv, func(i, j int) bool { return notciv[i].Name < notciv[j].Name })
		for _, t := range notciv {
			tip := t.Tooltip
			if tip == "" {
				tip = i18n.DescribeModifications(t.Modifications)
			}
			fmt.Fprintf(&sb, "- **%s** — %s\n", t.GenericName, tip)
		}
	}
	fmt.Fprintln(&sb)
	fmt.Fprintln(&sb, "---")
	fmt.Fprintln(&sb)
	return sb.String()
}
