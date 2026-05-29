package webui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/replay/output"
)

func TestListReplaysEmptyRoot(t *testing.T) {
	dir := t.TempDir()
	srv := NewServer(dir, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/replays", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q", ct)
	}
	body := strings.TrimSpace(w.Body.String())
	if body != "[]" {
		t.Errorf("body = %q, want []", body)
	}
}

func TestGetReplayNotFound(t *testing.T) {
	dir := t.TempDir()
	srv := NewServer(dir, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/replays/DOESNOTEXIST", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("body parse: %v", err)
	}
	if body["error"] == "" {
		t.Errorf("body = %v, want error field", body)
	}
}

func TestGetReplayFound(t *testing.T) {
	dir := t.TempDir()
	// Use the existing fixture
	src := filepath.Join("..", "..", "..", "testdata", "replays", "short-germ-vs-3p")
	if _, err := os.Stat(src); err != nil {
		t.Skipf("fixture missing: %v", err)
	}
	rdir := filepath.Join(dir, "short-germ-vs-3p")
	if err := os.Mkdir(rdir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"commands.txt", "metadata.json"} {
		raw, err := os.ReadFile(filepath.Join(src, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(rdir, name), raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	a, err := replay.Run(rdir, nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	srv := NewServer(dir, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/replays/"+a.Game.MatchID, nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var got output.Analysis
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Game.MatchID != a.Game.MatchID {
		t.Errorf("MatchID = %q, want %q", got.Game.MatchID, a.Game.MatchID)
	}
	if len(got.Events) == 0 {
		t.Error("no events in response")
	}
}

func TestNonAPIReturns404(t *testing.T) {
	dir := t.TempDir()
	srv := NewServer(dir, nil)
	req := httptest.NewRequest(http.MethodGet, "/replay/DOESNOTEXIST", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (backend serves API only)", w.Code)
	}
}
