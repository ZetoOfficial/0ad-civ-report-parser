package aura

import (
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

// LoadTeamBonus reads data/auras/teambonuses/<civ>_player_teambonus.json
// and returns it as an *Aura. This is the structured source for the
// "Командный бонус" block in the civilization overview.
//
// If the file is missing, the returned error wraps *os.PathError; callers
// can detect absence with errors.Is(err, fs.ErrNotExist). In R28 every
// civ has exactly one team bonus, so callers may treat absence as "civ
// has no team bonus" without further branching.
func LoadTeamBonus(layout paths.Layout, civCode string) (*Aura, error) {
	return Load(layout.TeamBonus(civCode))
}
