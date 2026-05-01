package render

import (
	"fmt"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/i18n"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
)

// formatTechCost formats a tech.Cost as a human-readable resource list,
// e.g. "200 Еда, 100 Дерево". Returns "—" if all fields are zero.
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

// formatTechRow returns a markdown table row for a normal (non-pair) technology.
func formatTechRow(t *tech.Technology, idx *tech.Index, civ string) string {
	name := i18n.TechDisplayName(t, civ)
	suffix := chainSuffix(idx, t, civ)
	return fmt.Sprintf("| %s%s | %s | %s | %s | %s |",
		escapeTable(name), escapeTable(suffix),
		formatTechCost(t.Cost),
		formatTechTime(t),
		requirementPhase(t, civ, idx),
		escapeTable(formatTechEffect(t)),
	)
}

// formatPairRow returns one markdown table row for one half of a pair tech.
// The name is prefixed with U+25D0 ("◐") to mark it as a paired option.
func formatPairRow(t *tech.Technology, idx *tech.Index, civ string) string {
	name := "◐ " + i18n.TechDisplayName(t, civ) + " — парная (выбрать одно)"
	suffix := chainSuffix(idx, t, civ)
	return fmt.Sprintf("| %s%s | %s | %s | %s | %s |",
		escapeTable(name), escapeTable(suffix),
		formatTechCost(t.Cost),
		formatTechTime(t),
		requirementPhase(t, civ, idx),
		escapeTable(formatTechEffect(t)),
	)
}

// displayNameByName resolves a technology name to a human-readable label
// using the Index. Prefers civ-specific display names via ResolveForCiv,
// then falls back to the raw name.
func displayNameByName(idx *tech.Index, name, civ string) string {
	if idx == nil || name == "" {
		return name
	}
	if t := idx.ResolveForCiv(name, civ); t != nil {
		return i18n.TechDisplayName(t, civ)
	}
	return name
}

// displayNamesByName is the batch variant of displayNameByName.
func displayNamesByName(idx *tech.Index, names []string, civ string) []string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = displayNameByName(idx, n, civ)
	}
	return out
}

// chainSuffix returns parenthetical chain text like " (заменяет: X; апгрейд от Y)"
// or "" if there is nothing to report. Civ-aware for ReplacedBy.
func chainSuffix(idx *tech.Index, t *tech.Technology, civ string) string {
	if idx == nil {
		return ""
	}
	ch := idx.Chain(t.Name)
	var parts []string
	if len(ch.Replaces) > 0 {
		parts = append(parts, fmt.Sprintf("заменяет: %s", strings.Join(displayNamesByName(idx, ch.Replaces, civ), ", ")))
	}
	if ch.Supersedes != "" {
		parts = append(parts, fmt.Sprintf("апгрейд от %s", displayNameByName(idx, ch.Supersedes, civ)))
	}
	// ReplacedBy: only mention if the active tech for civ differs from t.
	if len(ch.ReplacedBy) > 0 {
		active := idx.ResolveForCiv(t.Name, civ)
		if active != nil && active.Name != t.Name {
			parts = append(parts, fmt.Sprintf("заменяется на: %s", i18n.TechDisplayName(active, civ)))
		}
	}
	if ch.SupersededBy != "" {
		parts = append(parts, fmt.Sprintf("апгрейдится в: %s", displayNameByName(idx, ch.SupersededBy, civ)))
	}
	if len(parts) == 0 {
		return ""
	}
	return " (" + strings.Join(parts, "; ") + ")"
}

// formatTechTime returns "N сек" if ResearchTime > 0, otherwise "—".
func formatTechTime(t *tech.Technology) string {
	if t.ResearchTime > 0 {
		return fmt.Sprintf("%s сек", i18n.FormatNumber(t.ResearchTime))
	}
	return "—"
}

// formatTechEffect returns the human-readable tech effect string.
// Uses t.Tooltip if non-empty, else derives text from t.Modifications.
func formatTechEffect(t *tech.Technology) string {
	if t.Tooltip != "" {
		return t.Tooltip
	}
	return i18n.DescribeModifications(t.Modifications)
}
