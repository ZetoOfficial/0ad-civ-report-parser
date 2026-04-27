package render

import (
	"path/filepath"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/testutil"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

func gamedataRoot() string          { return testutil.GameDataRoot() }
func skipIfNoGamedata(t *testing.T) { testutil.SkipIfNoGameData(t) }

func newResolver(t *testing.T) *tmpl.Resolver {
	t.Helper()
	idx, err := tmpl.NewIndex(filepath.Join(gamedataRoot(), "simulation/templates"))
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	return tmpl.NewResolver(idx)
}

type numericCheck struct {
	civ      string
	template string
	field    string
	want     float64
}

func TestNumericSpotChecks(t *testing.T) {
	skipIfNoGamedata(t)
	r := newResolver(t)

	cases := []numericCheck{
		{"spart", "units/spart/infantry_spearman_b", "Health/Max", 100},
		{"spart", "units/spart/infantry_spearman_b", "Attack/Melee/Damage/Hack", 4.5},
		{"spart", "units/spart/infantry_spearman_b", "Cost/Resources/wood", 50},
		{"spart", "units/spart/infantry_spearman_b", "Cost/Resources/food", 50},
		{"spart", "structures/spart/civil_centre", "Health/Max", 3000},
		{"spart", "structures/spart/civil_centre", "TerritoryInfluence/Radius", 140},
	}
	for _, tc := range cases {
		path := filepath.Join(testutil.GameDataRoot(), "simulation/templates", tc.template+".xml")
		e, err := r.Resolve(path)
		if err != nil {
			t.Errorf("%s: resolve: %v", tc.template, err)
			continue
		}
		got, ok := e.GetFloat(tc.field)
		if !ok {
			t.Errorf("%s/%s: not found", tc.template, tc.field)
			continue
		}
		if got != tc.want {
			t.Errorf("%s/%s = %v; want %v", tc.template, tc.field, got, tc.want)
		}
	}
}
