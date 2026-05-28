package analytics

import (
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/events"
)

func mkResearch(player int, tMs int64, tmpl string) events.Event {
	return events.Event{TMs: tMs, Player: player, Type: events.TypeResearch, Data: events.ResearchData{Template: tmpl}}
}
func mkAttack(player int, tMs int64, target, units int) events.Event {
	return events.Event{TMs: tMs, Player: player, Type: events.TypeAttack, Data: events.AttackData{Target: target, Entities: units}}
}
func mkGarrison(player int, tMs int64, target int) events.Event {
	return events.Event{TMs: tMs, Player: player, Type: events.TypeGarrison, Data: events.GarrisonData{Target: target, Entities: 1}}
}

func TestPhaseTimings(t *testing.T) {
	evs := []events.Event{
		mkResearch(2, 295000, "phase_town_generic"),
		mkResearch(2, 1547000, "phase_city_generic"),
		mkResearch(3, 320000, "phase_town_athen"),
		mkResearch(2, 200000, "gather_wicker_baskets"), // ignored
	}
	out := PhaseTimings(evs)
	if out[2]["town"] != 295 {
		t.Errorf("p2 town = %d", out[2]["town"])
	}
	if out[2]["city"] != 1547 {
		t.Errorf("p2 city = %d", out[2]["city"])
	}
	if out[3]["town"] != 320 {
		t.Errorf("p3 town = %d", out[3]["town"])
	}
}

func TestEngagementsGroups(t *testing.T) {
	evs := []events.Event{
		mkAttack(2, 10000, 99, 5),
		mkAttack(2, 11000, 99, 10), // within 3s window
		mkAttack(2, 12500, 99, 12),
		mkAttack(2, 20000, 99, 6), // gap > 3s → new engagement
	}
	out := Engagements(evs, 3000)
	if len(out[2]) != 2 {
		t.Fatalf("got %d engagements, want 2: %+v", len(out[2]), out[2])
	}
	if out[2][0].PeakUnits != 12 || out[2][0].CommandCount != 3 {
		t.Errorf("eng[0] = %+v", out[2][0])
	}
}

func TestPanicGarrison(t *testing.T) {
	var evs []events.Event
	for i := int64(0); i < 5; i++ {
		evs = append(evs, mkGarrison(2, 10000+i*400, 77))
	}
	out := PanicGarrison(evs)
	if len(out[2]) != 1 {
		t.Fatalf("got %d anomalies, want 1: %+v", len(out[2]), out[2])
	}
	if out[2][0].Type != "panic_garrison" {
		t.Errorf("Type = %q", out[2][0].Type)
	}
}

func TestPanicGarrisonNoFalsePositive(t *testing.T) {
	evs := []events.Event{
		mkGarrison(2, 0, 77),
		mkGarrison(2, 4000, 77),
		mkGarrison(2, 8000, 77),
		mkGarrison(2, 12000, 77),
		mkGarrison(2, 16000, 77),
	}
	out := PanicGarrison(evs)
	if len(out[2]) != 0 {
		t.Fatalf("got %d, want 0 (gaps too large)", len(out[2]))
	}
}

func TestActionDensity(t *testing.T) {
	evs := []events.Event{
		mkResearch(2, 5000, "x"),                                                          // bin 0 research
		mkAttack(2, 15000, 1, 1),                                                          // bin 0 military
		mkResearch(2, 35000, "y"),                                                         // bin 1 research
		{TMs: 65000, Player: 2, Type: events.TypeConstruct, Data: events.ConstructData{}}, // bin 2 build
	}
	out := ActionDensity(evs, 30)
	if len(out) != 3 {
		t.Fatalf("bins = %d, want 3", len(out))
	}
	if out[0].Counts["research"] != 1 || out[0].Counts["military"] != 1 {
		t.Errorf("bin0 = %+v", out[0].Counts)
	}
	if out[2].Counts["build"] != 1 {
		t.Errorf("bin2 = %+v", out[2].Counts)
	}
}
