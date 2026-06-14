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

	api "github.com/ZetoOfficial/0ad-civ-report-parser/internal/api/gen"
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
func Run(replayDir string, lib *techlib.Lib) (*api.Analysis, error) {
	cmdsPath := filepath.Join(replayDir, "commands.txt")
	metaPath := filepath.Join(replayDir, "metadata.json")
	outPath := filepath.Join(replayDir, "analysis.json")

	if _, err := os.Stat(metaPath); err != nil {
		return nil, fmt.Errorf("replay: %s: metadata.json missing (skipping)", replayDir)
	}

	sandboxMeta := sandbox.MetadataPath(sandbox.DefaultRoot(), filepath.Base(replayDir))
	if output.IsFresh(outPath, cmdsPath, metaPath, sandboxMeta) {
		raw, err := os.ReadFile(outPath)
		if err == nil {
			var a api.Analysis
			if err := json.Unmarshal(raw, &a); err == nil && a.SchemaVersion == output.SchemaVersion {
				// Cached entry is fresh by mtime, but if it predates the
				// sequences-from-source-metadata fix it has empty metrics.
				// Force regen when sequences are missing but metadata has them.
				if hasAnySequences(&a) || !sourceHasSequences(metaPath) {
					return &a, nil
				}
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

	a := buildAnalysis(game, players, meta, evs, replayDir, lib)

	if err := output.Write(outPath, a); err != nil {
		return nil, fmt.Errorf("replay: write analysis: %w", err)
	}
	return a, nil
}

func parseStart(r io.Reader) (api.GameInfo, []api.Player, error) {
	rd := commands.New(r)
	ln, ok, err := rd.Next()
	if err != nil {
		return api.GameInfo{}, nil, err
	}
	if !ok || ln.Kind != commands.KindStart {
		return api.GameInfo{}, nil, fmt.Errorf("replay: first line is not 'start'")
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
		return api.GameInfo{}, nil, fmt.Errorf("replay: parse start: %w", err)
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
	g := api.GameInfo{
		MatchId:           matchID,
		Map:               s.Settings.MapName,
		MapType:           s.MapType,
		Timestamp:         s.Timestamp,
		EngineVersion:     ev,
		VictoryConditions: s.Settings.VictoryConditions,
	}
	players := make([]api.Player, 0, len(s.Settings.PlayerData))
	for i, pd := range s.Settings.PlayerData {
		_, isAI := pd.AI.(string)
		aiDiff := parseAIDiff(pd.AIDiff)
		p := api.Player{
			Id:    i + 1,
			Name:  pd.Name,
			Civ:   pd.Civ,
			Team:  pd.Team,
			IsAi:  isAI,
			Color: api.Color{R: pd.Color.R, G: pd.Color.G, B: pd.Color.B},
		}
		if aiDiff != 0 {
			p.AiDiff = &aiDiff
		}
		players = append(players, p)
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

func streamEvents(r io.Reader) ([]api.Event, int64, error) {
	rd := commands.New(r)
	var evs []api.Event
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
			evs = append(evs, api.Event{
				T: ev.TMs, Player: ev.Player, Type: ev.Type, Data: ev.Data,
			})
		}
	}
	return evs, tMs, nil
}

// internalEvents reconstructs the typed events.Event slice for analytics.
func internalEvents(evs []api.Event) []events.Event {
	out := make([]events.Event, len(evs))
	for i, e := range evs {
		out[i] = events.Event{TMs: e.T, Player: e.Player, Type: e.Type, Data: e.Data}
	}
	return out
}

// strKey converts an int player ID to the string key used in API maps.
func strKey(id int) string {
	return strconv.Itoa(id)
}

// hasAnySequences reports whether at least one player's metrics carries a
// non-empty time-series block. Used to detect stale analysis.json files that
// were written before sequences were wired through.
func hasAnySequences(a *api.Analysis) bool {
	for _, pm := range a.Metrics.Players {
		if pm.Sequences != nil && len(pm.Sequences.Time) > 0 {
			return true
		}
	}
	return false
}

// sourceHasSequences reports whether the user's metadata.json contains a
// `sequences` key. Cheap substring check — avoids parsing the whole file.
func sourceHasSequences(metaPath string) bool {
	raw, err := os.ReadFile(metaPath)
	if err != nil {
		return false
	}
	return strings.Contains(string(raw), `"sequences"`)
}

func buildAnalysis(g api.GameInfo, players []api.Player, m *metadata.Metadata, evs []api.Event, replayDir string, lib *techlib.Lib) *api.Analysis {
	tev := internalEvents(evs)
	phaseT := analytics.PhaseTimings(tev)
	eng := analytics.Engagements(tev, 3000)
	pg := analytics.PanicGarrison(tev)
	replayBasename := filepath.Base(replayDir)
	// Sequences come from metadata.json. Try the user's own first (populated
	// when the summary screen was viewed), then fall back to the sandbox copy
	// if a headless re-run was done.
	seqs, sandboxErr := sandbox.LoadFromMeta(filepath.Join(replayDir, "metadata.json"))
	if (seqs == nil || sandboxErr != nil) {
		if alt, err := sandbox.Load(replayBasename); err == nil && alt != nil {
			seqs, sandboxErr = alt, nil
		}
	}
	if sandboxErr != nil {
		fmt.Fprintf(os.Stderr, "replay: sequences load %s: %v\n", replayBasename, sandboxErr)
	}

	resignByPlayer := map[int]bool{}
	for _, e := range tev {
		if e.Type == events.TypeResign {
			resignByPlayer[e.Player] = true
		}
	}

	finalByPlayer := map[string]api.PlayerFinalState{}
	for i, ps := range m.PlayerStates {
		if i == 0 {
			continue // gaia
		}
		outcome := ps.State
		if resignByPlayer[i] {
			outcome = "defeated"
		}
		rc := api.Resources{
			Food:  ps.ResourceCounts["food"],
			Wood:  ps.ResourceCounts["wood"],
			Stone: ps.ResourceCounts["stone"],
			Metal: ps.ResourceCounts["metal"],
		}
		finalByPlayer[strKey(i)] = api.PlayerFinalState{
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
	improvsByPlayer := map[int][]api.ImprovementEntry{}
	for _, e := range tev {
		if e.Type != events.TypeResearch {
			continue
		}
		data, ok := e.Data.(events.ResearchData)
		if !ok {
			continue
		}
		entry := api.ImprovementEntry{
			TMs:      e.TMs,
			Template: data.Template,
		}
		if lib != nil {
			if info := lib.Resolve(data.Template); info != nil {
				if info.GenericName != "" {
					entry.GenericName = &info.GenericName
				}
				if info.Description != "" {
					entry.Description = &info.Description
				}
				if info.AutoResearch {
					b := true
					entry.AutoResearch = &b
				}
				if info.ResearchTime != 0 {
					entry.ResearchTime = &info.ResearchTime
				}
				if info.Cost.Food != 0 || info.Cost.Wood != 0 || info.Cost.Stone != 0 || info.Cost.Metal != 0 {
					cost := api.ImprovementCost{}
					if info.Cost.Food != 0 {
						cost.Food = &info.Cost.Food
					}
					if info.Cost.Wood != 0 {
						cost.Wood = &info.Cost.Wood
					}
					if info.Cost.Stone != 0 {
						cost.Stone = &info.Cost.Stone
					}
					if info.Cost.Metal != 0 {
						cost.Metal = &info.Cost.Metal
					}
					entry.Cost = &cost
				}
				if len(info.Buildings) > 0 {
					b0 := info.Buildings[0]
					entry.Building = &b0
					if len(info.Buildings) > 1 {
						bs := info.Buildings
						entry.Buildings = &bs
					}
				}
			}
		}
		improvsByPlayer[e.Player] = append(improvsByPlayer[e.Player], entry)
	}
	// Sort each player's improvements by time ascending.
	for p := range improvsByPlayer {
		sort.Slice(improvsByPlayer[p], func(i, j int) bool {
			return improvsByPlayer[p][i].TMs < improvsByPlayer[p][j].TMs
		})
	}

	metricsByPlayer := map[string]api.PlayerMetrics{}
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
			es = []api.Engagement{}
		}
		an := pg[p]
		if an == nil {
			an = []api.Anomaly{}
		}
		imps := improvsByPlayer[p]
		if imps == nil {
			imps = []api.ImprovementEntry{}
		}
		metricsByPlayer[strKey(p)] = api.PlayerMetrics{
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
		k := strKey(p)
		if _, ok := metricsByPlayer[k]; ok {
			continue
		}
		metricsByPlayer[k] = api.PlayerMetrics{
			PhaseTimings: map[string]int{},
			Engagements:  []api.Engagement{},
			Anomalies:    []api.Anomaly{},
			Sequences:    seqs[p],
			Improvements: []api.ImprovementEntry{},
		}
	}

	return &api.Analysis{
		SchemaVersion: output.SchemaVersion,
		Game:          g,
		Players:       players,
		Events:        evs,
		Snapshots:     []api.Snapshot{},
		FinalState:    api.FinalState{Players: finalByPlayer},
		Metrics:       api.Metrics{Players: metricsByPlayer},
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
