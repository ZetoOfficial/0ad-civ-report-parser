package civdata

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/testutil"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

func TestLoadPlayerTemplate_Spart(t *testing.T) {
	testutil.SkipIfNoGameData(t)

	layout := paths.Layout{Root: testutil.GameDataRoot()}
	idx, err := tmpl.NewIndex(filepath.Join(testutil.GameDataRoot(), "simulation/templates"))
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
