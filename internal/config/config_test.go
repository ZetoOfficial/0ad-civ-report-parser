package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

func TestDefaults(t *testing.T) {
	c := Defaults()
	if c.Gamedata != paths.DefaultGameDataRoot {
		t.Errorf("Gamedata = %q; want %q", c.Gamedata, paths.DefaultGameDataRoot)
	}
	if c.OutDir != "." {
		t.Errorf("OutDir = %q; want %q", c.OutDir, ".")
	}
	if c.Lang != "" {
		t.Errorf("Lang = %q; want empty", c.Lang)
	}
	if c.IncludeHistory {
		t.Errorf("IncludeHistory = true; want false")
	}
	if c.IncludeIcons {
		t.Errorf("IncludeIcons = true; want false")
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	c, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\") error: %v", err)
	}
	if c.Gamedata != paths.DefaultGameDataRoot {
		t.Errorf("expected defaults; Gamedata = %q", c.Gamedata)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/no-such-config.json")
	if err == nil {
		t.Fatal("expected error for missing file path")
	}
}

func TestLoad_PartialFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	body := `{"out_dir": "/tmp/out", "include_history": true}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.OutDir != "/tmp/out" {
		t.Errorf("OutDir = %q; want /tmp/out", c.OutDir)
	}
	if !c.IncludeHistory {
		t.Error("IncludeHistory = false; want true")
	}
	if c.Gamedata != paths.DefaultGameDataRoot {
		t.Errorf("Gamedata = %q; want default %q", c.Gamedata, paths.DefaultGameDataRoot)
	}
}

func TestLoad_BadJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
