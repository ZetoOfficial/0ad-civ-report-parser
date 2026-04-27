package tech

import (
	"encoding/json"
	"reflect"
	"testing"
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
