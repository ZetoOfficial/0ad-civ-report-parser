package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"
)

type handlers struct {
	repRoot string
}

type replayListItem struct {
	MatchID    string          `json:"match_id"`
	Map        string          `json:"map"`
	Timestamp  int64           `json:"timestamp"`
	DurationMs int64           `json:"duration_ms"`
	Players    []output.Player `json:"players"`
	Outcome    string          `json:"outcome"`
}

func (h *handlers) listReplays(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(h.repRoot)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("scan: %v", err))
		return
	}
	var items []replayListItem
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
		items = append(items, replayListItem{
			MatchID:    a.Game.MatchID,
			Map:        a.Game.Map,
			Timestamp:  a.Game.Timestamp,
			DurationMs: a.Game.DurationMs,
			Players:    a.Players,
			Outcome:    outcome,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Timestamp > items[j].Timestamp })
	writeJSON(w, http.StatusOK, items)
}

func (h *handlers) getReplay(w http.ResponseWriter, r *http.Request) {
	matchID := strings.TrimPrefix(r.URL.Path, "/api/replays/")
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

func (h *handlers) findByMatchID(matchID string) (*output.Analysis, error) {
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
