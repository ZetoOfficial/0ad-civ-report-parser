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
func (g *Generator) renderStructree(civCode string, res *civdata.ReachResult,
	heroAuras, catafalqueAuras []*aura.Aura) string {
	var sb strings.Builder
	g.renderPhases(&sb, civCode, res)
	g.renderUnitsDetail(&sb, res.Units, heroAuras, catafalqueAuras)
	g.renderSummary(&sb, res.Buildings)
	return sb.String()
}

func (g *Generator) renderPhases(sb *strings.Builder, civCode string, res *civdata.ReachResult) {
	groups := civdata.GroupByPhase(res.Buildings)
	phases := []struct {
		p     civdata.Phase
		title string
	}{
		{civdata.PhaseVillage, "VILLAGE PHASE"},
		{civdata.PhaseTown, "TOWN PHASE"},
		{civdata.PhaseCity, "CITY PHASE"},
	}
	unitByID := indexUnits(civCode, res.Units)
	for _, ph := range phases {
		fmt.Fprintf(sb, "## %s\n\n", ph.title)
		list := groups[ph.p]
		wallsetsHere := filterWallSetsByPhase(res.WallSets, ph.p)
		if len(list) == 0 && len(wallsetsHere) == 0 {
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
		for _, ws := range wallsetsHere {
			g.renderWallSetBlock(sb, civCode, ws)
			fmt.Fprintln(sb, "---")
			fmt.Fprintln(sb)
		}
	}
}

// filterWallSetsByPhase returns wallsets belonging to the given phase.
// The order is preserved from the input slice, which IdentifyWallSets already
// sorts by BuildingSortKey(Wrapper) for deterministic rendering.
func filterWallSetsByPhase(wallsets []*civdata.WallSetGroup, phase civdata.Phase) []*civdata.WallSetGroup {
	var out []*civdata.WallSetGroup
	for _, ws := range wallsets {
		if ws.Phase == phase {
			out = append(out, ws)
		}
	}
	return out
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
	if g2 := FormatGarrison(b.Element); g2 != "" {
		fmt.Fprintf(sb, "| Гарнизон | %s |\n", g2)
	}
	if v := FormatVision(b.Element); v != "—" {
		fmt.Fprintf(sb, "| Обзор | %s |\n", v)
	}
	fmt.Fprintln(sb)

	g.renderTrains(sb, civCode, b, unitByID)
	g.renderResearches(sb, civCode, b)
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

// renderResearches writes the "#### Исследует" sub-table for a building.
// It collects techs from three paths (Researcher, Trainer, ProductionQueue),
// deduplicates, and renders each row using Index-aware pair expansion.
func (g *Generator) renderResearches(sb *strings.Builder, civCode string, b civdata.Entity) {
	// Tech sources: Researcher/Technologies is where R28 stores researchable
	// techs; Trainer/ProductionQueue Technologies retained for compatibility.
	var rawTokens []string
	for _, path := range []string{
		"Researcher/Technologies",
		"Trainer/Technologies",
		"ProductionQueue/Technologies",
	} {
		rawTokens = append(rawTokens, b.Element.GetTokens(path)...)
	}

	// Filter empty / removal tokens.
	var cleaned []string
	for _, t := range rawTokens {
		t = strings.TrimSpace(t)
		if t == "" || strings.HasPrefix(t, "-") {
			continue
		}
		cleaned = append(cleaned, t)
	}
	if len(cleaned) == 0 {
		return
	}

	// Render, deduplicating by resolved tech name.
	seen := map[string]struct{}{}
	var rows []string

	for _, tok := range cleaned {
		// Substitute {civ}/{native} placeholders so that tokens like
		// "phase_town_{civ}" resolve correctly for each civilization.
		tok = tmpl.SubstCiv(tok, civCode)

		// Try pair expansion first.
		if top, bot, ok := tech.ExpandPair(g.Catalog, tok); ok {
			for _, t := range []*tech.Technology{top, bot} {
				if _, dup := seen[t.Name]; dup {
					continue
				}
				seen[t.Name] = struct{}{}
				rows = append(rows, formatPairRow(t, g.Index, civCode))
			}
			continue
		}

		// Resolve to civ-specific variant via Index.
		var t *tech.Technology
		if g.Index != nil {
			t = g.Index.ResolveForCiv(tok, civCode)
		} else {
			// Fallback: direct catalog lookup.
			if ct, err := g.Catalog.ByName(tok); err == nil {
				t = ct
			}
		}
		if t == nil {
			rows = append(rows, fmt.Sprintf("| %s | — | — | — | (не найдено) |", escapeTable(tok)))
			continue
		}
		if _, dup := seen[t.Name]; dup {
			continue
		}
		seen[t.Name] = struct{}{}
		rows = append(rows, formatTechRow(t, g.Index, civCode))
	}

	if len(rows) == 0 {
		return
	}

	fmt.Fprintln(sb, "#### Исследует")
	fmt.Fprintln(sb)
	fmt.Fprintln(sb, "| Технология | Стоимость | Время | Фаза | Эффект |")
	fmt.Fprintln(sb, "|-----------|-----------|-------|------|--------|")
	for _, row := range rows {
		fmt.Fprintln(sb, row)
	}
	fmt.Fprintln(sb)
}

// requirementPhase extracts the phase label from a tech requirement.
// idx is used for civ-variant phase tech resolution; may be nil (legacy mode).
func requirementPhase(t *tech.Technology, civ string, idx *tech.Index) string {
	if t == nil {
		return "—"
	}
	req := t.Requirements
	if req == nil {
		// Fallback: derive phase from t.Supersedes chain when no explicit requirement.
		if idx != nil && t.Supersedes != "" {
			if lbl := phaseLabelFromSupersedes(t.Supersedes); lbl != "" {
				return lbl
			}
		}
		return "—"
	}

	raw := extractRawPhase(req)

	if raw == "" {
		// No tech/all-tech requirement found — try Supersedes-based fallback.
		if idx != nil && t.Supersedes != "" {
			if lbl := phaseLabelFromSupersedes(t.Supersedes); lbl != "" {
				return lbl
			}
		}
		return "—"
	}

	// Direct known labels.
	if label := i18n.PhaseRequirement(raw); label != "" && label != raw {
		return label
	}

	// Phase-variant token: resolve to civ-specific tech and derive label.
	if strings.HasPrefix(raw, "phase_") && idx != nil {
		resolved := idx.ResolveForCiv(raw, civ)
		if resolved != nil {
			if label := i18n.PhaseRequirement(resolved.Name); label != "" && label != resolved.Name {
				return label
			}
			// Derive from chain: Supersedes tells us which phase this upgrades from.
			if lbl := phaseLabelFromSupersedes(resolved.Supersedes); lbl != "" {
				return lbl
			}
		}
	}

	return i18n.PhaseRequirement(raw)
}

// extractRawPhase returns the raw phase token string from a Requirements map,
// or "" if none found.
func extractRawPhase(req tech.Requirements) string {
	if v, ok := req["tech"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if all, ok := req["all"]; ok {
		if list, ok := all.([]any); ok {
			for _, item := range list {
				if m, ok := item.(map[string]any); ok {
					if v, ok := m["tech"].(string); ok {
						if strings.HasPrefix(v, "phase_") {
							return v
						}
					}
				}
			}
		}
	}
	return ""
}

// phaseLabelFromSupersedes returns the phase label of the next phase above the
// one named by supersedes. E.g. if a tech supersedes "phase_village" it is the
// Town phase tech, if it supersedes "phase_town" it is the City phase tech.
func phaseLabelFromSupersedes(s string) string {
	switch {
	case strings.HasPrefix(s, "phase_village"):
		return "Town"
	case strings.HasPrefix(s, "phase_town"):
		return "City"
	}
	return ""
}
