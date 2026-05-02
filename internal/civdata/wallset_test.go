package civdata

import (
	"path/filepath"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/testutil"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

func TestIdentifyWallSets_Spart(t *testing.T) {
	testutil.SkipIfNoGameData(t)

	root := testutil.GameDataRoot()
	layout := paths.Layout{Root: root}

	civ, err := LoadCiv(layout.CivJSON("spart"))
	if err != nil {
		t.Fatalf("LoadCiv: %v", err)
	}

	idx, err := tmpl.NewIndex(filepath.Join(root, "simulation/templates"))
	if err != nil {
		t.Fatalf("NewIndex: %v", err)
	}
	resolver := tmpl.NewResolver(idx)
	catalog := tech.NewCatalog(layout.Technologies())

	res, err := Reach(civ, idx, resolver, catalog)
	if err != nil {
		t.Fatalf("Reach: %v", err)
	}

	// There must be exactly 2 wallset groups for spartans.
	if len(res.WallSets) != 2 {
		t.Errorf("WallSets count = %d; want 2", len(res.WallSets))
	}

	// Build a map of wrapper basename → group for easy lookup.
	byWrapper := make(map[string]*WallSetGroup, len(res.WallSets))
	for _, g := range res.WallSets {
		byWrapper[g.Wrapper.Basename()] = g
	}

	// Both expected wrappers must be present.
	for _, wantWrapper := range []string{"wallset_stone", "wallset_palisade"} {
		if _, ok := byWrapper[wantWrapper]; !ok {
			t.Errorf("WallSets: missing wrapper %q; got %v", wantWrapper, wallsetWrapperNames(res.WallSets))
		}
	}

	// stone wallset: exactly 5 pieces with the required roles.
	if stone, ok := byWrapper["wallset_stone"]; ok {
		if len(stone.Pieces) != 5 {
			t.Errorf("wallset_stone pieces count = %d; want 5", len(stone.Pieces))
		}
		wantRoles := map[string]bool{
			"Tower": true, "Gate": true, "WallLong": true, "WallMedium": true, "WallShort": true,
		}
		gotRoles := make(map[string]bool, len(stone.Pieces))
		for _, p := range stone.Pieces {
			gotRoles[p.Role] = true
		}
		for role := range wantRoles {
			if !gotRoles[role] {
				t.Errorf("wallset_stone: missing role %q; got %v", role, mapKeys(gotRoles))
			}
		}
	}

	// palisade wallset: at least 5 pieces.
	if palisade, ok := byWrapper["wallset_palisade"]; ok {
		if len(palisade.Pieces) < 5 {
			t.Errorf("wallset_palisade pieces count = %d; want >= 5", len(palisade.Pieces))
		}
	}

	// Buildings must not contain wallset wrappers or their pieces.
	bannedBasenames := []string{
		"wallset_stone", "wallset_palisade",
		"wall_short", "wall_medium", "wall_long", "wall_gate", "wall_tower",
	}
	buildingBasenames := make(map[string]struct{}, len(res.Buildings))
	for _, b := range res.Buildings {
		buildingBasenames[b.Basename()] = struct{}{}
	}
	for _, banned := range bannedBasenames {
		if _, ok := buildingBasenames[banned]; ok {
			t.Errorf("Buildings: %q should have been moved to WallSets", banned)
		}
	}
}

func wallsetWrapperNames(groups []*WallSetGroup) []string {
	names := make([]string, len(groups))
	for i, g := range groups {
		names[i] = g.Wrapper.Basename()
	}
	return names
}

func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
