package webui

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/events"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/webui/templates"
)

//go:embed static/*
var staticFS embed.FS

type handlers struct {
	repRoot string
}

func (h *handlers) index(w http.ResponseWriter, r *http.Request) {
	cards := h.loadAllCards()
	_ = templates.Index(cards).Render(r.Context(), w)
}

func (h *handlers) replay(w http.ResponseWriter, r *http.Request) {
	matchID := strings.TrimPrefix(r.URL.Path, "/replay/")
	a, dir, err := h.findByMatchID(matchID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = dir
	bo := buildOrderRows(a)
	an := anomalyRows(a)
	chartData := buildDensityChartData(a)
	phaseMarkers := buildPhaseMarkers(a)
	engMarkers := buildEngagementMarkers(a)
	_ = templates.Replay(a, bo, an, chartData, phaseMarkers, engMarkers).Render(r.Context(), w)
}

func (h *handlers) loadAllCards() []templates.ReplayCard {
	entries, _ := os.ReadDir(h.repRoot)
	var cards []templates.ReplayCard
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(h.repRoot, e.Name())
		if _, err := os.Stat(filepath.Join(dir, "metadata.json")); err != nil {
			continue
		}
		a, err := replay.Run(dir)
		if err != nil {
			continue
		}
		outcome := "—"
		if fs, ok := a.FinalState.Players[1]; ok {
			outcome = fs.Outcome
		}
		cards = append(cards, templates.ReplayCard{
			MatchID:  a.Game.MatchID,
			Map:      a.Game.Map,
			When:     time.Unix(a.Game.Timestamp, 0).Format("02 Jan 15:04"),
			Duration: fmt.Sprintf("%d мин", a.Game.DurationMs/60000),
			Players:  a.Players,
			Outcome:  outcome,
		})
	}
	sort.Slice(cards, func(i, j int) bool { return cards[i].When > cards[j].When })
	return cards
}

func (h *handlers) findByMatchID(matchID string) (*output.Analysis, string, error) {
	entries, _ := os.ReadDir(h.repRoot)
	for _, e := range entries {
		dir := filepath.Join(h.repRoot, e.Name())
		path := filepath.Join(dir, "analysis.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var a output.Analysis
		if err := json.Unmarshal(raw, &a); err != nil {
			continue
		}
		if a.Game.MatchID == matchID {
			return &a, dir, nil
		}
	}
	return nil, "", fmt.Errorf("not found")
}

func buildOrderRows(a *output.Analysis) []templates.BuildOrderRow {
	playerName := map[int]string{}
	for _, p := range a.Players {
		playerName[p.ID] = p.Name
	}
	var rows []templates.BuildOrderRow
	for _, e := range a.Events {
		var label string
		switch e.Type {
		case events.TypeResearch:
			if d, ok := e.Data.(events.ResearchData); ok {
				label = "research " + d.Template
			} else if m, ok := e.Data.(map[string]any); ok {
				if t, _ := m["template"].(string); t != "" {
					label = "research " + t
				}
			}
		case events.TypeConstruct:
			if d, ok := e.Data.(events.ConstructData); ok {
				label = "construct " + d.Template
			} else if m, ok := e.Data.(map[string]any); ok {
				if t, _ := m["template"].(string); t != "" {
					label = "construct " + t
				}
			}
		case events.TypeResign:
			label = "RESIGN"
		default:
			continue
		}
		rows = append(rows, templates.BuildOrderRow{
			Time:   fmt.Sprintf("%02d:%02d", e.T/60000, (e.T/1000)%60),
			Player: playerName[e.Player],
			Event:  label,
		})
	}
	return rows
}

func anomalyRows(a *output.Analysis) []templates.AnomalyRow {
	var rows []templates.AnomalyRow
	for _, m := range a.Metrics.Players {
		for _, an := range m.Anomalies {
			detail := ""
			if t, ok := an.Details["target"]; ok {
				detail = fmt.Sprintf("target=%v", t)
			}
			rows = append(rows, templates.AnomalyRow{
				Type:     an.Type,
				Severity: an.Severity,
				TStart:   an.TStartSec,
				TEnd:     an.TEndSec,
				Detail:   detail,
			})
		}
	}
	return rows
}

type densityTrace struct {
	Name string `json:"name"`
	X    []int  `json:"x"`
	Y    []int  `json:"y"`
}

type phaseMarker struct {
	X     int    `json:"x"`
	Label string `json:"label"`
}

type engMarker struct {
	X    int `json:"x"`
	Size int `json:"size"`
}

func buildDensityChartData(a *output.Analysis) []densityTrace {
	cats := []string{"military", "build", "research", "economy", "other"}
	x := make([]int, len(a.Metrics.Density))
	for i, b := range a.Metrics.Density {
		x[i] = b.TSec
	}
	out := make([]densityTrace, 0, len(cats))
	for _, c := range cats {
		y := make([]int, len(a.Metrics.Density))
		for i, b := range a.Metrics.Density {
			y[i] = b.Counts[c]
		}
		out = append(out, densityTrace{Name: c, X: x, Y: y})
	}
	return out
}

func buildPhaseMarkers(a *output.Analysis) []phaseMarker {
	var out []phaseMarker
	for _, m := range a.Metrics.Players {
		for name, t := range m.PhaseTimings {
			out = append(out, phaseMarker{X: t, Label: name})
		}
	}
	return out
}

func buildEngagementMarkers(a *output.Analysis) []engMarker {
	const minPeak = 5
	var out []engMarker
	for _, m := range a.Metrics.Players {
		for _, e := range m.Engagements {
			if e.PeakUnits < minPeak {
				continue
			}
			out = append(out, engMarker{X: e.TStartSec, Size: e.PeakUnits})
		}
	}
	return out
}
