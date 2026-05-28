package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "analysis.json")
	in := &Analysis{
		SchemaVersion: SchemaVersion,
		Game:          GameInfo{MatchID: "ABC", Map: "punjab_2", DurationMs: 1860400},
		Players:       []Player{{ID: 1, Name: "zeto", Civ: "spart"}},
		Events:        []Event{{T: 1200, Player: 1, Type: "train"}},
		Snapshots:     []Snapshot{},
	}
	if err := Write(path, in); err != nil {
		t.Fatalf("Write: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got Analysis
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d, want 1", got.SchemaVersion)
	}
	if got.Game.MatchID != "ABC" {
		t.Errorf("MatchID = %q, want ABC", got.Game.MatchID)
	}
	if len(got.Players) != 1 || got.Players[0].Civ != "spart" {
		t.Errorf("Players = %+v", got.Players)
	}
}

func TestIsFresh(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "commands.txt")
	out := filepath.Join(dir, "analysis.json")
	os.WriteFile(src, []byte("x"), 0o644)
	if IsFresh(out, src) {
		t.Fatal("missing analysis must be stale")
	}
	os.WriteFile(out, []byte("{}"), 0o644)
	// out written after src (Go writes are monotonic in practice; but force via Chtimes)
	now := time.Now()
	os.Chtimes(src, now.Add(-time.Minute), now.Add(-time.Minute))
	os.Chtimes(out, now, now)
	if !IsFresh(out, src) {
		t.Fatal("analysis newer than commands must be fresh")
	}
}
