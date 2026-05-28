package metadata

import (
	"path/filepath"
	"testing"
)

func TestLoadSample(t *testing.T) {
	m, err := Load(filepath.Join("testdata", "sample.json"))
	if err != nil {
		t.Fatal(err)
	}
	if m.TimeElapsed != 1860400 {
		t.Errorf("TimeElapsed = %d", m.TimeElapsed)
	}
	if len(m.PlayerStates) != 2 {
		t.Fatalf("PlayerStates len = %d", len(m.PlayerStates))
	}
	p := m.PlayerStates[1]
	if p.Civ != "spart" || p.State != "defeated" || p.Phase != "town" {
		t.Errorf("p[1] = %+v", p)
	}
	if p.ResourceCounts["food"] != 420 {
		t.Errorf("food = %d", p.ResourceCounts["food"])
	}
	if len(p.ResearchedTechs) != 2 {
		t.Errorf("techs = %v", p.ResearchedTechs)
	}
}
