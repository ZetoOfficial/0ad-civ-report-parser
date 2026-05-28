// Package analytics derives per-player metrics from a typed event stream.
package analytics

import (
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/events"
)

// PhaseTimings: per player, "town" → seconds when phase_town_* was researched.
// Returns the earliest research time per phase.
func PhaseTimings(evs []events.Event) map[int]map[string]int {
	out := map[int]map[string]int{}
	for _, e := range evs {
		if e.Type != events.TypeResearch {
			continue
		}
		d, ok := e.Data.(events.ResearchData)
		if !ok {
			continue
		}
		var phase string
		switch {
		case strings.HasPrefix(d.Template, "phase_town"):
			phase = "town"
		case strings.HasPrefix(d.Template, "phase_city"):
			phase = "city"
		default:
			continue
		}
		if _, exists := out[e.Player]; !exists {
			out[e.Player] = map[string]int{}
		}
		sec := int(e.TMs / 1000)
		if prev, ok := out[e.Player][phase]; !ok || sec < prev {
			out[e.Player][phase] = sec
		}
	}
	return out
}
