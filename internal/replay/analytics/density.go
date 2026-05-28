package analytics

import (
	"sort"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/events"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"
)

// ActionDensity groups events into binSec-second buckets, classified by category.
// Categories: military (train/attack/garrison), economy (gather/tribute/barter via raw),
// build (construct), research (research), other.
func ActionDensity(evs []events.Event, binSec int) []output.DensityBin {
	if binSec <= 0 {
		binSec = 30
	}
	classify := func(e events.Event) string {
		switch e.Type {
		case events.TypeTrain, events.TypeAttack, events.TypeGarrison:
			return "military"
		case events.TypeConstruct:
			return "build"
		case events.TypeResearch:
			return "research"
		case events.TypeTribute:
			return "economy"
		case events.TypeRaw:
			d, ok := e.Data.(events.RawData)
			if !ok {
				return "other"
			}
			switch d.Kind {
			case "gather", "returnresource", "barter", "setup-trade-route":
				return "economy"
			default:
				return "other"
			}
		default:
			return "other"
		}
	}
	bins := map[int]map[string]int{}
	maxBin := 0
	for _, e := range evs {
		b := int(e.TMs/1000) / binSec
		if b > maxBin {
			maxBin = b
		}
		if bins[b] == nil {
			bins[b] = map[string]int{}
		}
		bins[b][classify(e)]++
	}
	out := make([]output.DensityBin, 0, maxBin+1)
	for b := 0; b <= maxBin; b++ {
		c := bins[b]
		if c == nil {
			c = map[string]int{}
		}
		out = append(out, output.DensityBin{TSec: b * binSec, Counts: c})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TSec < out[j].TSec })
	return out
}
