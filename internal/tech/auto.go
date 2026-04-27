package tech

import (
	"path/filepath"
	"sort"
)

// GlobalAutoResearch returns autoResearch technologies from the root
// technologies/ directory whose requirements have no civ filter — i.e.
// they apply automatically to every civ. In R28 this includes
// unit_advanced, unit_elite, phase_village, soldier_ranged_experience,
// upgrade_rank_advanced_mercenary.
//
// Civ-specific bonuses live in technologies/civbonuses/ and are returned
// by AllCivBonuses(civ); they are not included here even though they
// also have autoResearch=true.
func (c *Catalog) GlobalAutoResearch() ([]*Technology, error) {
	matches, err := filepath.Glob(filepath.Join(c.dir, "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	out := []*Technology{}
	for _, p := range matches {
		t, err := Load(p)
		if err != nil {
			return nil, err
		}
		if !t.AutoResearch {
			continue
		}
		if RequiresCiv(t.Requirements) != "" {
			continue
		}
		out = append(out, t)
	}
	return out, nil
}
