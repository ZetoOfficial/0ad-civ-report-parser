package civdata

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

type Entity struct {
	TemplateID string
	Path       string
	Element    *tmpl.Element
}

func (e Entity) Basename() string {
	return strings.TrimSuffix(filepath.Base(e.Path), ".xml")
}

func (e Entity) HasClass(name string) bool {
	classes := append([]string{}, e.Element.GetTokens("Identity/Classes")...)
	classes = append(classes, e.Element.GetTokens("Identity/VisibleClasses")...)
	for _, c := range classes {
		if c == name {
			return true
		}
	}
	return false
}

func IsHero(e Entity) bool {
	if strings.HasPrefix(e.Basename(), "hero_") {
		return true
	}
	if e.HasClass("Hero") {
		return true
	}
	return false
}

func IsChampion(e Entity) bool {
	if strings.HasPrefix(e.Basename(), "champion_") {
		return true
	}
	return e.HasClass("Champion")
}

func IsSupport(e Entity) bool {
	return strings.HasPrefix(e.Basename(), "support_")
}

func IsShip(e Entity) bool {
	if strings.HasPrefix(e.Basename(), "ship_") {
		return true
	}
	return e.HasClass("Ship")
}

func IsSiege(e Entity) bool {
	if strings.HasPrefix(e.Basename(), "siege_") {
		return true
	}
	return e.HasClass("Siege")
}

func IsCatafalque(e Entity) bool {
	return e.Basename() == "catafalque"
}

func IsHealer(e Entity) bool {
	return strings.Contains(e.Basename(), "healer")
}

type Phase int

const (
	PhaseVillage Phase = iota
	PhaseTown
	PhaseCity
)

func (p Phase) String() string {
	switch p {
	case PhaseVillage:
		return "village"
	case PhaseTown:
		return "town"
	case PhaseCity:
		return "city"
	}
	return ""
}

func (p Phase) RU() string {
	switch p {
	case PhaseVillage:
		return "Village"
	case PhaseTown:
		return "Town"
	case PhaseCity:
		return "City"
	}
	return ""
}

func BuildingPhase(e Entity) Phase {
	// R28 format: Identity/Requirements/Techs (tokens-list after merge).
	techs := e.Element.GetTokens("Identity/Requirements/Techs")
	for _, t := range techs {
		if t == "" || strings.HasPrefix(t, "-") {
			continue
		}
		switch {
		case strings.HasPrefix(t, "phase_city"):
			return PhaseCity
		case strings.HasPrefix(t, "phase_town"):
			return PhaseTown
		case strings.HasPrefix(t, "phase_village"):
			return PhaseVillage
		}
	}
	// Legacy fallback for any old-format templates with Identity/RequiredTechnology.
	req := e.Element.GetText("Identity/RequiredTechnology")
	switch {
	case strings.HasPrefix(req, "phase_city"):
		return PhaseCity
	case strings.HasPrefix(req, "phase_town"):
		return PhaseTown
	case strings.HasPrefix(req, "phase_village"):
		return PhaseVillage
	}
	return PhaseVillage
}

var buildingOrderHints = []string{
	"civil_centre",
	"house",
	"storehouse",
	"farmstead",
	"corral",
	"field",
	"dock",
	"barracks",
	"stable",
	"range",
	"outpost",
	"sentry_tower",
	"defense_tower",
	"palisade",
	"temple",
	"forge",
	"market",
	"tower_bolt",
	"wall_short",
	"wall_medium",
	"wall_long",
	"wall_gate",
	"wall_tower",
	"wallset_palisade",
	"wallset_stone",
	"fortress",
	"arsenal",
	"wonder",
}

func BuildingSortKey(b Entity) (rank int, name string) {
	base := b.Basename()
	for i, hint := range buildingOrderHints {
		if base == hint {
			return i, base
		}
	}
	for i, hint := range buildingOrderHints {
		if strings.Contains(base, hint) {
			return 1000 + i, base
		}
	}
	return 9999, base
}

func SortBuildingsByOrder(buildings []Entity) {
	sort.SliceStable(buildings, func(i, j int) bool {
		ai, _ := BuildingSortKey(buildings[i])
		aj, _ := BuildingSortKey(buildings[j])
		if ai != aj {
			return ai < aj
		}
		return buildings[i].Basename() < buildings[j].Basename()
	})
}

func GroupByPhase(buildings []Entity) map[Phase][]Entity {
	out := map[Phase][]Entity{
		PhaseVillage: {},
		PhaseTown:    {},
		PhaseCity:    {},
	}
	for _, b := range buildings {
		p := BuildingPhase(b)
		out[p] = append(out[p], b)
	}
	for k := range out {
		SortBuildingsByOrder(out[k])
	}
	return out
}
