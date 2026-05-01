package civdata

import (
	"errors"
	"io/fs"
	"os"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

// SkipNote records a template or technology token that could not be resolved.
type SkipNote struct {
	Token  string
	Reason string
}

// WallSetGroup groups a wallset wrapper entity with its piece entities.
// Populated by IdentifyWallSets (Task 4); declared here to establish the
// final type shape.
type WallSetGroup struct {
	Wrapper Entity
	Pieces  []WallPiece
	Phase   Phase
}

// WallPiece is a single piece of a wallset (gate, long wall, tower, etc.).
type WallPiece struct {
	Role   string
	Entity Entity
}

// ReachResult holds the transitive closure of buildings, units, techs and
// wallsets reachable from a civilization's StartEntities.
type ReachResult struct {
	Buildings []Entity
	Units     []Entity
	Techs     map[string]*tech.Technology // pair wrappers are expanded into top+bottom
	WallSets  []*WallSetGroup             // populated by IdentifyWallSets (Task 4)
	Skipped   []SkipNote
}

// Reach computes the transitive closure from civ.StartEntities via
// Trainer/Builder/ProductionQueue Entities, Trainer/ProductionQueue/
// Researcher Technologies, and WallSet/Templates children. Pair-techs
// are expanded at scan time; the wrapper itself is NOT placed in
// res.Techs. Idempotent.
func Reach(civ *Civ, idx *tmpl.Index, resolver *tmpl.Resolver, catalog *tech.Catalog) (*ReachResult, error) {
	seen := map[string]struct{}{}  // resolved template tokens
	seenT := map[string]struct{}{} // tech names (separate namespace)

	queueE := make([]string, 0, len(civ.StartEntities))
	var queueT []string

	for _, se := range civ.StartEntities {
		queueE = append(queueE, se.Template)
	}

	res := &ReachResult{Techs: map[string]*tech.Technology{}}

	for len(queueE) > 0 || len(queueT) > 0 {
		// Process all entities first, then all techs (as per spec).
		for len(queueE) > 0 {
			tok := queueE[0]
			queueE = queueE[1:]

			tok = strings.TrimSpace(tok)
			if tok == "" || strings.HasPrefix(tok, "-") {
				continue
			}
			tok = tmpl.SubstCiv(tok, civ.Code)

			if _, ok := seen[tok]; ok {
				continue
			}
			seen[tok] = struct{}{}

			path, ok := idx.Lookup(tok)
			if !ok {
				res.Skipped = append(res.Skipped, SkipNote{tok, "template not found"})
				continue
			}
			// Lookup returns a candidate path for slash-refs even when the file
			// does not exist. Verify existence before resolving so that absent
			// civ-specific templates (e.g. structures/{civ}/crannog for civs
			// that don't have a crannog) are treated as a normal skip.
			if _, statErr := os.Stat(path); errors.Is(statErr, fs.ErrNotExist) {
				res.Skipped = append(res.Skipped, SkipNote{tok, "template not found"})
				continue
			}
			el, err := resolver.Resolve(path)
			if err != nil {
				return nil, err
			}

			ent := Entity{TemplateID: tok, Path: path, Element: el}
			classifyAndAppend(res, ent)

			for _, t := range el.GetTokens("Trainer/Entities") {
				queueE = append(queueE, t)
			}
			for _, t := range el.GetTokens("Builder/Entities") {
				queueE = append(queueE, t)
			}
			for _, t := range el.GetTokens("ProductionQueue/Entities") {
				queueE = append(queueE, t)
			}
			for _, t := range el.GetTokens("Trainer/Technologies") {
				queueT = append(queueT, t)
			}
			for _, t := range el.GetTokens("ProductionQueue/Technologies") {
				queueT = append(queueT, t)
			}
			for _, t := range el.GetTokens("Researcher/Technologies") {
				queueT = append(queueT, t)
			}

			// WallSet/Templates children are named elements, not space-separated tokens.
			if ws := el.Get("WallSet/Templates"); ws != nil {
				for _, child := range ws.Children {
					if v := strings.TrimSpace(child.Text); v != "" {
						queueE = append(queueE, v)
					}
				}
			}
		}

		for len(queueT) > 0 {
			t := queueT[0]
			queueT = queueT[1:]

			t = strings.TrimSpace(t)
			if t == "" || strings.HasPrefix(t, "-") {
				continue
			}
			if _, ok := seenT[t]; ok {
				continue
			}
			seenT[t] = struct{}{}

			techRec, err := catalog.ByName(t)
			if err != nil {
				res.Skipped = append(res.Skipped, SkipNote{t, "tech not in catalog"})
				continue
			}

			// Only pair wrappers have both Top and Bottom set; sub-techs that
			// carry a back-pointer (t.Pair != "") arriving directly in a tech
			// list are handled as plain techs.
			if techRec.Top != "" && techRec.Bottom != "" {
				top, bot, ok := tech.ExpandPair(catalog, t)
				if !ok {
					res.Skipped = append(res.Skipped, SkipNote{t, "pair expansion failed"})
					continue
				}
				res.Techs[top.Name] = top
				res.Techs[bot.Name] = bot
				continue
			}

			res.Techs[t] = techRec
		}
	}

	return res, nil
}

// classifyAndAppend appends ent to the appropriate slice in res based on its
// TemplateID prefix. Entities with other prefixes (e.g. gaia/) are ignored.
func classifyAndAppend(res *ReachResult, e Entity) {
	switch {
	case strings.HasPrefix(e.TemplateID, "structures/"):
		res.Buildings = append(res.Buildings, e)
	case strings.HasPrefix(e.TemplateID, "units/"):
		res.Units = append(res.Units, e)
		// Other prefixes (gaia/, etc.) are not included in the report.
	}
}
