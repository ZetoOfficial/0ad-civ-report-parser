package tmpl

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type Index struct {
	TemplatesRoot string
	byBasename    map[string]string
	allPaths      []string
}

func NewIndex(templatesRoot string) (*Index, error) {
	idx := &Index{
		TemplatesRoot: templatesRoot,
		byBasename:    make(map[string]string),
	}
	err := filepath.WalkDir(templatesRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".xml") {
			return nil
		}
		base := strings.TrimSuffix(filepath.Base(path), ".xml")
		if existing, ok := idx.byBasename[base]; ok {
			if existing != path {
				existing = relPath(templatesRoot, existing)
				now := relPath(templatesRoot, path)
				if conflictPriority(now) > conflictPriority(existing) {
					idx.byBasename[base] = path
				}
				_ = fmt.Sprintf("template basename collision %q: %s vs %s", base, existing, now)
			}
			return nil
		}
		idx.byBasename[base] = path
		idx.allPaths = append(idx.allPaths, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("index walk %s: %w", templatesRoot, err)
	}
	return idx, nil
}

func relPath(root, p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return r
}

func conflictPriority(rel string) int {
	switch {
	case strings.HasPrefix(rel, "mixins/"):
		return 5
	case strings.HasPrefix(rel, "structures/"):
		return 4
	case strings.HasPrefix(rel, "units/"):
		return 4
	case strings.HasPrefix(rel, "template_"):
		return 3
	}
	return 0
}

func (i *Index) Lookup(parentRef string) (string, bool) {
	parentRef = strings.TrimSpace(parentRef)
	if parentRef == "" {
		return "", false
	}
	if !strings.Contains(parentRef, "/") {
		if abs, ok := i.byBasename[parentRef]; ok {
			return abs, true
		}
		return "", false
	}
	suffix := "/" + parentRef + ".xml"
	for _, p := range i.allPaths {
		if strings.HasSuffix(p, suffix) {
			return p, true
		}
	}
	candidate := filepath.Join(i.TemplatesRoot, parentRef+".xml")
	return candidate, true
}
