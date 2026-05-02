package render

import (
	"fmt"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

// renderWallSetBlock writes a wallset section: header, piece-roles table.
func (g *Generator) renderWallSetBlock(sb *strings.Builder, civCode string, ws *civdata.WallSetGroup) {
	_ = civCode // reserved for future civ-specific overrides

	// Header: "### Стены: <SpecificName> (<GenericName>)" — spec requires specific first.
	var name string
	if ws.Wrapper.Element != nil {
		generic := strings.TrimSpace(ws.Wrapper.Element.GetText("Identity/GenericName"))
		specific := strings.TrimSpace(ws.Wrapper.Element.GetText("Identity/SpecificName"))
		if specific != "" {
			if generic != "" && generic != specific {
				name = fmt.Sprintf("%s (%s)", specific, generic)
			} else {
				name = specific
			}
		} else {
			name = generic
		}
	}
	if name == "" {
		name = ws.Wrapper.Basename()
	}
	fmt.Fprintf(sb, "### Стены: %s\n\n", name)

	if len(ws.Pieces) == 0 {
		fmt.Fprintln(sb, "*Нет данных о секциях.*")
		fmt.Fprintln(sb)
		return
	}

	fmt.Fprintln(sb, "| Часть | Стоимость | Время | ОЗ | Броня (H/P/C) | Атака | Прочее |")
	fmt.Fprintln(sb, "|-------|-----------|-------|-----|---------------|-------|--------|")
	for _, p := range ws.Pieces {
		label := roleLabel(p.Role)
		cost := FormatCost(p.Entity.Element)
		buildTime := FormatBuildTime(p.Entity.Element)
		hp := FormatHP(p.Entity.Element)
		armor := FormatArmorHPC(p.Entity.Element)
		atk := wallPieceAttack(p)
		misc := wallPieceMisc(p, ws.Wrapper.Element)
		fmt.Fprintf(sb, "| %s | %s | %s | %s | %s | %s | %s |\n",
			escapeTable(label),
			escapeTable(cost),
			escapeTable(buildTime),
			escapeTable(hp),
			escapeTable(armor),
			escapeTable(atk),
			escapeTable(misc),
		)
	}
	fmt.Fprintln(sb)
}

// roleLabel translates a WallSet piece role name to Russian.
func roleLabel(role string) string {
	switch role {
	case "Tower":
		return "Башня"
	case "Gate":
		return "Ворота"
	case "WallLong":
		return "Длинная секция"
	case "WallMedium":
		return "Средняя секция"
	case "WallShort":
		return "Короткая секция"
	case "Fort":
		return "Форт"
	case "WallCurves":
		return "Изгиб стены"
	case "WallEnd":
		return "Окончание стены"
	}
	return role
}

// wallPieceAttack returns the attack summary for Tower pieces; "—" for others.
func wallPieceAttack(p civdata.WallPiece) string {
	if p.Role == "Tower" {
		return FormatAttackShort(p.Entity.Element)
	}
	return "—"
}

// wallPieceMisc returns extra info for a wall piece based on its role.
func wallPieceMisc(p civdata.WallPiece, wrapperEl *tmpl.Element) string {
	switch p.Role {
	case "Tower":
		if wrapperEl != nil {
			if v := wrapperEl.GetText("WallSet/MaxTowerOverlap"); v != "" {
				return fmt.Sprintf("MaxOverlap=%s", v)
			}
		}
		return ""
	case "Gate":
		return "проходные ворота"
	case "WallLong", "WallMedium", "WallShort":
		// Width from Footprint/Square/@width attribute.
		fp := p.Entity.Element.Get("Footprint/Square")
		if fp != nil {
			if w := fp.Attr("width"); w != "" {
				return fmt.Sprintf("длина %s", w)
			}
		}
		return ""
	}
	return ""
}
