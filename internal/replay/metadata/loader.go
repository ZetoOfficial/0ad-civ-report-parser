// Package metadata loads metadata.json next to a 0ad replay.
package metadata

import (
	"encoding/json"
	"fmt"
	"os"
)

type Metadata struct {
	TimeElapsed  int64         `json:"timeElapsed"`
	PlayerStates []PlayerState `json:"playerStates"`
}

type PlayerState struct {
	Name            string         `json:"name"`
	Civ             string         `json:"civ"`
	State           string         `json:"state"`
	Phase           string         `json:"phase"`
	PopCount        int            `json:"popCount"`
	PopLimit        int            `json:"popLimit"`
	PopMax          int            `json:"popMax"`
	Team            int            `json:"team"`
	Color           Color          `json:"color"`
	ResourceCounts  map[string]int `json:"resourceCounts"`
	ResearchedTechs []string       `json:"researchedTechs"`
}

type Color struct{ R, G, B int }

// Load reads and parses metadata.json at path.
func Load(path string) (*Metadata, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("metadata: read %s: %w", path, err)
	}
	var m Metadata
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("metadata: parse %s: %w", path, err)
	}
	return &m, nil
}
