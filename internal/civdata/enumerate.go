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
	req := e.Element.GetText("Identity/RequiredTechnology")
	switch req {
	case "phase_town", "phase_town_generic", "phase_town_athen", "phase_town_brit", "phase_town_cart", "phase_town_gaul", "phase_town_germ", "phase_town_han", "phase_town_iber", "phase_town_kush", "phase_town_mace", "phase_town_maur", "phase_town_pers", "phase_town_ptol", "phase_town_rome", "phase_town_sele", "phase_town_spart":
		return PhaseTown
	case "phase_city", "phase_city_generic", "phase_city_athen", "phase_city_brit", "phase_city_cart", "phase_city_gaul", "phase_city_germ", "phase_city_han", "phase_city_iber", "phase_city_kush", "phase_city_mace", "phase_city_maur", "phase_city_pers", "phase_city_ptol", "phase_city_rome", "phase_city_sele", "phase_city_spart":
		return PhaseCity
	}
	if strings.HasPrefix(req, "phase_town") {
		return PhaseTown
	}
	if strings.HasPrefix(req, "phase_city") {
		return PhaseCity
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
