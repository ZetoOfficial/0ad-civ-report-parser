package render

import (
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
)

func TestExtractRawPhase(t *testing.T) {
	cases := []struct {
		name string
		req  tech.Requirements
		want string
	}{
		{"nil", nil, ""},
		{"tech_top", tech.Requirements{"tech": "phase_town"}, "phase_town"},
		{"all_with_tech", tech.Requirements{"all": []any{
			map[string]any{"tech": "phase_city"},
			map[string]any{"civ": "athen"},
		}}, "phase_city"},
		{"entity_only", tech.Requirements{"entity": map[string]any{"class": "Village", "number": 5}}, ""},
		{"non_phase_tech", tech.Requirements{"tech": "blacksmith_advanced"}, "blacksmith_advanced"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := extractRawPhase(c.req)
			if got != c.want {
				t.Errorf("extractRawPhase(%v) = %q, want %q", c.req, got, c.want)
			}
		})
	}
}

func TestPhaseLabelFromSupersedes(t *testing.T) {
	cases := []struct{ in, want string }{
		{"phase_village", "Town"},
		{"phase_town", "City"},
		{"phase_city", ""},
		{"", ""},
		{"unknown", ""},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := phaseLabelFromSupersedes(c.in)
			if got != c.want {
				t.Errorf("phaseLabelFromSupersedes(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestChainSuffix_Empty(t *testing.T) {
	if got := chainSuffix(nil, &tech.Technology{Name: "x"}, ""); got != "" {
		t.Errorf("chainSuffix(nil idx) = %q, want empty", got)
	}
}
