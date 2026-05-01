package render

import (
	"strings"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

func TestStructree_Germ_TwoWallSets(t *testing.T) {
	skipIfNoGamedata(t)
	out := generateFor(t, "germ")
	n := strings.Count(out.Structree, "### Стены")
	if n != 2 {
		t.Errorf("expected 2 wallset headers for germ, got %d", n)
	}
}

func TestStructree_Han_HasPairMarker(t *testing.T) {
	skipIfNoGamedata(t)
	out := generateFor(t, "han")
	if !strings.Contains(out.Structree, "◐") {
		t.Errorf("expected pair marker ◐ in han structree")
	}
}

func TestStructree_Athen_PhaseTownAthen(t *testing.T) {
	skipIfNoGamedata(t)
	out := generateFor(t, "athen")
	hasName := strings.Contains(out.Structree, "phase_town_athen") ||
		strings.Contains(out.Structree, "Kōmopolis")
	if !hasName {
		t.Errorf("athen structree missing phase_town_athen / Kōmopolis row")
	}
	if strings.Contains(out.Structree, "phase_town_generic") {
		t.Errorf("athen structree must not include phase_town_generic")
	}
}

// generateFor builds a full Output for the named civ using real gamedata.
func generateFor(t *testing.T, civ string) Output {
	t.Helper()
	layout := paths.Layout{Root: gamedataRoot()}
	resolver := newResolver(t)
	g := NewGenerator(layout, resolver)
	info, ok := civdata.ResolveCivInput(civ)
	if !ok {
		t.Fatalf("%s resolution failed", civ)
	}
	out, err := g.Generate(info)
	if err != nil {
		t.Fatalf("generate %s: %v", civ, err)
	}
	return out
}
