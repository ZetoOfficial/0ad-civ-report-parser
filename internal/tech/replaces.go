package tech

import (
	"fmt"
	"slices"
	"sort"
)

// Index holds the replaces/supersedes graph over all technologies in a Catalog.
type Index struct {
	techs        map[string]*Technology // basename → tech (alias of Catalog cache)
	replacedBy   map[string][]string    // X → all Y where Y.Replaces ⊇ {X}
	supersededBy map[string]string      // X → Y where Y.Supersedes == X
	replaces     map[string][]string    // mirror: Y → list of X where Y replaces X
	Warnings     []string               // diagnostics produced during build
}

// ChainInfo holds the replacement/supersession links for a single technology.
type ChainInfo struct {
	Replaces     []string // techs this one replaces (= t.Replaces)
	ReplacedBy   []string // all techs that replace this one (raw, not civ-aware)
	Supersedes   string   // this tech upgrades from (= t.Supersedes)
	SupersededBy string   // tech that supersedes this one
}

// NewIndex calls c.LoadAll then builds the in-memory graph.
func NewIndex(c *Catalog) (*Index, error) {
	if err := c.LoadAll(); err != nil {
		return nil, fmt.Errorf("tech.NewIndex LoadAll: %w", err)
	}
	idx := &Index{
		techs:        c.AllLoaded(),
		replacedBy:   make(map[string][]string),
		supersededBy: make(map[string]string),
		replaces:     make(map[string][]string),
	}
	for _, t := range idx.techs {
		for _, r := range t.Replaces {
			idx.replacedBy[r] = append(idx.replacedBy[r], t.Name)
			idx.replaces[t.Name] = append(idx.replaces[t.Name], r)
		}
		if t.Supersedes != "" {
			if existing, ok := idx.supersededBy[t.Supersedes]; ok {
				if t.Name < existing {
					idx.supersededBy[t.Supersedes] = t.Name
				}
				idx.Warnings = append(idx.Warnings,
					fmt.Sprintf("multiple supersedes for %q: %s vs %s", t.Supersedes, t.Name, existing))
			} else {
				idx.supersededBy[t.Supersedes] = t.Name
			}
		}
	}
	// Sort and deduplicate each slice for determinism.
	for k := range idx.replacedBy {
		idx.replacedBy[k] = sortDedup(idx.replacedBy[k])
	}
	for k := range idx.replaces {
		idx.replaces[k] = sortDedup(idx.replaces[k])
	}
	return idx, nil
}

// Chain returns the replacement/supersession links for name.
// Returns zero ChainInfo if name is not indexed.
func (i *Index) Chain(name string) ChainInfo {
	out := ChainInfo{
		Replaces:     i.replaces[name],
		ReplacedBy:   i.replacedBy[name],
		SupersededBy: i.supersededBy[name],
	}
	if t, ok := i.techs[name]; ok {
		out.Supersedes = t.Supersedes
	}
	return out
}

// sortDedup sorts ss in place and returns a deduplicated slice.
func sortDedup(ss []string) []string {
	sort.Strings(ss)
	out := ss[:0]
	for i, s := range ss {
		if i == 0 || s != ss[i-1] {
			out = append(out, s)
		}
	}
	return out
}

// civAffinity returns the single civ a technology is specific to.
// It checks requirements.civ first; if empty, falls back to specificName:
// a single-key specificName map is treated as civ affinity (e.g.
// specificName: {"athen": "..."} → "athen"). Multi-key maps indicate
// a generic tech and return "".
//
// Note: phase_town_athen.json (and similar civ-variant phase techs)
// express civ-affinity via a single-key specificName map rather than
// requirements.civ; the fallback here is necessary because those files
// use requirements.entity for the phase gate instead.
func civAffinity(t *Technology) string {
	if civ := RequiresCiv(t.Requirements); civ != "" {
		return civ
	}
	if len(t.SpecificName) == 1 {
		for k := range t.SpecificName {
			return k
		}
	}
	return ""
}

// Get returns the Technology with the given name, or nil if not indexed.
// Unlike ResolveForCiv, Get does NOT resolve civ-specific replacements —
// it returns the literal tech for the given name.
func (i *Index) Get(name string) *Technology {
	return i.techs[name]
}

// ResolveForCiv returns the technology that replaces name for the given civ.
// Returns idx.techs[name] (possibly nil) if no replacement applies.
//
// Algorithm:
//  1. Among replacedBy[name], pick the one with civAffinity == civ.
//  2. Else among the same candidates, pick those with civAffinity == "" and
//     civ not in notciv list. If exactly one, return it.
//  3. Else return idx.techs[name].
//  4. If step 2 yields multiple candidates: stable sort by Name, return
//     first, append to idx.Warnings.
func (i *Index) ResolveForCiv(name, civ string) *Technology {
	candidates := make([]*Technology, 0, len(i.replacedBy[name]))
	for _, rname := range i.replacedBy[name] {
		if t, ok := i.techs[rname]; ok {
			candidates = append(candidates, t)
		}
	}

	// Step 1: civ-specific replacement.
	for _, t := range candidates {
		if civAffinity(t) == civ {
			return t
		}
	}

	// Step 2: generic replacements not blocked by notciv.
	var generic []*Technology
	for _, t := range candidates {
		if civAffinity(t) != "" {
			continue // already for another civ
		}
		if !slices.Contains(NotCivList(t.Requirements), civ) {
			generic = append(generic, t)
		}
	}
	if len(generic) == 1 {
		return generic[0]
	}
	if len(generic) > 1 {
		sort.Slice(generic, func(a, b int) bool {
			return generic[a].Name < generic[b].Name
		})
		names := make([]string, len(generic))
		for j, g := range generic {
			names[j] = g.Name
		}
		i.Warnings = append(i.Warnings,
			fmt.Sprintf("multiple generic replacements for %q (civ %s): %v", name, civ, names))
		return generic[0]
	}

	// Step 3: fall back to the original tech.
	return i.techs[name]
}
