// Package output provides atomic JSON writing and mtime-freshness check for
// analysis files. Types are defined in internal/api/gen.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	api "github.com/ZetoOfficial/0ad-civ-report-parser/internal/api/gen"
)

const SchemaVersion = 6

// Write atomically marshals a as JSON to path (temp file + rename).
func Write(path string, a *api.Analysis) error {
	raw, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("output: marshal: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".analysis-*.json")
	if err != nil {
		return fmt.Errorf("output: tempfile: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return fmt.Errorf("output: write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("output: close: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("output: rename: %w", err)
	}
	return nil
}

// IsFresh reports whether path exists and is newer than every source provided.
// Missing sources are ignored (treated as "no constraint"); a missing path is
// treated as not-fresh.
func IsFresh(path string, sources ...string) bool {
	a, err := os.Stat(path)
	if err != nil {
		return false
	}
	aTime := a.ModTime()
	for _, src := range sources {
		b, err := os.Stat(src)
		if err != nil {
			continue
		}
		if b.ModTime().After(aTime) {
			return false
		}
	}
	return true
}
