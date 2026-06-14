package techlib

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

func gamedataRoot() string {
	if env := os.Getenv(paths.EnvGameDataRoot); env != "" {
		return env
	}
	return paths.DefaultGameDataRoot
}

func skipIfNoGameData(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(gamedataRoot(), "simulation/templates")); err != nil {
		t.Skipf("gamedata unavailable: %v", err)
	}
}

func TestLoad(t *testing.T) {
	skipIfNoGameData(t)

	lib, err := Load(gamedataRoot())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if lib == nil {
		t.Fatal("lib is nil")
	}

	// Known tech that every civ can research.
	info := lib.Resolve("phase_town_athen")
	if info == nil {
		t.Fatal("Resolve(phase_town_athen) returned nil")
	}
	if info.GenericName == "" {
		t.Error("phase_town_athen: GenericName is empty")
	}
	if !slices.Contains(info.Buildings, "civil_centre") {
		t.Errorf("phase_town_athen: Buildings = %v, want to contain civil_centre", info.Buildings)
	}

	// Non-existent tech must return nil.
	if got := lib.Resolve("does_not_exist"); got != nil {
		t.Errorf("Resolve(does_not_exist) = %v, want nil", got)
	}
}

func TestNilLib(t *testing.T) {
	var l *Lib
	if l.Resolve("anything") != nil {
		t.Error("nil Lib.Resolve must return nil")
	}
}
