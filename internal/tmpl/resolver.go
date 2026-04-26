package tmpl

import (
	"fmt"
	"strings"
)

type Resolver struct {
	Index *Index
	cache map[string]*Element
	stack map[string]bool
}

func NewResolver(idx *Index) *Resolver {
	return &Resolver{
		Index: idx,
		cache: make(map[string]*Element),
		stack: make(map[string]bool),
	}
}

func (r *Resolver) Resolve(absPath string) (*Element, error) {
	if cached, ok := r.cache[absPath]; ok {
		return cached, nil
	}
	if r.stack[absPath] {
		return nil, fmt.Errorf("template inheritance cycle detected at %s", absPath)
	}
	r.stack[absPath] = true
	defer delete(r.stack, absPath)

	raw, err := ParseFile(absPath)
	if err != nil {
		return nil, err
	}

	parentAttr := raw.Attr("parent")
	if parentAttr == "" {
		clone := raw.Clone()
		stripParentAttr(clone)
		r.cache[absPath] = clone
		return clone, nil
	}

	var merged *Element
	for _, parentRef := range strings.Split(parentAttr, "|") {
		parentRef = strings.TrimSpace(parentRef)
		if parentRef == "" {
			continue
		}
		parentPath, ok := r.Index.Lookup(parentRef)
		if !ok {
			return nil, fmt.Errorf("template %s: parent %q not found in index", absPath, parentRef)
		}
		parentResolved, err := r.Resolve(parentPath)
		if err != nil {
			return nil, fmt.Errorf("resolving parent %q of %s: %w", parentRef, absPath, err)
		}
		if merged == nil {
			merged = parentResolved.Clone()
		} else {
			merged = Merge(merged, parentResolved)
		}
	}

	selfStripped := raw.Clone()
	stripParentAttr(selfStripped)
	final := Merge(merged, selfStripped)
	r.cache[absPath] = final
	return final, nil
}

func stripParentAttr(e *Element) {
	if e == nil {
		return
	}
	if e.Attrs != nil {
		delete(e.Attrs, "parent")
		if len(e.Attrs) == 0 {
			e.Attrs = nil
		}
	}
}
