package tech

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/testutil"
)

func TestModification_AffectsList_String(t *testing.T) {
	raw := `{"value":"Attack/Melee/Damage/Hack","multiply":1.1,"affects":"Melee"}`
	var m Modification
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := m.AffectsList()
	want := []string{"Melee"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("AffectsList = %v; want %v", got, want)
	}
}

func TestModification_AffectsList_Array(t *testing.T) {
	raw := `{"value":"Health/Max","multiply":1.25,"affects":["Melee","Cavalry"]}`
	var m Modification
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := m.AffectsList()
	want := []string{"Melee", "Cavalry"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("AffectsList = %v; want %v", got, want)
	}
}

func TestModification_AffectsList_Absent(t *testing.T) {
	raw := `{"value":"Health/Max","add":10}`
	var m Modification
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := m.AffectsList(); got != nil {
		t.Errorf("AffectsList = %v; want nil", got)
	}
}

func TestCatalog_LoadAll_Idempotent(t *testing.T) {
	testutil.SkipIfNoGameData(t)
	c := NewCatalog(filepath.Join(testutil.GameDataRoot(), "simulation/data/technologies"))
	if err := c.LoadAll(); err != nil {
		t.Fatalf("first LoadAll: %v", err)
	}
	first := c.AllLoaded()
	n1 := len(first)
	if n1 == 0 {
		t.Fatal("expected non-empty cache after LoadAll")
	}
	if err := c.LoadAll(); err != nil {
		t.Fatalf("second LoadAll: %v", err)
	}
	second := c.AllLoaded()
	if len(second) != n1 {
		t.Errorf("LoadAll not idempotent: first=%d, second=%d", n1, len(second))
	}
	// pointer identity check on a known sentinel: phase_town must point
	// to the same Technology object on both calls
	if t1, ok1 := first["phase_town"]; ok1 {
		if t2, ok2 := second["phase_town"]; !ok2 || t1 != t2 {
			t.Error("phase_town pointer not stable across LoadAll calls")
		}
	}
}
