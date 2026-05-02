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

func TestAllowsCiv(t *testing.T) {
	cases := []struct {
		name string
		req  Requirements
		civ  string
		want bool
	}{
		{"empty", nil, "germ", true},
		{"top_civ_match", Requirements{"civ": "germ"}, "germ", true},
		{"top_civ_mismatch", Requirements{"civ": "rome"}, "germ", false},
		{"top_notciv_block", Requirements{"notciv": []any{"germ"}}, "germ", false},
		{"top_notciv_pass", Requirements{"notciv": []any{"rome"}}, "germ", true},
		{"all_with_civ_block", Requirements{"all": []any{
			map[string]any{"tech": "phase_town"},
			map[string]any{"civ": "rome"},
		}}, "germ", false},
		{"all_with_civ_pass", Requirements{"all": []any{
			map[string]any{"tech": "phase_town"},
			map[string]any{"civ": "germ"},
		}}, "germ", true},
		{"any_civ_match", Requirements{"any": []any{
			map[string]any{"civ": "athen"},
			map[string]any{"civ": "germ"},
		}}, "germ", true},
		{"any_civ_no_match", Requirements{"any": []any{
			map[string]any{"civ": "athen"},
			map[string]any{"civ": "spart"},
		}}, "germ", false},
		{"all_with_any_inner_block", Requirements{"all": []any{
			map[string]any{"tech": "phase_village"},
			map[string]any{"any": []any{
				map[string]any{"civ": "kush"},
				map[string]any{"civ": "maur"},
				map[string]any{"civ": "pers"},
			}},
		}}, "germ", false},
		{"all_with_any_inner_pass", Requirements{"all": []any{
			map[string]any{"tech": "phase_village"},
			map[string]any{"any": []any{
				map[string]any{"civ": "kush"},
				map[string]any{"civ": "germ"},
			}},
		}}, "germ", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := AllowsCiv(c.req, c.civ)
			if got != c.want {
				t.Errorf("AllowsCiv(%v, %q) = %v, want %v", c.req, c.civ, got, c.want)
			}
		})
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
