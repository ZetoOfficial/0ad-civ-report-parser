package render

import (
	"fmt"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/i18n"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

func FormatCost(e *tmpl.Element) string {
	res := e.Get("Cost/Resources")
	if res == nil {
		return "—"
	}
	parts := []string{}
	for _, key := range []string{"food", "wood", "stone", "metal"} {
		c := res.Child(key)
		if c == nil {
			continue
		}
		v, ok := tmpl.ParseInt(strings.TrimSpace(c.Text))
		if !ok || v == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d %s", v, i18n.ResourceName(key)))
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, ", ")
}

func FormatBuildTime(e *tmpl.Element) string {
	v, ok := e.GetFloat("Cost/BuildTime")
	if !ok {
		return "—"
	}
	return fmt.Sprintf("%s сек", i18n.FormatNumber(v))
}

func FormatHP(e *tmpl.Element) string {
	v, ok := e.GetFloat("Health/Max")
	if !ok {
		return "—"
	}
	return i18n.FormatNumber(v)
}

func FormatArmor(e *tmpl.Element) string {
	dmg := e.Get("Resistance/Entity/Damage")
	if dmg == nil {
		return "—"
	}
	parts := []string{}
	for _, key := range []string{"Hack", "Pierce", "Crush"} {
		c := dmg.Child(key)
		if c == nil {
			continue
		}
		v, ok := tmpl.ParseFloatTrim(c.Text)
		if !ok {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s %s", i18n.DamageType(key), i18n.FormatNumber(v)))
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, ", ")
}

func FormatArmorHPC(e *tmpl.Element) string {
	dmg := e.Get("Resistance/Entity/Damage")
	if dmg == nil {
		return "—"
	}
	get := func(name string) string {
		c := dmg.Child(name)
		if c == nil {
			return "0"
		}
		v, ok := tmpl.ParseFloatTrim(c.Text)
		if !ok {
			return "0"
		}
		return i18n.FormatNumber(v)
	}
	return fmt.Sprintf("%s/%s/%s", get("Hack"), get("Pierce"), get("Crush"))
}

func FormatVision(e *tmpl.Element) string {
	v, ok := e.GetFloat("Vision/Range")
	if !ok {
		return "—"
	}
	return i18n.FormatNumber(v)
}

func FormatGarrison(e *tmpl.Element) string {
	gh := e.Get("GarrisonHolder")
	if gh == nil {
		return ""
	}
	maxV, ok := tmpl.ParseFloatTrim(gh.GetText("Max"))
	if !ok {
		return ""
	}
	classes := gh.GetTokens("List")
	if len(classes) == 0 {
		return i18n.FormatNumber(maxV)
	}
	return fmt.Sprintf("%s (%s)", i18n.FormatNumber(maxV), strings.Join(classes, ", "))
}

func FormatPopulationBonus(e *tmpl.Element) string {
	v, ok := e.GetFloat("Cost/PopulationBonus")
	if !ok || v == 0 {
		return ""
	}
	return fmt.Sprintf("+%s", i18n.FormatNumber(v))
}

func FormatTerritory(e *tmpl.Element) string {
	root := e.Get("TerritoryInfluence")
	if root == nil {
		return ""
	}
	radius := root.GetText("Radius")
	rootFlag := root.GetText("Root")
	if radius == "" {
		return ""
	}
	if strings.EqualFold(rootFlag, "true") {
		return fmt.Sprintf("радиус %s (root)", radius)
	}
	return fmt.Sprintf("радиус %s", radius)
}

func FormatMeleeAttack(e *tmpl.Element) string {
	melee := e.Get("Attack/Melee")
	if melee == nil {
		return ""
	}
	dmgs := []string{}
	for _, t := range []string{"Hack", "Pierce", "Crush"} {
		v, ok := tmpl.ParseFloatTrim(melee.GetText("Damage/" + t))
		if !ok || v == 0 {
			continue
		}
		dmgs = append(dmgs, fmt.Sprintf("%s %s", i18n.DamageType(t), i18n.FormatNumber(v)))
	}
	if len(dmgs) == 0 {
		return ""
	}
	rangeV := melee.GetText("MaxRange")
	repeat := melee.GetText("RepeatTime")
	out := strings.Join(dmgs, ", ")
	if rangeV != "" {
		out += fmt.Sprintf(" (%sм", rangeV)
		if repeat != "" {
			out += fmt.Sprintf(", %sмс", repeat)
		}
		out += ")"
	}
	return out
}

func FormatRangedAttack(e *tmpl.Element) string {
	ranged := e.Get("Attack/Ranged")
	if ranged == nil {
		return ""
	}
	dmgs := []string{}
	for _, t := range []string{"Hack", "Pierce", "Crush"} {
		v, ok := tmpl.ParseFloatTrim(ranged.GetText("Damage/" + t))
		if !ok || v == 0 {
			continue
		}
		dmgs = append(dmgs, fmt.Sprintf("%s %s", i18n.DamageType(t), i18n.FormatNumber(v)))
	}
	if len(dmgs) == 0 {
		return ""
	}
	rangeV := ranged.GetText("MaxRange")
	repeat := ranged.GetText("RepeatTime")
	out := strings.Join(dmgs, ", ")
	if rangeV != "" {
		out += fmt.Sprintf(" (%sм", rangeV)
		if repeat != "" {
			out += fmt.Sprintf(", %sмс", repeat)
		}
		out += ")"
	}
	return out
}

func FormatAttackShort(e *tmpl.Element) string {
	if a := FormatMeleeAttack(e); a != "" {
		return a
	}
	if a := FormatRangedAttack(e); a != "" {
		return a
	}
	return "—"
}

func FormatWalkSpeed(e *tmpl.Element) string {
	v, ok := e.GetFloat("UnitMotion/WalkSpeed")
	if !ok {
		return "—"
	}
	return i18n.FormatNumber(v)
}

func FormatPopulation(e *tmpl.Element) string {
	v, ok := e.GetFloat("Cost/Population")
	if !ok {
		return "—"
	}
	return i18n.FormatNumber(v)
}

func FormatGenericName(e *tmpl.Element) string {
	gen := strings.TrimSpace(e.GetText("Identity/GenericName"))
	spec := strings.TrimSpace(e.GetText("Identity/SpecificName"))
	if gen == "" && spec == "" {
		return ""
	}
	if spec != "" && gen != "" && spec != gen {
		return fmt.Sprintf("%s (%s)", gen, spec)
	}
	if gen != "" {
		return gen
	}
	return spec
}

// formatAttackBonuses returns a formatted list of attack bonuses from a Melee/Ranged/Capture node.
// Each bonus with Classes and a valid Multiplier is rendered as "×<mul> vs <c1>+<c2>+...".
func formatAttackBonuses(modeEl *tmpl.Element) string {
	if modeEl == nil {
		return ""
	}
	bonuses := modeEl.Get("Bonuses")
	if bonuses == nil {
		return ""
	}
	var parts []string
	for _, child := range bonuses.Children {
		classes := child.GetTokens("Classes")
		if len(classes) == 0 {
			continue
		}
		mul, ok := tmpl.ParseFloatTrim(child.GetText("Multiplier"))
		if !ok {
			continue
		}
		parts = append(parts, fmt.Sprintf("×%s vs %s", i18n.FormatNumber(mul), strings.Join(classes, "+")))
	}
	return strings.Join(parts, ", ")
}

// formatPreferredClasses returns the PreferredClasses tokens from a Melee/Ranged node joined by ", ".
func formatPreferredClasses(modeEl *tmpl.Element) string {
	if modeEl == nil {
		return ""
	}
	tokens := modeEl.GetTokens("PreferredClasses")
	return strings.Join(tokens, ", ")
}

// formatSplash formats the Splash sub-element of a Melee/Ranged node.
// formatAttackBonuses is called on the Splash node itself because Splash can
// carry its own <Bonuses> (splash multipliers vs class, e.g. ×2 vs Infantry).
func formatSplash(modeEl *tmpl.Element) string {
	if modeEl == nil {
		return ""
	}
	splash := modeEl.Get("Splash")
	if splash == nil {
		return ""
	}

	var dmgs []string
	for _, t := range []string{"Hack", "Pierce", "Crush", "Fire", "Poison"} {
		v, ok := tmpl.ParseFloatTrim(splash.GetText("Damage/" + t))
		if !ok || v == 0 {
			continue
		}
		dmgs = append(dmgs, fmt.Sprintf("%s %s", i18n.DamageType(t), i18n.FormatNumber(v)))
	}

	shapeRaw := splash.GetText("Shape")
	var shape string
	switch shapeRaw {
	case "Linear":
		shape = "линия"
	case "", "Circular":
		shape = "круг"
	default:
		shape = shapeRaw
	}

	rangeV := splash.GetText("Range")
	ffRaw := splash.GetText("FriendlyFire")
	var ffText string
	if strings.EqualFold(ffRaw, "false") {
		ffText = "не задевает союзников"
	} else {
		ffText = "задевает союзников"
	}

	if len(dmgs) == 0 && rangeV == "" {
		return ""
	}

	var sb strings.Builder
	if len(dmgs) > 0 {
		sb.WriteString(strings.Join(dmgs, ", "))
		sb.WriteString(", ")
	}
	fmt.Fprintf(&sb, "%s R=%s, %s", shape, rangeV, ffText)

	if bonuses := formatAttackBonuses(splash); bonuses != "" {
		sb.WriteString(", ")
		sb.WriteString(bonuses)
	}

	return sb.String()
}

// formatApplyStatuses returns a formatted slice of ApplyStatus entries from a
// Melee/Ranged attack-mode element.  byCode maps status code (e.g. "Poisoned")
// to its statusEffect descriptor loaded from data/status_effects/.
//
// Entries with zero/negative Duration are skipped (no-op).  Entries where
// BlockChance > 0 AND Duration <= 0 (the template_unit_siege default) are also
// skipped.
func formatApplyStatuses(modeEl *tmpl.Element, byCode map[string]statusEffect) []string {
	if modeEl == nil {
		return nil
	}
	as := modeEl.Get("ApplyStatus")
	if as == nil {
		return nil
	}
	var out []string
	for _, child := range as.Children {
		code := child.Name

		// Parse Duration once; used for both the BlockChance check and the
		// seconds conversion below.
		durRaw := child.GetText("Duration")
		dur, durOK := tmpl.ParseFloatTrim(durRaw)

		// BlockChance edge-case: skip no-op defaults (BlockChance>0 and Duration<=0).
		blockRaw := child.GetText("BlockChance")
		if blockRaw != "" {
			bc, ok := tmpl.ParseFloatTrim(blockRaw)
			if ok && bc > 0 {
				if durRaw == "" || !durOK || dur <= 0 {
					continue
				}
			}
		}

		// Skip any entry with zero or missing Duration.
		if !durOK || dur <= 0 {
			continue
		}
		durationSec := dur / 1000.0

		// Resolve display name.
		name := code
		if effect, ok := byCode[code]; ok {
			name = effect.StatusName
		}

		// Build damage string.
		var dmgParts []string
		for _, t := range []string{"Hack", "Pierce", "Crush", "Fire", "Poison"} {
			v, ok := tmpl.ParseFloatTrim(child.GetText("Damage/" + t))
			if !ok || v == 0 {
				continue
			}
			dmgParts = append(dmgParts, fmt.Sprintf("%s %s", i18n.DamageType(t), i18n.FormatNumber(v)))
		}
		dmgStr := strings.Join(dmgParts, ", ")

		// Interval in seconds (optional).
		intervalSec := ""
		if intervalRaw := child.GetText("Interval"); intervalRaw != "" {
			if iv, ok := tmpl.ParseFloatTrim(intervalRaw); ok && iv > 0 {
				intervalSec = i18n.FormatNumber(iv / 1000.0)
			}
		}

		anchor := strings.ToLower(code)
		durationStr := i18n.FormatNumber(durationSec)

		var line string
		switch {
		case dmgStr != "" && intervalSec != "":
			line = fmt.Sprintf("%s: %s/%sс × %sс (см. common.md#%s)", name, dmgStr, intervalSec, durationStr, anchor)
		case dmgStr == "" && intervalSec != "":
			line = fmt.Sprintf("%s: каждые %sс × %sс (см. common.md#%s)", name, intervalSec, durationStr, anchor)
		case dmgStr != "":
			// damage present, no interval
			line = fmt.Sprintf("%s: %s × %sс (см. common.md#%s)", name, dmgStr, durationStr, anchor)
		default:
			// neither damage nor interval: omit the empty prefix to avoid double space
			line = fmt.Sprintf("%s: %sс (см. common.md#%s)", name, durationStr, anchor)
		}
		out = append(out, line)
	}
	return out
}

// FormatCaptureResistance returns the Resistance/Entity/Capture value as a
// formatted string, or "" if the field is absent.
func FormatCaptureResistance(e *tmpl.Element) string {
	v, ok := e.GetFloat("Resistance/Entity/Capture")
	if !ok {
		return ""
	}
	return i18n.FormatNumber(v)
}

// FormatStatusEffectResistance returns a formatted list of per-status
// resistance multipliers from Resistance/Entity/StatusEffect.
// Each child element is treated as "<code>×<multiplier>"; codes are
// forced to lowercase per spec convention (resistance side).
func FormatStatusEffectResistance(e *tmpl.Element) string {
	se := e.Get("Resistance/Entity/StatusEffect")
	if se == nil {
		return ""
	}
	var parts []string
	for _, child := range se.Children {
		code := strings.ToLower(child.Name)
		mul, ok := tmpl.ParseFloatTrim(child.Text)
		if !ok {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s ×%s", code, i18n.FormatNumber(mul)))
	}
	return strings.Join(parts, ", ")
}

// formatCaptureAttack formats the Capture attack from the parent Attack node.
// The parameter is the parent Attack element (not Capture directly) because
// call sites already hold the Attack node in scope.
func formatCaptureAttack(attackEl *tmpl.Element) string {
	if attackEl == nil {
		return ""
	}
	cap := attackEl.Get("Capture")
	if cap == nil {
		return ""
	}
	rate, ok := tmpl.ParseFloatTrim(cap.GetText("Capture"))
	if !ok {
		return ""
	}

	out := fmt.Sprintf("захват %s", i18n.FormatNumber(rate))

	rangeV := cap.GetText("MaxRange")
	repeat := cap.GetText("RepeatTime")
	if rangeV != "" && repeat != "" {
		out += fmt.Sprintf(" (%sм, %sмс)", rangeV, repeat)
	} else if rangeV != "" {
		out += fmt.Sprintf(" (%sм)", rangeV)
	} else if repeat != "" {
		out += fmt.Sprintf(" (%sмс)", repeat)
	}

	restricted := cap.GetTokens("RestrictedClasses")
	if len(restricted) > 0 {
		out += "; исключает: " + strings.Join(restricted, ", ")
	}

	if bonuses := formatAttackBonuses(cap); bonuses != "" {
		out += "; " + bonuses
	}

	return out
}
