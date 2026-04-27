package paths

import (
	"path/filepath"
)

const DefaultGameDataRoot = "/Users/zeto/Projects/study/0ad/binaries/data/mods/public"

const EnvGameDataRoot = "OAD_GAMEDATA_ROOT"

type Layout struct {
	Root string
}

func (l Layout) Templates() string {
	return filepath.Join(l.Root, "simulation", "templates")
}

func (l Layout) Civs() string {
	return filepath.Join(l.Root, "simulation", "data", "civs")
}

func (l Layout) Technologies() string {
	return filepath.Join(l.Root, "simulation", "data", "technologies")
}

func (l Layout) CivBonuses() string {
	return filepath.Join(l.Root, "simulation", "data", "technologies", "civbonuses")
}
func (l Layout) Auras() string { return filepath.Join(l.Root, "simulation", "data", "auras") }
func (l Layout) HeroAuras() string {
	return filepath.Join(l.Root, "simulation", "data", "auras", "units", "heroes")
}

func (l Layout) CatafalqueAuras() string {
	return filepath.Join(l.Root, "simulation", "data", "auras", "units", "catafalques")
}

func (l Layout) StructureAuras() string {
	return filepath.Join(l.Root, "simulation", "data", "auras", "structures")
}

func (l Layout) StructuresOf(civ string) string {
	return filepath.Join(l.Templates(), "structures", civ)
}
func (l Layout) UnitsOf(civ string) string   { return filepath.Join(l.Templates(), "units", civ) }
func (l Layout) CivJSON(civ string) string   { return filepath.Join(l.Civs(), civ+".json") }
func (l Layout) TechJSON(name string) string { return filepath.Join(l.Technologies(), name+".json") }
func (l Layout) AuraJSON(rel string) string  { return filepath.Join(l.Auras(), rel+".json") }
func (l Layout) PlayerTemplate(civ string) string {
	return filepath.Join(l.Templates(), "special", "players", civ+".xml")
}

func (l Layout) TeamBonus(civ string) string {
	return filepath.Join(l.Auras(), "teambonuses", civ+"_player_teambonus.json")
}
