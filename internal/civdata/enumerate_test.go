package civdata

import (
	"path/filepath"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/testutil"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

func TestBuildingPhase_GermFromReachData(t *testing.T) {
	testutil.SkipIfNoGameData(t)
	root := testutil.GameDataRoot()
	idx, err := tmpl.NewIndex(filepath.Join(root, "simulation/templates"))
	if err != nil {
		t.Fatalf("NewIndex: %v", err)
	}
	resolver := tmpl.NewResolver(idx)

	cases := []struct {
		basename string
		want     Phase
	}{
		{"civil_centre", PhaseVillage}, // village (always available)
		{"house", PhaseVillage},
		{"barracks", PhaseVillage},
		{"great_hall", PhaseTown},    // -phase_city phase_town → Town
		{"fortress", PhaseCity},      // template_structure_military_fortress: phase_city
		{"defense_tower", PhaseTown}, // template_structure_defensive_tower_stone: phase_town
		{"wonder", PhaseCity},        // template_structure_wonder: phase_city
	}
	for _, c := range cases {
		t.Run(c.basename, func(t *testing.T) {
			path, ok := idx.Lookup("structures/germ/" + c.basename)
			if !ok {
				t.Fatalf("lookup: structures/germ/%s not found", c.basename)
			}
			el, err := resolver.Resolve(path)
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			ent := Entity{TemplateID: "structures/germ/" + c.basename, Path: path, Element: el}
			got := BuildingPhase(ent)
			if got != c.want {
				t.Errorf("BuildingPhase(%s) = %v, want %v", c.basename, got, c.want)
			}
		})
	}
}
