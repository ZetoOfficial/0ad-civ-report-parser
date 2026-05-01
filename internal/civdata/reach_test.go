package civdata

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/testutil"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

func newReachFixtures(t *testing.T) (*tmpl.Index, *tmpl.Resolver, *tech.Catalog) {
	t.Helper()
	root := testutil.GameDataRoot()
	idx, err := tmpl.NewIndex(filepath.Join(root, "simulation/templates"))
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	resolver := tmpl.NewResolver(idx)
	layout := paths.Layout{Root: root}
	catalog := tech.NewCatalog(layout.Technologies())
	return idx, resolver, catalog
}

func TestReach_Spart(t *testing.T) {
	testutil.SkipIfNoGameData(t)

	layout := paths.Layout{Root: testutil.GameDataRoot()}
	civ, err := LoadCiv(layout.CivJSON("spart"))
	if err != nil {
		t.Fatalf("LoadCiv: %v", err)
	}

	idx, resolver, catalog := newReachFixtures(t)
	res, err := Reach(civ, idx, resolver, catalog)
	if err != nil {
		t.Fatalf("Reach: %v", err)
	}

	// Buildings: at least 10
	if len(res.Buildings) < 10 {
		t.Errorf("Buildings count = %d; want >= 10", len(res.Buildings))
	}

	// Expected building basenames
	requiredBuildings := []string{
		"civil_centre",
		"wonder",
		"barracks",
		"forge",
		"wallset_stone",
		"wallset_palisade",
		"wall_tower",
	}
	buildingBasenames := make(map[string]struct{}, len(res.Buildings))
	for _, b := range res.Buildings {
		buildingBasenames[b.Basename()] = struct{}{}
	}
	for _, want := range requiredBuildings {
		if _, ok := buildingBasenames[want]; !ok {
			t.Errorf("Buildings: missing %q; got basenames: %v", want, sortedKeys(buildingBasenames))
		}
	}

	// Units: at least 5
	if len(res.Units) < 5 {
		t.Errorf("Units count = %d; want >= 5", len(res.Units))
	}

	// Expected unit basenames
	requiredUnits := []string{
		"support_civilian",
		"infantry_spearman_b",
	}
	unitBasenames := make(map[string]struct{}, len(res.Units))
	for _, u := range res.Units {
		unitBasenames[u.Basename()] = struct{}{}
	}
	for _, want := range requiredUnits {
		if _, ok := unitBasenames[want]; !ok {
			t.Errorf("Units: missing %q; got basenames: %v", want, sortedKeys(unitBasenames))
		}
	}

	// Techs: at least one
	if len(res.Techs) == 0 {
		t.Error("Techs: want > 0 entries")
	}

	// No pair_ prefix keys in Techs — all pairs must have been expanded
	for name := range res.Techs {
		if strings.HasPrefix(name, "pair_") {
			t.Errorf("Techs: found unexpanded pair key %q", name)
		}
	}
}

func TestReach_Idempotent(t *testing.T) {
	testutil.SkipIfNoGameData(t)

	layout := paths.Layout{Root: testutil.GameDataRoot()}
	civ, err := LoadCiv(layout.CivJSON("spart"))
	if err != nil {
		t.Fatalf("LoadCiv: %v", err)
	}

	// Use fresh resolver+catalog for each call so caches don't interfere.
	idx1, resolver1, catalog1 := newReachFixtures(t)
	res1, err := Reach(civ, idx1, resolver1, catalog1)
	if err != nil {
		t.Fatalf("Reach (first): %v", err)
	}

	idx2, resolver2, catalog2 := newReachFixtures(t)
	res2, err := Reach(civ, idx2, resolver2, catalog2)
	if err != nil {
		t.Fatalf("Reach (second): %v", err)
	}

	if len(res1.Buildings) != len(res2.Buildings) {
		t.Errorf("Buildings: %d vs %d", len(res1.Buildings), len(res2.Buildings))
	}
	if len(res1.Units) != len(res2.Units) {
		t.Errorf("Units: %d vs %d", len(res1.Units), len(res2.Units))
	}
	if len(res1.Techs) != len(res2.Techs) {
		t.Errorf("Techs: %d vs %d", len(res1.Techs), len(res2.Techs))
	}

	// Same-instance call must yield identical counts (cache idempotency).
	res3, err := Reach(civ, idx1, resolver1, catalog1)
	if err != nil {
		t.Fatalf("Reach (same instance, third call): %v", err)
	}
	if len(res3.Buildings) != len(res1.Buildings) {
		t.Errorf("same-instance Buildings differs: first=%d third=%d",
			len(res1.Buildings), len(res3.Buildings))
	}
	if len(res3.Units) != len(res1.Units) {
		t.Errorf("same-instance Units differs: first=%d third=%d",
			len(res1.Units), len(res3.Units))
	}
	if len(res3.Techs) != len(res1.Techs) {
		t.Errorf("same-instance Techs differs: first=%d third=%d",
			len(res1.Techs), len(res3.Techs))
	}
}

// sortedKeys returns the keys of a string-keyed map in sorted order,
// for use in diagnostic messages.
func sortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
