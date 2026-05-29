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
	"sort"
	"strconv"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/analytics"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/commands"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/events"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/metadata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/sandbox"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/techlib"
)

// Run parses replayDir and returns Analysis. It writes analysis.json next to
// the replay; if analysis.json is newer than commands.txt it is reused.
// lib may be nil; when nil, research events are still recorded but without
// human-readable metadata.
func Run(replayDir string, lib *techlib.Lib) (*output.Analysis, error) {
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
	game.DurationMs = max(durationMs, meta.TimeElapsed)

	a := buildAnalysis(game, players, meta, evs, filepath.Base(replayDir), lib)

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

func buildAnalysis(g output.GameInfo, players []output.Player, m *metadata.Metadata, evs []output.Event, replayBasename string, lib *techlib.Lib) *output.Analysis {
	tev := internalEvents(evs)
	phaseT := analytics.PhaseTimings(tev)
	eng := analytics.Engagements(tev, 3000)
	pg := analytics.PanicGarrison(tev)
	// Optional: sequences from headless-replay sandbox. nil if not yet regenerated.
	seqs, sandboxErr := sandbox.Load(replayBasename)
	if sandboxErr != nil {
		fmt.Fprintf(os.Stderr, "replay: sandbox load %s: %v\n", replayBasename, sandboxErr)
	}

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

	// Build per-player improvements from research events.
	improvsByPlayer := map[int][]output.ImprovementEntry{}
	for _, e := range tev {
		if e.Type != events.TypeResearch {
			continue
		}
		data, ok := e.Data.(events.ResearchData)
		if !ok {
			continue
		}
		entry := output.ImprovementEntry{
			TMs:      e.TMs,
			Template: data.Template,
		}
		if lib != nil {
			if info := lib.Resolve(data.Template); info != nil {
				entry.GenericName = info.GenericName
				entry.Description = info.Description
				entry.AutoResearch = info.AutoResearch
				entry.ResearchTime = info.ResearchTime
				entry.Cost = output.ImprovementCost{
					Food:  info.Cost.Food,
					Wood:  info.Cost.Wood,
					Stone: info.Cost.Stone,
					Metal: info.Cost.Metal,
				}
				if len(info.Buildings) > 0 {
					entry.Building = info.Buildings[0]
					if len(info.Buildings) > 1 {
						entry.Buildings = info.Buildings
					}
				}
			}
		}
		improvsByPlayer[e.Player] = append(improvsByPlayer[e.Player], entry)
	}
	// Sort each player's improvements by time ascending (events arrive in order
	// but explicit sort ensures correctness).
	for p := range improvsByPlayer {
		sort.Slice(improvsByPlayer[p], func(i, j int) bool {
			return improvsByPlayer[p][i].TMs < improvsByPlayer[p][j].TMs
		})
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
	for p := range improvsByPlayer {
		allPlayers[p] = struct{}{}
	}
	for p := range allPlayers {
		pt := phaseT[p]
		if pt == nil {
			pt = map[string]int{}
		}
		es := eng[p]
		if es == nil {
			es = []output.Engagement{}
		}
		an := pg[p]
		if an == nil {
			an = []output.Anomaly{}
		}
		imps := improvsByPlayer[p]
		if imps == nil {
			imps = []output.ImprovementEntry{}
		}
		metricsByPlayer[p] = output.PlayerMetrics{
			PhaseTimings: pt,
			Engagements:  es,
			Anomalies:    an,
			Sequences:    seqs[p],
			Improvements: imps,
		}
	}
	// Also surface players that only appear in sequences (e.g. allies who
	// issued no commands but show up in metadata.json).
	for p := range seqs {
		if _, ok := metricsByPlayer[p]; ok {
			continue
		}
		metricsByPlayer[p] = output.PlayerMetrics{
			PhaseTimings: map[string]int{},
			Engagements:  []output.Engagement{},
			Anomalies:    []output.Anomaly{},
			Sequences:    seqs[p],
			Improvements: []output.ImprovementEntry{},
		}
	}

	return &output.Analysis{
		SchemaVersion: output.SchemaVersion,
		Game:          g,
		Players:       players,
		Events:        evs,
		Snapshots:     []output.Snapshot{},
		FinalState:    output.FinalState{Players: finalByPlayer},
		Metrics:       output.Metrics{Players: metricsByPlayer},
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
