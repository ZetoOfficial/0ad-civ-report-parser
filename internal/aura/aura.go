package aura

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
)

type Aura struct {
	Path            string
	Name            string
	Type            string              `json:"type"`
	Radius          float64             `json:"radius"`
	AffectedPlayers []string            `json:"affectedPlayers"`
	Affects         []any               `json:"affects"`
	Modifications   []tech.Modification `json:"modifications"`
	AuraName        string              `json:"auraName"`
	AuraDescription string              `json:"auraDescription"`
}

func (a *Aura) AffectsHumanReadable() []string {
	out := []string{}
	for _, raw := range a.Affects {
		switch v := raw.(type) {
		case string:
			out = append(out, v)
		case []any:
			parts := []string{}
			for _, p := range v {
				if s, ok := p.(string); ok {
					parts = append(parts, s)
				}
			}
			if len(parts) > 0 {
				out = append(out, strings.Join(parts, "+"))
			}
		}
	}
	return out
}

func Load(path string) (*Aura, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read aura %s: %w", path, err)
	}
	var a Aura
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("parse aura %s: %w", path, err)
	}
	a.Path = path
	a.Name = strings.TrimSuffix(filepath.Base(path), ".json")
	return &a, nil
}

func ListInDir(dir, prefix string) ([]*Aura, error) {
	out := []*Aura{}
	if _, err := os.Stat(dir); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return out, nil
		}
		return nil, err
	}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		base := strings.TrimSuffix(filepath.Base(path), ".json")
		if prefix != "" && !strings.HasPrefix(base, prefix) {
			return nil
		}
		a, err := Load(path)
		if err != nil {
			return err
		}
		out = append(out, a)
		return nil
	})
	return out, err
}
