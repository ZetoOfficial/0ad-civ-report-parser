package civdata

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type CivBonus struct {
	Name        string `json:"Name"`
	History     string `json:"History"`
	Description string `json:"Description"`
}

type TeamBonus struct {
	Name        string `json:"Name"`
	History     string `json:"History"`
	Description string `json:"Description"`
}

type StartEntity struct {
	Template string `json:"Template"`
	Count    int    `json:"Count,omitempty"`
}

type Civ struct {
	Code                  string            `json:"Code"`
	CultureRaw            json.RawMessage   `json:"Culture"`
	CivBonuses            []CivBonus        `json:"CivBonuses"`
	TeamBonuses           []TeamBonus       `json:"TeamBonuses"`
	WallSets              []string          `json:"WallSets"`
	StartEntities         []StartEntity     `json:"StartEntities"`
	SkirmishReplacements  map[string]string `json:"SkirmishReplacements"`
	SelectableInGameSetup bool              `json:"SelectableInGameSetup"`
	AINames               []string          `json:"AINames"`
}

func (c *Civ) Culture() string {
	if len(c.CultureRaw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(c.CultureRaw, &s); err == nil {
		return s
	}
	var arr []string
	if err := json.Unmarshal(c.CultureRaw, &arr); err == nil {
		return strings.Join(arr, ", ")
	}
	return string(c.CultureRaw)
}

func LoadCiv(path string) (*Civ, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read civ %s: %w", path, err)
	}
	var c Civ
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse civ %s: %w", path, err)
	}
	return &c, nil
}
