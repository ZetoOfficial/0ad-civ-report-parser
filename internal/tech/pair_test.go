package tech

import (
	"path/filepath"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/testutil"
)

func TestExpandPair_Han(t *testing.T) {
	testutil.SkipIfNoGameData(t)

	c := NewCatalog(filepath.Join(testutil.GameDataRoot(), "simulation", "data", "technologies"))
	top, bottom, ok := ExpandPair(c, "pair_unlock_civil_service_han")
	if !ok {
		t.Fatal("ExpandPair returned ok=false, want true")
	}
	if top.Name != "civil_service_01" {
		t.Errorf("top.Name = %q, want %q", top.Name, "civil_service_01")
	}
	if bottom.Name != "civil_service_02" {
		t.Errorf("bottom.Name = %q, want %q", bottom.Name, "civil_service_02")
	}
}

func TestExpandPair_NotPair(t *testing.T) {
	testutil.SkipIfNoGameData(t)

	c := NewCatalog(filepath.Join(testutil.GameDataRoot(), "simulation", "data", "technologies"))
	top, bottom, ok := ExpandPair(c, "phase_town")
	if ok {
		t.Error("ExpandPair returned ok=true for non-pair tech, want false")
	}
	if top != nil || bottom != nil {
		t.Errorf("ExpandPair returned non-nil top/bottom for non-pair tech: top=%v bottom=%v", top, bottom)
	}
}

func TestExpandPair_Missing(t *testing.T) {
	testutil.SkipIfNoGameData(t)

	c := NewCatalog(filepath.Join(testutil.GameDataRoot(), "simulation", "data", "technologies"))
	top, bottom, ok := ExpandPair(c, "no_such_pair_xxx")
	if ok {
		t.Error("ExpandPair returned ok=true for missing tech, want false")
	}
	if top != nil || bottom != nil {
		t.Error("ExpandPair returned non-nil top/bottom for missing tech")
	}
}
