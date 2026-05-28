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
	ResearchedTechs []string       `json:"-"` // set by custom UnmarshalJSON
}

// UnmarshalJSON handles the real-replay schema where researchedTechs is an
// object (map of tech name → bool/int) as well as test fixtures where it is
// an array of strings.  Color values may be float64 (0.0–1.0 normalised) or
// integer (0–255) depending on the replay version.
func (ps *PlayerState) UnmarshalJSON(data []byte) error {
	type raw struct {
		Name            string          `json:"name"`
		Civ             string          `json:"civ"`
		State           string          `json:"state"`
		Phase           string          `json:"phase"`
		PopCount        int             `json:"popCount"`
		PopLimit        int             `json:"popLimit"`
		PopMax          int             `json:"popMax"`
		Team            int             `json:"team"`
		Color           Color           `json:"color"`
		ResourceCounts  map[string]int  `json:"resourceCounts"`
		ResearchedTechs json.RawMessage `json:"researchedTechs"`
	}
	var r raw
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	ps.Name = r.Name
	ps.Civ = r.Civ
	ps.State = r.State
	ps.Phase = r.Phase
	ps.PopCount = r.PopCount
	ps.PopLimit = r.PopLimit
	ps.PopMax = r.PopMax
	ps.Team = r.Team
	ps.Color = r.Color
	ps.ResourceCounts = r.ResourceCounts

	// researchedTechs can be a JSON array or a JSON object.
	if len(r.ResearchedTechs) > 0 && r.ResearchedTechs[0] == '[' {
		var arr []string
		if err := json.Unmarshal(r.ResearchedTechs, &arr); err != nil {
			return fmt.Errorf("metadata: researchedTechs array: %w", err)
		}
		ps.ResearchedTechs = arr
	} else if len(r.ResearchedTechs) > 0 && r.ResearchedTechs[0] == '{' {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(r.ResearchedTechs, &obj); err != nil {
			return fmt.Errorf("metadata: researchedTechs object: %w", err)
		}
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		ps.ResearchedTechs = keys
	}
	// null or empty → leave nil slice
	return nil
}

// Color holds player color values.  In modern replays (0.28+) the values are
// normalised floats in the 0.0–1.0 range; older fixtures may use 0–255 ints.
// We store as float64 to cover both.
type Color struct {
	R float64 `json:"r"`
	G float64 `json:"g"`
	B float64 `json:"b"`
}

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
