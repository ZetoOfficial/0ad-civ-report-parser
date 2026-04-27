package render

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type damageType struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}

type resourceType struct {
	Code        string            `json:"code"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Order       int               `json:"order"`
	Subtypes    map[string]string `json:"subtypes"`
}

type statusEffect struct {
	Code            string `json:"code"`
	StatusName      string `json:"statusName"`
	ApplierTooltip  string `json:"applierTooltip"`
	ReceiverTooltip string `json:"receiverTooltip"`
}

func loadDamageTypes(dir string) ([]damageType, error) {
	out := []damageType{}
	if err := loadJSONDir(dir, func(raw []byte, _ string) error {
		var d damageType
		if err := json.Unmarshal(raw, &d); err != nil {
			return err
		}
		out = append(out, d)
		return nil
	}); err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].Code < out[j].Code
	})
	return out, nil
}

func loadResources(dir string) ([]resourceType, error) {
	out := []resourceType{}
	if err := loadJSONDir(dir, func(raw []byte, _ string) error {
		var r resourceType
		if err := json.Unmarshal(raw, &r); err != nil {
			return err
		}
		out = append(out, r)
		return nil
	}); err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].Code < out[j].Code
	})
	return out, nil
}

func loadStatusEffects(dir string) ([]statusEffect, error) {
	out := []statusEffect{}
	if err := loadJSONDir(dir, func(raw []byte, _ string) error {
		var s statusEffect
		if err := json.Unmarshal(raw, &s); err != nil {
			return err
		}
		out = append(out, s)
		return nil
	}); err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out, nil
}

func loadJSONDir(dir string, fn func(raw []byte, path string) error) error {
	matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return err
	}
	sort.Strings(matches)
	for _, p := range matches {
		raw, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("read %s: %w", p, err)
		}
		if err := fn(raw, p); err != nil {
			return fmt.Errorf("parse %s: %w", p, err)
		}
	}
	return nil
}

// resourceSubtypeKeys returns subtypes ordered alphabetically for
// deterministic rendering.
func resourceSubtypeKeys(r resourceType) []string {
	keys := make([]string, 0, len(r.Subtypes))
	for k := range r.Subtypes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// joinSubtypes formats subtypes as "fish, fruit, grain, meat".
func joinSubtypes(r resourceType) string {
	keys := resourceSubtypeKeys(r)
	return strings.Join(keys, ", ")
}
