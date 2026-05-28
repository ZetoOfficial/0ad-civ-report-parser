package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReaderSample(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "sample.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	r := New(f)
	want := []LineKind{KindStart, KindTurn, KindHash, KindHashQuick, KindTurn, KindCmd, KindCmd, KindTurn, KindCmd, KindEnd}
	var got []LineKind
	for {
		ln, ok, err := r.Next()
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		if !ok {
			break
		}
		got = append(got, ln.Kind)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d lines, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d: got %d, want %d", i, got[i], want[i])
		}
	}
}

func TestReaderTurnFields(t *testing.T) {
	f, _ := os.Open(filepath.Join("testdata", "sample.txt"))
	defer f.Close()
	r := New(f)
	r.Next() // start
	ln, _, _ := r.Next()
	if ln.Kind != KindTurn || ln.TurnN != 0 || ln.TickMs != 200 {
		t.Errorf("turn 0 200: got %+v", ln)
	}
}

func TestReaderCmdPlayer(t *testing.T) {
	f, _ := os.Open(filepath.Join("testdata", "sample.txt"))
	defer f.Close()
	r := New(f)
	for {
		ln, ok, err := r.Next()
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("no cmd found")
		}
		if ln.Kind == KindCmd {
			if ln.Player != 2 {
				t.Errorf("player = %d, want 2", ln.Player)
			}
			return
		}
	}
}
