package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	api "github.com/ZetoOfficial/0ad-civ-report-parser/internal/api/gen"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/techlib"
)

const listCacheTTL = 5 * time.Second

type handlers struct {
	repRoot string
	lib     *techlib.Lib

	mu       sync.Mutex
	cached   []api.ReplaySummary
	cachedAt time.Time
}

func (h *handlers) GetApiReplays(w http.ResponseWriter, r *http.Request) {
	items, err := h.buildList()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("scan: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *handlers) buildList() ([]api.ReplaySummary, error) {
	h.mu.Lock()
	if h.cached != nil && time.Since(h.cachedAt) < listCacheTTL {
		cached := h.cached
		h.mu.Unlock()
		return cached, nil
	}
	h.mu.Unlock()

	entries, err := os.ReadDir(h.repRoot)
	if err != nil {
		return nil, err
	}
	items := []api.ReplaySummary{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(h.repRoot, e.Name())
		if _, err := os.Stat(filepath.Join(dir, "metadata.json")); err != nil {
			continue
		}
		a, err := replay.Run(dir, h.lib)
		if err != nil {
			continue
		}
		// Hide replays without sequences — without them we can't show the
		// charts that justify being in the list.
		if !hasSequences(a) {
			continue
		}
		items = append(items, api.ReplaySummary{
			MatchId:    a.Game.MatchId,
			Map:        a.Game.Map,
			Timestamp:  a.Game.Timestamp,
			DurationMs: a.Game.DurationMs,
			Players:    a.Players,
			Outcome:    resolveOutcome(a),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Timestamp > items[j].Timestamp })

	h.mu.Lock()
	h.cached = items
	h.cachedAt = time.Now()
	h.mu.Unlock()
	return items, nil
}

// hasSequences reports whether any player has time-series data attached.
// Used to filter the list to only replays that have been regenerated through
// the headless-replay sandbox.
func hasSequences(a *api.Analysis) bool {
	for _, pm := range a.Metrics.Players {
		if pm.Sequences != nil && len(pm.Sequences.Time) > 0 {
			return true
		}
	}
	return false
}

// resolveOutcome picks the most informative outcome: prefer the first
// human (non-AI) player; fall back to the lowest player ID with a recorded
// state; fall back to "—" if nothing usable.
func resolveOutcome(a *api.Analysis) string {
	humanIDs := []int{}
	for _, p := range a.Players {
		if !p.IsAi {
			humanIDs = append(humanIDs, p.Id)
		}
	}
	sort.Ints(humanIDs)
	for _, id := range humanIDs {
		k := fmt.Sprintf("%d", id)
		if fs, ok := a.FinalState.Players[k]; ok && fs.Outcome != "" {
			return fs.Outcome
		}
	}
	// fallback: any player with an outcome; iterate in sorted string-key order
	keys := make([]string, 0, len(a.FinalState.Players))
	for k := range a.FinalState.Players {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if fs := a.FinalState.Players[k]; fs.Outcome != "" {
			return fs.Outcome
		}
	}
	return "—"
}

func (h *handlers) GetApiReplaysByMatchID(w http.ResponseWriter, r *http.Request, matchID string) {
	if matchID == "" {
		writeJSONError(w, http.StatusBadRequest, "missing matchID")
		return
	}
	a, err := h.findByMatchID(matchID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *handlers) findByMatchID(matchID string) (*api.Analysis, error) {
	entries, _ := os.ReadDir(h.repRoot)
	for _, e := range entries {
		dir := filepath.Join(h.repRoot, e.Name())
		path := filepath.Join(dir, "analysis.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var a api.Analysis
		if err := json.Unmarshal(raw, &a); err != nil {
			continue
		}
		if a.Game.MatchId == matchID {
			return &a, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// Ensure handlers implements the generated ServerInterface.
var _ api.ServerInterface = (*handlers)(nil)
