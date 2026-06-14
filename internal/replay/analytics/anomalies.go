package analytics

import (
	"sort"

	api "github.com/ZetoOfficial/0ad-civ-report-parser/internal/api/gen"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/events"
)

// PanicGarrison flags ≥5 garrison commands into the same building within < 3 sec.
func PanicGarrison(evs []events.Event) map[int][]api.Anomaly {
	type key struct {
		player int
		target int
	}
	type win struct {
		ts []int64
	}
	seen := map[key]*win{}
	out := map[int][]api.Anomaly{}
	sorted := append([]events.Event{}, evs...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].TMs < sorted[j].TMs })

	const windowMs = 3000
	const threshold = 5

	for _, e := range sorted {
		if e.Type != events.TypeGarrison {
			continue
		}
		d, ok := e.Data.(events.GarrisonData)
		if !ok {
			continue
		}
		k := key{player: e.Player, target: d.Target}
		w, ok2 := seen[k]
		if !ok2 {
			w = &win{}
			seen[k] = w
		}
		w.ts = append(w.ts, e.TMs)
		// trim window
		cut := 0
		for cut < len(w.ts) && w.ts[len(w.ts)-1]-w.ts[cut] > windowMs {
			cut++
		}
		w.ts = w.ts[cut:]
		if len(w.ts) >= threshold {
			details := map[string]any{
				"target": d.Target,
				"count":  len(w.ts),
			}
			out[e.Player] = append(out[e.Player], api.Anomaly{
				Type:      "panic_garrison",
				TStartSec: int(w.ts[0] / 1000),
				TEndSec:   int(w.ts[len(w.ts)-1] / 1000),
				Severity:  "warning",
				Details:   &details,
			})
			w.ts = w.ts[:0] // reset so we don't double-fire
		}
	}
	return out
}
