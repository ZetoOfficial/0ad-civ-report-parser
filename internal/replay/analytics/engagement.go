package analytics

import (
	"sort"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/events"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"
)

// Engagements groups consecutive attack commands on the same target by the same
// player; a new engagement starts when the gap to the previous attack exceeds windowMs.
func Engagements(evs []events.Event, windowMs int64) map[int][]output.Engagement {
	type key struct {
		player int
		target int
	}
	type cur struct {
		tStart, tEnd int64
		peak, count  int
	}
	open := map[key]*cur{}
	out := map[int][]output.Engagement{}

	flush := func(k key, c *cur) {
		out[k.player] = append(out[k.player], output.Engagement{
			TStartSec:    int(c.tStart / 1000),
			TEndSec:      int(c.tEnd / 1000),
			Target:       k.target,
			PeakUnits:    c.peak,
			CommandCount: c.count,
		})
	}

	// Stable order
	sorted := append([]events.Event{}, evs...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].TMs < sorted[j].TMs })

	for _, e := range sorted {
		if e.Type != events.TypeAttack {
			continue
		}
		d, ok := e.Data.(events.AttackData)
		if !ok {
			continue
		}
		k := key{player: e.Player, target: d.Target}
		c, exists := open[k]
		if exists && e.TMs-c.tEnd > windowMs {
			flush(k, c)
			exists = false
		}
		if !exists {
			c = &cur{tStart: e.TMs, tEnd: e.TMs, peak: d.Entities, count: 1}
			open[k] = c
			continue
		}
		c.tEnd = e.TMs
		c.count++
		if d.Entities > c.peak {
			c.peak = d.Entities
		}
	}
	for k, c := range open {
		flush(k, c)
	}
	for p := range out {
		es := out[p]
		sort.Slice(es, func(i, j int) bool { return es[i].TStartSec < es[j].TStartSec })
		out[p] = es
	}
	return out
}
