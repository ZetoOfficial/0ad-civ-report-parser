// Package testutil hosts helpers shared by tests across the module.
// Functions here may import "testing" — they're meant to be called
// only from _test.go files in other packages.
package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// GameDataRoot returns the path to the 0 A.D. mods/public root used
// by tests. The OAD_GAMEDATA_ROOT env var overrides the built-in
// default; absence of the env var falls back to the path the project
// hardcodes for local development.
func GameDataRoot() string {
	if env := os.Getenv("OAD_GAMEDATA_ROOT"); env != "" {
		return env
	}
	return "/Users/zeto/Projects/study/0ad/binaries/data/mods/public"
}

// SkipIfNoGameData skips the test when the 0 A.D. data root is not
// available. The sentinel is simulation/templates — present iff the
// root is a real mod directory.
func SkipIfNoGameData(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(GameDataRoot(), "simulation/templates")); err != nil {
		t.Skipf("gamedata unavailable: %v", err)
	}
}
