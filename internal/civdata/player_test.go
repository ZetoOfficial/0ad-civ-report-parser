package civdata

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

func gamedataRoot() string {
	if env := os.Getenv("OAD_GAMEDATA_ROOT"); env != "" {
		return env
	}
	return "/Users/zeto/Projects/study/0ad/binaries/data/mods/public"
}

func skipIfNoGamedata(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(gamedataRoot(), "simulation/templates")); err != nil {
		t.Skipf("gamedata unavailable: %v", err)
	}
}

func TestLoadPlayerTemplate_Spart(t *testing.T) {
	skipIfNoGamedata(t)

	layout := paths.Layout{Root: gamedataRoot()}
	idx, err := tmpl.NewIndex(filepath.Join(gamedataRoot(), "simulation/templates"))
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	resolver := tmpl.NewResolver(idx)

	pt, err := LoadPlayerTemplate(layout, "spart", resolver)
	if err != nil {
		t.Fatalf("LoadPlayerTemplate: %v", err)
	}
	if pt.GenericName != "Spartans" {
		t.Errorf("GenericName = %q; want %q", pt.GenericName, "Spartans")
	}
	if !strings.HasPrefix(pt.History, "Sparta was a prominent city-state") {
		t.Errorf("History prefix mismatch: %q", pt.History[:min(60, len(pt.History))])
	}
	if pt.IconPath != "emblems/emblem_spartans.png" {
		t.Errorf("IconPath = %q", pt.IconPath)
	}
	if strings.Contains(pt.History, `\n`) {
		t.Errorf("History still contains literal \\n: %q", pt.History)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
