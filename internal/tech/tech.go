package tech

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Modification struct {
	Value      string          `json:"value"`
	Multiply   float64         `json:"multiply,omitempty"`
	Add        float64         `json:"add,omitempty"`
	Replace    any             `json:"replace,omitempty"`
	AffectsRaw json.RawMessage `json:"affects,omitempty"`
}

// AffectsList parses the per-modification "affects" field, which the JSON
// source may store either as a single string ("Melee") or an array
// (["Melee", "Cavalry"]). Returns nil if the field is absent.
func (m Modification) AffectsList() []string {
	if len(m.AffectsRaw) == 0 {
		return nil
	}
	var s string
	if err := json.Unmarshal(m.AffectsRaw, &s); err == nil {
		if s == "" {
			return nil
		}
		return []string{s}
	}
	var arr []string
	if err := json.Unmarshal(m.AffectsRaw, &arr); err == nil {
		return arr
	}
	return nil
}

type Cost struct {
	Wood  int `json:"wood,omitempty"`
	Stone int `json:"stone,omitempty"`
	Metal int `json:"metal,omitempty"`
	Food  int `json:"food,omitempty"`
}

type Requirements map[string]any

type Technology struct {
	Name          string
	Path          string
	GenericName   string         `json:"genericName"`
	Description   string         `json:"description"`
	SpecificName  map[string]any `json:"specificName,omitempty"`
	AutoResearch  bool           `json:"autoResearch"`
	Cost          Cost           `json:"cost"`
	ResearchTime  float64        `json:"researchTime"`
	Tooltip       string         `json:"tooltip"`
	Modifications []Modification `json:"modifications"`
	Affects       []string       `json:"affects"`
	Requirements  Requirements   `json:"requirements"`
	Supersedes    string         `json:"supersedes"`
	Pair          string         `json:"pair"`
	Top           string         `json:"top"`
	Bottom        string         `json:"bottom"`
	Icon                string         `json:"icon"`
	ReplacedBy          string         `json:"replacedBy"`
	RequirementsTooltip string         `json:"requirementsTooltip"`
	Replaces            []string       `json:"replaces"`
}

func Load(path string) (*Technology, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read tech %s: %w", path, err)
	}
	var t Technology
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("parse tech %s: %w", path, err)
	}
	t.Path = path
	t.Name = strings.TrimSuffix(filepath.Base(path), ".json")
	return &t, nil
}

type Catalog struct {
	dir   string
	cache map[string]*Technology
}

func NewCatalog(techDir string) *Catalog {
	return &Catalog{dir: techDir, cache: make(map[string]*Technology)}
}

func (c *Catalog) ByName(name string) (*Technology, error) {
	if t, ok := c.cache[name]; ok {
		return t, nil
	}
	candidates := []string{
		filepath.Join(c.dir, name+".json"),
		filepath.Join(c.dir, "civbonuses", name+".json"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			t, err := Load(p)
			if err != nil {
				return nil, err
			}
			c.cache[name] = t
			return t, nil
		}
	}
	return nil, fmt.Errorf("tech %q not found in %s", name, c.dir)
}

func (c *Catalog) AllCivBonuses(civ string) ([]*Technology, error) {
	bonusDir := filepath.Join(c.dir, "civbonuses")
	out := []*Technology{}
	err := filepath.WalkDir(bonusDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		t, err := Load(path)
		if err != nil {
			return err
		}
		if RequiresCiv(t.Requirements) == civ {
			out = append(out, t)
		}
		return nil
	})
	return out, err
}

func RequiresCiv(req Requirements) string {
	if req == nil {
		return ""
	}
	if v, ok := req["civ"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if all, ok := req["all"]; ok {
		if list, ok := all.([]any); ok {
			for _, item := range list {
				if m, ok := item.(map[string]any); ok {
					if s, ok := m["civ"].(string); ok {
						return s
					}
				}
			}
		}
	}
	return ""
}

func NotCivList(req Requirements) []string {
	out := []string{}
	if req == nil {
		return out
	}
	if v, ok := req["notciv"]; ok {
		switch x := v.(type) {
		case string:
			out = append(out, x)
		case []any:
			for _, e := range x {
				if s, ok := e.(string); ok {
					out = append(out, s)
				}
			}
		}
	}
	if all, ok := req["all"]; ok {
		if list, ok := all.([]any); ok {
			for _, item := range list {
				if m, ok := item.(map[string]any); ok {
					out = append(out, NotCivList(m)...)
				}
			}
		}
	}
	if anyClause, ok := req["any"]; ok {
		if list, ok := anyClause.([]any); ok {
			for _, item := range list {
				if m, ok := item.(map[string]any); ok {
					out = append(out, NotCivList(m)...)
				}
			}
		}
	}
	return out
}

func (c *Catalog) AllNotCiv(civ string) ([]*Technology, error) {
	out := []*Technology{}
	err := filepath.WalkDir(c.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		t, err := Load(path)
		if err != nil {
			return err
		}
		for _, blocked := range NotCivList(t.Requirements) {
			if blocked == civ {
				out = append(out, t)
				return nil
			}
		}
		return nil
	})
	return out, err
}
