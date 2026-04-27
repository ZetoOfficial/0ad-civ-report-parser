package tmpl

import (
	"path/filepath"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/testutil"
)

func gamedataRoot() string          { return testutil.GameDataRoot() }
func skipIfNoGamedata(t *testing.T) { testutil.SkipIfNoGameData(t) }

func newResolver(t *testing.T) *Resolver {
	t.Helper()
	skipIfNoGamedata(t)
	idx, err := NewIndex(filepath.Join(gamedataRoot(), "simulation/templates"))
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	return NewResolver(idx)
}

func TestResolveSpartanSpearman(t *testing.T) {
	r := newResolver(t)
	path := filepath.Join(gamedataRoot(), "simulation/templates/units/spart/infantry_spearman_b.xml")
	e, err := r.Resolve(path)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	hp, ok := e.GetFloat("Health/Max")
	if !ok || hp != 100 {
		t.Errorf("Health/Max = %v (ok=%v); want 100", hp, ok)
	}

	hack, ok := e.GetFloat("Attack/Melee/Damage/Hack")
	if !ok || hack != 4.5 {
		t.Errorf("Attack/Melee/Damage/Hack = %v (ok=%v); want 4.5", hack, ok)
	}

	pierce, ok := e.GetFloat("Attack/Melee/Damage/Pierce")
	if !ok || pierce != 4 {
		t.Errorf("Attack/Melee/Damage/Pierce = %v (ok=%v); want 4", pierce, ok)
	}

	wood, ok := e.GetFloat("Cost/Resources/wood")
	if !ok || wood != 50 {
		t.Errorf("Cost/Resources/wood = %v (ok=%v); want 50", wood, ok)
	}

	formations := e.GetTokens("UnitAI/Formations")
	if !contains(formations, "special/formations/phalanx") {
		t.Errorf("hoplite mixin not applied; formations=%v", formations)
	}

	visibleClasses := e.GetTokens("Identity/VisibleClasses")
	if !contains(visibleClasses, "Spearman") {
		t.Errorf("VisibleClasses missing Spearman; got %v", visibleClasses)
	}

	civ := e.GetText("Identity/Civ")
	if civ != "spart" {
		t.Errorf("Identity/Civ = %q; want spart", civ)
	}
}

func TestResolveCavalryWalkSpeedMul(t *testing.T) {
	r := newResolver(t)
	path := filepath.Join(gamedataRoot(), "simulation/templates/template_unit_cavalry.xml")
	e, err := r.Resolve(path)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	walk, ok := e.GetFloat("UnitMotion/WalkSpeed")
	if !ok {
		t.Fatalf("WalkSpeed not found")
	}
	if walk != 18 {
		t.Errorf("WalkSpeed (cavalry, op=mul 2 over base 9) = %v; want 18", walk)
	}
}

func TestResolveTokensRemoval(t *testing.T) {
	r := newResolver(t)
	path := filepath.Join(gamedataRoot(), "simulation/templates/units/spart/infantry_spearman_b.xml")
	e, err := r.Resolve(path)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	builders := e.GetTokens("Builder/Entities")
	for _, b := range builders {
		if b == "structures/{civ}/wallset_stone" {
			t.Errorf("wallset_stone should have been removed via -prefix; got %v", builders)
		}
	}
	if !contains(builders, "structures/spart/gerousia") {
		t.Errorf("expected gerousia in builders; got %v", builders)
	}
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
