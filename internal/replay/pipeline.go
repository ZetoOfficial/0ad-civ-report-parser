// Package replay orchestrates the per-replay pipeline.
package replay

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/analytics"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/commands"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/events"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/metadata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"
)

// Run parses replayDir and returns Analysis. It writes analysis.json next to
// the replay; if analysis.json is newer than commands.txt it is reused.
func Run(replayDir string) (*output.Analysis, error) {
	cmdsPath := filepath.Join(replayDir, "commands.txt")
	metaPath := filepath.Join(replayDir, "metadata.json")
	outPath := filepath.Join(replayDir, "analysis.json")

	if _, err := os.Stat(metaPath); err != nil {
		return nil, fmt.Errorf("replay: %s: metadata.json missing (skipping)", replayDir)
	}

	if output.IsFresh(outPath, cmdsPath) {
		raw, err := os.ReadFile(outPath)
		if err == nil {
			var a output.Analysis
			if err := json.Unmarshal(raw, &a); err == nil && a.SchemaVersion == output.SchemaVersion {
				return &a, nil
			}
		}
	}

	meta, err := metadata.Load(metaPath)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(cmdsPath)
	if err != nil {
		return nil, fmt.Errorf("replay: open commands.txt: %w", err)
	}
	defer f.Close()

	game, players, err := parseStart(f)
	if err != nil {
		return nil, err
	}
	// Re-open for full streaming (parseStart consumed only the first line)
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	evs, durationMs, err := streamEvents(f)
	if err != nil {
		return nil, err
	}
	game.DurationMs = durationMs
	if meta.TimeElapsed > durationMs {
		game.DurationMs = meta.TimeElapsed
	}

	a := buildAnalysis(game, players, meta, evs)

	if err := output.Write(outPath, a); err != nil {
		return nil, fmt.Errorf("replay: write analysis: %w", err)
	}
	return a, nil
}

func parseStart(r io.Reader) (output.GameInfo, []output.Player, error) {
	rd := commands.New(r)
	ln, ok, err := rd.Next()
	if err != nil {
		return output.GameInfo{}, nil, err
	}
	if !ok || ln.Kind != commands.KindStart {
		return output.GameInfo{}, nil, fmt.Errorf("replay: first line is not 'start'")
	}
	var s struct {
		Settings struct {
			MapName           string   `json:"mapName"`
			VictoryConditions []string `json:"VictoryConditions"`
			PlayerData        []struct {
				AI     any             `json:"AI"`
				AIDiff json.RawMessage `json:"AIDiff"` // int or quoted-string or null
				Civ    string          `json:"Civ"`
				Name   string          `json:"Name"`
				Team   int             `json:"Team"`
				Color  struct {
					R, G, B int
				} `json:"Color"`
			} `json:"PlayerData"`
		} `json:"settings"`
		MatchID   string `json:"matchID"`
		Map       string `json:"map"`
		MapType   string `json:"mapType"`
		Timestamp int64  `json:"timestamp"`
		Mods      []struct {
			Version string `json:"version"`
		} `json:"mods"`
	}
	if err := json.Unmarshal(ln.StartJSON, &s); err != nil {
		return output.GameInfo{}, nil, fmt.Errorf("replay: parse start: %w", err)
	}
	matchID := s.MatchID
	if matchID == "" {
		h := sha1.Sum(ln.StartJSON)
		matchID = hex.EncodeToString(h[:8])
	}
	ev := ""
	if len(s.Mods) > 0 {
		ev = s.Mods[0].Version
	}
	g := output.GameInfo{
		MatchID:       matchID,
		Map:           s.Settings.MapName,
		MapType:       s.MapType,
		Timestamp:     s.Timestamp,
		EngineVersion: ev,
		VictoryConds:  s.Settings.VictoryConditions,
	}
	players := make([]output.Player, 0, len(s.Settings.PlayerData))
	for i, pd := range s.Settings.PlayerData {
		_, isAI := pd.AI.(string)
		aiDiff := parseAIDiff(pd.AIDiff)
		players = append(players, output.Player{
			ID:     i + 1,
			Name:   pd.Name,
			Civ:    pd.Civ,
			Team:   pd.Team,
			IsAI:   isAI,
			AIDiff: aiDiff,
			Color:  output.Color{R: pd.Color.R, G: pd.Color.G, B: pd.Color.B},
		})
	}
	return g, players, nil
}

// parseAIDiff converts AIDiff JSON (int, quoted string, or null) to int.
func parseAIDiff(raw json.RawMessage) int {
	if len(raw) == 0 || string(raw) == "null" {
		return 0
	}
	// Try int first
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n
	}
	// Try quoted string "3" → 3
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		n, _ = strconv.Atoi(s)
		return n
	}
	return 0
}

func streamEvents(r io.Reader) ([]output.Event, int64, error) {
	rd := commands.New(r)
	var evs []output.Event
	var tMs int64
	for {
		ln, ok, err := rd.Next()
		if err != nil {
			return nil, 0, err
		}
		if !ok {
			break
		}
		switch ln.Kind {
		case commands.KindTurn:
			tMs += int64(ln.TickMs)
		case commands.KindCmd:
			ev := events.Decode(ln.Player, tMs, ln.CmdJSON)
			evs = append(evs, output.Event{
				T: ev.TMs, Player: ev.Player, Type: ev.Type, Data: ev.Data,
			})
		}
	}
	return evs, tMs, nil
}

// internalEvents reconstructs the typed events.Event slice for analytics.
func internalEvents(evs []output.Event) []events.Event {
	out := make([]events.Event, len(evs))
	for i, e := range evs {
		out[i] = events.Event{TMs: e.T, Player: e.Player, Type: e.Type, Data: e.Data}
	}
	return out
}

func buildAnalysis(g output.GameInfo, players []output.Player, m *metadata.Metadata, evs []output.Event) *output.Analysis {
	tev := internalEvents(evs)
	phaseT := analytics.PhaseTimings(tev)
	eng := analytics.Engagements(tev, 3000)
	pg := analytics.PanicGarrison(tev)
	density := analytics.ActionDensity(tev, 30)

	resignByPlayer := map[int]bool{}
	for _, e := range tev {
		if e.Type == events.TypeResign {
			resignByPlayer[e.Player] = true
		}
	}

	finalByPlayer := map[int]output.PlayerFinalState{}
	for i, ps := range m.PlayerStates {
		if i == 0 {
			continue // gaia
		}
		outcome := ps.State
		if resignByPlayer[i] {
			outcome = "defeated"
		}
		rc := output.Resources{
			Food:  ps.ResourceCounts["food"],
			Wood:  ps.ResourceCounts["wood"],
			Stone: ps.ResourceCounts["stone"],
			Metal: ps.ResourceCounts["metal"],
		}
		finalByPlayer[i] = output.PlayerFinalState{
			Phase:           ps.Phase,
			State:           ps.State,
			Outcome:         outcome,
			PopCount:        ps.PopCount,
			PopLimit:        ps.PopLimit,
			PopMax:          ps.PopMax,
			ResourceCounts:  rc,
			ResearchedTechs: ps.ResearchedTechs,
		}
	}

	metricsByPlayer := map[int]output.PlayerMetrics{}
	allPlayers := map[int]struct{}{}
	for p := range phaseT {
		allPlayers[p] = struct{}{}
	}
	for p := range eng {
		allPlayers[p] = struct{}{}
	}
	for p := range pg {
		allPlayers[p] = struct{}{}
	}
	for p := range allPlayers {
		metricsByPlayer[p] = output.PlayerMetrics{
			PhaseTimings: phaseT[p],
			Engagements:  eng[p],
			Anomalies:    pg[p],
		}
	}

	return &output.Analysis{
		SchemaVersion: output.SchemaVersion,
		Game:          g,
		Players:       players,
		Events:        evs,
		Snapshots:     []output.Snapshot{},
		FinalState:    output.FinalState{Players: finalByPlayer},
		Metrics:       output.Metrics{Players: metricsByPlayer, Density: density},
	}
}

// Outcome returns a human label for a player given final state and resign events.
func Outcome(ps metadata.PlayerState, resigned bool) string {
	switch {
	case resigned:
		return "defeated"
	case strings.EqualFold(ps.State, "won"):
		return "won"
	case strings.EqualFold(ps.State, "defeated"):
		return "defeated"
	default:
		return ps.State
	}
}
