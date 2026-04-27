package tech

import (
	"path/filepath"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/testutil"
)

func TestGlobalAutoResearch_R28Set(t *testing.T) {
	testutil.SkipIfNoGameData(t)

	c := NewCatalog(filepath.Join(testutil.GameDataRoot(), "simulation", "data", "technologies"))
	got, err := c.GlobalAutoResearch()
	if err != nil {
		t.Fatalf("GlobalAutoResearch: %v", err)
	}
	gotNames := map[string]bool{}
	for _, te := range got {
		gotNames[te.Name] = true
	}
	mustContain := []string{
		"unit_advanced",
		"unit_elite",
		"phase_village",
		"soldier_ranged_experience",
		"upgrade_rank_advanced_mercenary",
	}
	for _, name := range mustContain {
		if !gotNames[name] {
			t.Errorf("GlobalAutoResearch missing %q (got %v)", name, mapKeys(gotNames))
		}
	}
	mustExclude := []string{
		"germ_meat",      // civbonuses/ subdir, civ-specific (requirements.civ=germ)
		"maur_elephants", // civbonuses/ subdir, civ-specific (requirements.civ=maur)
	}
	for _, name := range mustExclude {
		if gotNames[name] {
			t.Errorf("GlobalAutoResearch unexpectedly contains %q", name)
		}
	}
}

func mapKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
