package tech

import (
	"strings"
	"testing"
)

func TestDescribeRequirements_Tech(t *testing.T) {
	got := DescribeRequirements(Requirements{"tech": "phase_town"})
	if got != "технология: phase_town" {
		t.Errorf("got %q", got)
	}
}

func TestDescribeRequirements_Entity(t *testing.T) {
	req := Requirements{
		"entity": map[string]any{"class": "Village", "number": float64(5)},
	}
	got := DescribeRequirements(req)
	want := "5+ зданий класса Village"
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

func TestDescribeRequirements_All(t *testing.T) {
	req := Requirements{
		"all": []any{
			map[string]any{"tech": "phase_town"},
			map[string]any{"entity": map[string]any{"class": "Village", "number": float64(5)}},
		},
	}
	got := DescribeRequirements(req)
	if !strings.Contains(got, "phase_town") || !strings.Contains(got, "Village") {
		t.Errorf("missing branches in %q", got)
	}
	if !strings.Contains(got, " И ") {
		t.Errorf("missing AND separator in %q", got)
	}
}

func TestDescribeRequirements_Any(t *testing.T) {
	req := Requirements{
		"any": []any{
			map[string]any{"tech": "tech_a"},
			map[string]any{"tech": "tech_b"},
		},
	}
	got := DescribeRequirements(req)
	if !strings.Contains(got, " ИЛИ ") {
		t.Errorf("missing OR separator in %q", got)
	}
}

func TestDescribeRequirements_NotCiv(t *testing.T) {
	req := Requirements{"notciv": []any{"spart", "athen"}}
	got := DescribeRequirements(req)
	if !strings.Contains(got, "не для цив") {
		t.Errorf("missing notciv prefix in %q", got)
	}
	if !strings.Contains(got, "spart") || !strings.Contains(got, "athen") {
		t.Errorf("missing civs in %q", got)
	}
}

func TestDescribeRequirements_Empty(t *testing.T) {
	if got := DescribeRequirements(nil); got != "" {
		t.Errorf("nil → %q; want empty", got)
	}
	if got := DescribeRequirements(Requirements{}); got != "" {
		t.Errorf("empty map → %q; want empty", got)
	}
}

func TestDescribeRequirements_Civ(t *testing.T) {
	got := DescribeRequirements(Requirements{"civ": "spart"})
	if got != "цивилизация: spart" {
		t.Errorf("got %q; want %q", got, "цивилизация: spart")
	}
}
