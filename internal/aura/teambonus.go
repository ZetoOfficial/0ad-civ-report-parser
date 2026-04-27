package aura

import (
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

// LoadTeamBonus reads data/auras/teambonuses/<civ>_player_teambonus.json
// and returns it as an *Aura. This is the structured source for the
// "Командный бонус" block in the civilization overview.
//
// Returns an os.PathError if the file is missing — callers may treat
// missing as "civ has no team bonus" if appropriate. In R28 every civ
// has exactly one team bonus.
func LoadTeamBonus(layout paths.Layout, civCode string) (*Aura, error) {
	return Load(layout.TeamBonus(civCode))
}
