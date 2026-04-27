package i18n

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
)

func TestDescribeModification_AffectsSuffix_String(t *testing.T) {
	raw := `{"value":"Attack/Melee/Damage/Hack","multiply":1.1,"affects":"Melee"}`
	var m tech.Modification
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := DescribeModification(m)
	if !strings.Contains(got, "(только Melee)") {
		t.Errorf("missing suffix in %q", got)
	}
	if !strings.Contains(got, "+10%") {
		t.Errorf("missing percent in %q", got)
	}
}

func TestDescribeModification_AffectsSuffix_Array(t *testing.T) {
	raw := `{"value":"Health/Max","multiply":1.25,"affects":["Melee","Cavalry"]}`
	var m tech.Modification
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := DescribeModification(m)
	if !strings.Contains(got, "(только Melee+Cavalry)") {
		t.Errorf("missing combined suffix in %q", got)
	}
}

func TestDescribeModification_NoAffects_NoSuffix(t *testing.T) {
	raw := `{"value":"Health/Max","multiply":1.25}`
	var m tech.Modification
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := DescribeModification(m)
	if strings.Contains(got, "только") {
		t.Errorf("unexpected suffix in %q", got)
	}
}

func TestTechDisplayName_NoSpecificName(t *testing.T) {
	te := &tech.Technology{Name: "x", GenericName: "Phase Town"}
	if got := TechDisplayName(te, "spart"); got != "Phase Town" {
		t.Errorf("got %q", got)
	}
}

func TestTechDisplayName_WithSpecificName(t *testing.T) {
	te := &tech.Technology{
		Name:        "phase_town",
		GenericName: "Town Phase",
		SpecificName: map[string]any{
			"spart": "Astiteia",
			"athen": "Astuteia",
		},
	}
	got := TechDisplayName(te, "spart")
	if got != "Town Phase (локально: Astiteia)" {
		t.Errorf("got %q", got)
	}
}

func TestTechDisplayName_OtherCivFallsBack(t *testing.T) {
	te := &tech.Technology{
		Name:         "phase_town",
		GenericName:  "Town Phase",
		SpecificName: map[string]any{"spart": "Astiteia"},
	}
	got := TechDisplayName(te, "germ") // not in specificName
	if got != "Town Phase" {
		t.Errorf("got %q", got)
	}
}

func TestDescribeModification_NegativeAdd(t *testing.T) {
	raw := `{"value":"Health/Max","add":-5}`
	var m tech.Modification
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := DescribeModification(m)
	if !strings.Contains(got, "−5") {
		t.Errorf("expected unicode minus and 5 in %q", got)
	}
}

func TestDescribeModification_ReplaceBranch(t *testing.T) {
	raw := `{"value":"Cost/Resources/food","replace":0,"affects":"Hero"}`
	var m tech.Modification
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := DescribeModification(m)
	if !strings.Contains(got, "= 0") {
		t.Errorf("expected `= 0` in %q", got)
	}
	if !strings.Contains(got, "(только Hero)") {
		t.Errorf("expected `(только Hero)` suffix in %q", got)
	}
}

func TestDescribeModification_AffectsArrayLong(t *testing.T) {
	raw := `{"value":"Attack/Melee/Damage/Hack","multiply":1.1,"affects":["Melee","Cavalry","Champion"]}`
	var m tech.Modification
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := DescribeModification(m)
	if !strings.Contains(got, "(только Melee+Cavalry+Champion)") {
		t.Errorf("expected three-way affects join in %q", got)
	}
}
