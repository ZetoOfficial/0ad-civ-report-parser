package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Write atomically marshals a as JSON to path (temp file + rename).
func Write(path string, a *Analysis) error {
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
	return os.Rename(tmp.Name(), path)
}

// IsFresh reports whether path exists and is newer than the source (commandsTxt).
func IsFresh(path, commandsTxt string) bool {
	a, err := os.Stat(path)
	if err != nil {
		return false
	}
	b, err := os.Stat(commandsTxt)
	if err != nil {
		return false
	}
	return a.ModTime().After(b.ModTime())
}
