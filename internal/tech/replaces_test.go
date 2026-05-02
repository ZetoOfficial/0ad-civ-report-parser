package tech

import (
	"path/filepath"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/testutil"
)

func newTestIndex(t *testing.T) *Index {
	t.Helper()
	testutil.SkipIfNoGameData(t)
	c := NewCatalog(filepath.Join(testutil.GameDataRoot(), "simulation", "data", "technologies"))
	idx, err := NewIndex(c)
	if err != nil {
		t.Fatalf("NewIndex: %v", err)
	}
	return idx
}

func TestIndex_ResolveForCiv_Athen(t *testing.T) {
	idx := newTestIndex(t)
	got := idx.ResolveForCiv("phase_town", "athen")
	if got == nil {
		t.Fatal("ResolveForCiv returned nil for athen")
	}
	if got.Name != "phase_town_athen" {
		t.Errorf("got %q; want %q", got.Name, "phase_town_athen")
	}
}

func TestIndex_ResolveForCiv_Pers(t *testing.T) {
	idx := newTestIndex(t)
	got := idx.ResolveForCiv("phase_town", "pers")
	if got == nil {
		t.Fatal("ResolveForCiv returned nil for pers")
	}
	if got.Name != "phase_town_pers" {
		t.Errorf("got %q; want %q", got.Name, "phase_town_pers")
	}
}

func TestIndex_ResolveForCiv_Germ(t *testing.T) {
	idx := newTestIndex(t)
	got := idx.ResolveForCiv("phase_town", "germ")
	if got == nil {
		t.Fatal("ResolveForCiv returned nil for germ")
	}
	if got.Name != "phase_town_generic" {
		t.Errorf("got %q; want %q", got.Name, "phase_town_generic")
	}
}

func TestIndex_Chain_PhaseTownAthen(t *testing.T) {
	idx := newTestIndex(t)
	ci := idx.Chain("phase_town_athen")
	if len(ci.Replaces) != 1 || ci.Replaces[0] != "phase_town" {
		t.Errorf("Replaces = %v; want [phase_town]", ci.Replaces)
	}
	if ci.Supersedes != "phase_village" {
		t.Errorf("Supersedes = %q; want %q", ci.Supersedes, "phase_village")
	}
}

func TestIndex_Get_Literal(t *testing.T) {
	idx := newTestIndex(t)

	// Get returns literal tech, NOT civ-resolved.
	got := idx.Get("phase_town")
	if got == nil {
		t.Fatal("Get(phase_town) returned nil")
	}
	if got.Name != "phase_town" {
		t.Errorf("Get(phase_town).Name = %q, want %q", got.Name, "phase_town")
	}

	// Unknown name returns nil.
	if idx.Get("no_such_tech_xxx") != nil {
		t.Error("Get(no_such_tech_xxx) should return nil")
	}
}

func TestIndex_CivSpecific_Germ(t *testing.T) {
	idx := newTestIndex(t)

	list := idx.CivSpecific("germ")
	names := map[string]bool{}
	for _, tt := range list {
		names[tt.Name] = true
	}

	// Files in civbonuses/ should be present.
	for _, expected := range []string{"germ_meat", "germ_women"} {
		if !names[expected] {
			t.Errorf("CivSpecific(germ) missing civbonuses entry %q", expected)
		}
	}
	// Files in root with all→civ:germ should also be present.
	for _, expected := range []string{"resettlement", "grove_of_fetters"} {
		if !names[expected] {
			t.Errorf("CivSpecific(germ) missing root entry %q", expected)
		}
	}
}

func TestIndex_Chain_PhaseTown_ReplacedByContains(t *testing.T) {
	idx := newTestIndex(t)
	ci := idx.Chain("phase_town")
	must := []string{"phase_town_generic", "phase_town_athen", "phase_town_pers"}
	rb := make(map[string]bool, len(ci.ReplacedBy))
	for _, s := range ci.ReplacedBy {
		rb[s] = true
	}
	for _, want := range must {
		if !rb[want] {
			t.Errorf("ReplacedBy missing %q (got %v)", want, ci.ReplacedBy)
		}
	}
}
