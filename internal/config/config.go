package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

type Config struct {
	Gamedata       string `json:"gamedata"`
	OutDir         string `json:"out_dir"`
	Lang           string `json:"lang"`
	IncludeHistory bool   `json:"include_history"`
	IncludeIcons   bool   `json:"include_icons"`
}

func Defaults() Config {
	return Config{
		Gamedata: paths.DefaultGameDataRoot,
		OutDir:   ".",
	}
}

// Load reads a JSON config file at path and overlays it on Defaults().
// If path is empty, returns Defaults() with no error.
// If the file is missing or malformed, returns an error.
func Load(path string) (*Config, error) {
	c := Defaults()
	if path == "" {
		return &c, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}
	return &c, nil
}
