package sandbox

import (
	"os"
	"testing"
)

func TestLoadFromRealSandbox(t *testing.T) {
	// Skip if the sandbox isn't populated (e.g. fresh checkout, no headless
	// regen run yet). Honors $OAD_REPLAY_SANDBOX so devs can point at any root.
	root := DefaultRoot()
	const sample = "2026-05-28_0001"
	if _, err := os.Stat(MetadataPath(root, sample)); err != nil {
		t.Skipf("sandbox sample missing at %s: %v", root, err)
	}

	seqs, err := LoadFromRoot(root, sample)
	if err != nil {
		t.Fatalf("LoadFromRoot: %v", err)
	}
	if len(seqs) == 0 {
		t.Fatal("expected at least one player with sequences")
	}
	for pid, s := range seqs {
		if len(s.Time) == 0 {
			t.Errorf("p%d: empty Time", pid)
		}
		if len(s.PopulationCount) != len(s.Time) {
			t.Errorf("p%d: PopulationCount len %d != Time len %d", pid, len(s.PopulationCount), len(s.Time))
		}
	}
}

func TestLoadMissingReturnsNil(t *testing.T) {
	seqs, err := LoadFromRoot("/nonexistent-sandbox", "anything")
	if err != nil {
		t.Errorf("LoadFromRoot on missing root: unexpected err: %v", err)
	}
	if seqs != nil {
		t.Errorf("expected nil seqs, got %v", seqs)
	}
}
