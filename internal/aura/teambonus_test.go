package aura

import (
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/testutil"
)

func TestLoadTeamBonus_Spart(t *testing.T) {
	testutil.SkipIfNoGameData(t)

	layout := paths.Layout{Root: testutil.GameDataRoot()}
	a, err := LoadTeamBonus(layout, "spart")
	if err != nil {
		t.Fatalf("LoadTeamBonus: %v", err)
	}
	if a.AuraName != "Peloponnesian League" {
		t.Errorf("AuraName = %q; want %q", a.AuraName, "Peloponnesian League")
	}
	if a.Type != "global" {
		t.Errorf("Type = %q; want %q", a.Type, "global")
	}
	// Spartans team bonus = heroes free across 4 resources.
	if len(a.Modifications) != 4 {
		t.Errorf("Modifications len = %d; want 4", len(a.Modifications))
	}
	for _, m := range a.Modifications {
		if m.Replace == nil {
			t.Errorf("modification %s: expected Replace to be set", m.Value)
		}
	}
}
