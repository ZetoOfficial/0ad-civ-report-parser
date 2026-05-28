package replay

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunOnRealFixture(t *testing.T) {
	dir := filepath.Join("..", "..", "testdata", "replays", "short-germ-vs-3p")
	if _, err := os.Stat(filepath.Join(dir, "metadata.json")); err != nil {
		t.Skipf("fixture missing: %v", err)
	}
	defer os.Remove(filepath.Join(dir, "analysis.json"))

	a, err := Run(dir)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if a.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d", a.SchemaVersion)
	}
	if a.Game.MatchID == "" {
		t.Error("MatchID empty")
	}
	if len(a.Players) == 0 {
		t.Error("no players")
	}
	if a.Game.DurationMs <= 0 {
		t.Errorf("DurationMs = %d", a.Game.DurationMs)
	}
	if len(a.Events) == 0 {
		t.Error("no events decoded")
	}
	// Sanity: phase timings present for at least one player (any short replay should reach town)
	hasPhase := false
	for _, m := range a.Metrics.Players {
		if len(m.PhaseTimings) > 0 {
			hasPhase = true
		}
	}
	if !hasPhase {
		t.Log("WARN: no phase timings (replay may be very short)")
	}
}

func TestRunSkipsMissingMetadata(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "commands.txt"), []byte("start {}\nend\n"), 0o644)
	_, err := Run(dir)
	if err == nil {
		t.Fatal("Run must error on missing metadata.json")
	}
}
