package civdata

import (
	"sort"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

// WallPiece is a single piece of a wallset (gate, long wall, tower, etc.).
type WallPiece struct {
	Role   string
	Entity Entity
}

// WallSetGroup groups a wallset wrapper entity with its piece entities.
type WallSetGroup struct {
	Wrapper Entity
	Pieces  []WallPiece
	Phase   Phase
}

// IdentifyWallSets scans buildings for wallset wrappers (entities with a
// non-empty WallSet/Templates child block), resolves their pieces against
// the same buildings slice, and returns the wallset groups together with
// a filtered buildings slice that excludes both wrappers and pieces.
//
// civCode is needed because piece references in shared wrapper templates
// (e.g. structures/wallset_palisade.xml) contain {civ}-placeholders that
// must be substituted before lookup.
//
// Invariant: groups are returned sorted by BuildingSortKey(Wrapper) for
// deterministic rendering. Callers must not re-sort.
func IdentifyWallSets(buildings []Entity, civCode string) (groups []*WallSetGroup, filtered []Entity) {
	// Build lookup map by TemplateID.
	byID := make(map[string]Entity, len(buildings))
	for _, e := range buildings {
		byID[e.TemplateID] = e
	}

	removeIDs := make(map[string]struct{})

	for _, e := range buildings {
		ws := e.Element.Get("WallSet/Templates")
		if ws == nil || len(ws.Children) == 0 {
			continue
		}

		g := &WallSetGroup{
			Wrapper: e,
			Phase:   BuildingPhase(e),
		}
		removeIDs[e.TemplateID] = struct{}{}

		for _, child := range ws.Children {
			role := child.Name
			if role == "" {
				continue
			}
			tokenRaw := strings.TrimSpace(child.Text)
			if tokenRaw == "" {
				continue
			}
			tok := tmpl.SubstCiv(tokenRaw, civCode)
			// Pieces missing from byID are silently skipped: Reach already
			// reported them via res.Skipped, so no extra logging is needed.
			if piece, ok := byID[tok]; ok {
				g.Pieces = append(g.Pieces, WallPiece{Role: role, Entity: piece})
				removeIDs[piece.TemplateID] = struct{}{}
			}
		}

		groups = append(groups, g)
	}

	// Sort groups deterministically by wrapper's sort key.
	sort.SliceStable(groups, func(i, j int) bool {
		ai, _ := BuildingSortKey(groups[i].Wrapper)
		aj, _ := BuildingSortKey(groups[j].Wrapper)
		if ai != aj {
			return ai < aj
		}
		return groups[i].Wrapper.Basename() < groups[j].Wrapper.Basename()
	})

	// Build filtered slice, preserving order.
	filtered = make([]Entity, 0, len(buildings))
	for _, e := range buildings {
		if _, skip := removeIDs[e.TemplateID]; !skip {
			filtered = append(filtered, e)
		}
	}

	return groups, filtered
}
