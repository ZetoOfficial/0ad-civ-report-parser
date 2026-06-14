// Package techlib loads 0ad tech metadata + reverse-maps which structure
// templates research each tech. Used by the replay pipeline to humanize
// research events.
package techlib

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

// Cost is the resource cost of a technology.
type Cost struct {
	Food  int `json:"food,omitempty"`
	Wood  int `json:"wood,omitempty"`
	Stone int `json:"stone,omitempty"`
	Metal int `json:"metal,omitempty"`
}

// Info holds resolved metadata for a single technology.
type Info struct {
	GenericName  string   `json:"generic_name"`
	Description  string   `json:"description"`
	Tooltip      string   `json:"tooltip"`
	AutoResearch bool     `json:"auto_research"`
	Cost         Cost     `json:"cost"`
	ResearchTime float64  `json:"research_time"`
	Buildings    []string `json:"buildings"` // structure basenames; sorted, deduped
}

// Lib holds the full resolved tech library.
type Lib struct {
	byName map[string]*Info
}

// Load builds a Lib from a gamedata mod root (the public/ dir). It walks
// simulation/data/technologies/ for tech metadata and
// simulation/templates/structures/ for the inverse tech->building map. The
// resolver applies parent-template inheritance, so techs declared on parent
// templates are credited to the actual concrete structures that inherit them.
//
// Returns an error if the gamedata root is missing or unreadable. Individual
// per-template errors are logged but do not abort the load.
func Load(gamedataRoot string) (*Lib, error) {
	layout := paths.Layout{Root: gamedataRoot}

	if _, err := os.Stat(gamedataRoot); err != nil {
		return nil, fmt.Errorf("techlib: gamedata root %q: %w", gamedataRoot, err)
	}

	// Step 1: load all tech metadata.
	catalog := tech.NewCatalog(layout.Technologies())
	if err := catalog.LoadAll(); err != nil {
		return nil, fmt.Errorf("techlib: load tech catalog: %w", err)
	}
	all := catalog.AllLoaded()

	byName := make(map[string]*Info, len(all))
	for name, t := range all {
		byName[name] = &Info{
			GenericName:  t.GenericName,
			Description:  t.Description,
			Tooltip:      t.Tooltip,
			AutoResearch: t.AutoResearch,
			Cost: Cost{
				Food:  t.Cost.Food,
				Wood:  t.Cost.Wood,
				Stone: t.Cost.Stone,
				Metal: t.Cost.Metal,
			},
			ResearchTime: t.ResearchTime,
		}
	}

	// Step 2: build reverse map: tech name -> set of building basenames.
	buildingsForTech := make(map[string]map[string]struct{})

	idx, err := tmpl.NewIndex(layout.Templates())
	if err != nil {
		return nil, fmt.Errorf("techlib: template index: %w", err)
	}
	r := tmpl.NewResolver(idx)

	structuresRoot := filepath.Join(layout.Templates(), "structures")
	walkErr := filepath.WalkDir(structuresRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable dirs
		}
		if d.IsDir() || !strings.HasSuffix(path, ".xml") {
			return nil
		}
		base := strings.TrimSuffix(filepath.Base(path), ".xml")
		// Skip parent templates.
		if strings.HasPrefix(base, "template_") {
			return nil
		}

		el, resolveErr := r.Resolve(path)
		if resolveErr != nil {
			fmt.Fprintf(os.Stderr, "techlib: resolve %s: %v\n", path, resolveErr)
			return nil
		}

		// Determine civ from path: structures/<civ>/building.xml.
		// If the parent dir name is "structures" itself (no civ subdir), civ is empty.
		civCode := ""
		rel, _ := filepath.Rel(structuresRoot, path)
		parts := strings.SplitN(rel, string(filepath.Separator), 2)
		if len(parts) == 2 {
			// parts[0] = civ dir, parts[1] = rest of path
			civCode = parts[0]
		}

		// Collect techs from all three possible token fields.
		var techTokens []string
		techTokens = append(techTokens, el.GetTokens("Researcher/Technologies")...)
		techTokens = append(techTokens, el.GetTokens("Trainer/Technologies")...)
		techTokens = append(techTokens, el.GetTokens("ProductionQueue/Technologies")...)

		for _, token := range techTokens {
			// Tokens prefixed with "-" are removals; skip them.
			if strings.HasPrefix(token, "-") {
				continue
			}
			// Apply {civ}/{native} substitution if we know the civ.
			if civCode != "" {
				token = tmpl.SubstCiv(token, civCode)
			}
			// Tokens that contain "/" are paths, not names; strip to basename.
			techName := token
			if strings.Contains(token, "/") {
				techName = filepath.Base(token)
			}
			// pair_ tokens or substituted tokens may not exist in the catalog; skip if not found.
			if _, ok := byName[techName]; !ok {
				continue
			}
			if buildingsForTech[techName] == nil {
				buildingsForTech[techName] = make(map[string]struct{})
			}
			buildingsForTech[techName][base] = struct{}{}
		}
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("techlib: walk structures: %w", walkErr)
	}

	// Flatten and sort building sets into Info.Buildings.
	for techName, bset := range buildingsForTech {
		info, ok := byName[techName]
		if !ok {
			continue
		}
		bs := make([]string, 0, len(bset))
		for b := range bset {
			bs = append(bs, b)
		}
		sort.Strings(bs)
		info.Buildings = bs
	}

	return &Lib{byName: byName}, nil
}

// Resolve returns metadata for techName, or nil if unknown.
func (l *Lib) Resolve(techName string) *Info {
	if l == nil {
		return nil
	}
	return l.byName[techName]
}
