package render

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/aura"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

type Generator struct {
	Layout            paths.Layout
	Resolver          *tmpl.Resolver
	Catalog           *tech.Catalog
	Index             *tech.Index // lazily built in Generate; set manually in tests if needed
	IncludeHistory    bool
	statusEffects     map[string]statusEffect // code → effect, lazily loaded in Generate
	statusEffectsList []statusEffect          // sorted slice; mirrors statusEffects, populated together
}

// Output holds the rendered markdown bodies for one civilization.
// They do not yet contain the skeleton header — main.go wraps them
// via internal/render/skeleton.
type Output struct {
	Overview  string
	Structree string
}

func NewGenerator(layout paths.Layout, resolver *tmpl.Resolver) *Generator {
	return &Generator{
		Layout:   layout,
		Resolver: resolver,
		Catalog:  tech.NewCatalog(layout.Technologies()),
	}
}

func (g *Generator) Generate(civInfo civdata.CivCode) (Output, error) {
	civ, err := civdata.LoadCiv(g.Layout.CivJSON(civInfo.Code))
	if err != nil {
		return Output{}, err
	}
	if g.Index == nil {
		g.Index, err = tech.NewIndex(g.Catalog)
		if err != nil {
			return Output{}, err
		}
	}
	res, err := civdata.Reach(civ, g.Resolver.Index, g.Resolver, g.Catalog)
	if err != nil {
		return Output{}, err
	}
	buildings := res.Buildings
	units := res.Units
	bonuses, err := g.Catalog.AllCivBonuses(civInfo.Code)
	if err != nil {
		return Output{}, err
	}
	// civSpecific covers root technologies/ files with civ-lock too, not just civbonuses/.
	civSpecific := g.Index.CivSpecific(civInfo.Code)
	notciv, err := g.Catalog.AllNotCiv(civInfo.Code)
	if err != nil {
		return Output{}, err
	}
	heroAuras, _ := aura.ListInDir(g.Layout.HeroAuras(), civInfo.Code+"_hero_")
	catafalqueAuras, _ := aura.ListInDir(g.Layout.CatafalqueAuras(), civInfo.Code+"_")

	// New in epic 2: Player template (Identity), team bonus aura.
	// LoadPlayerTemplate failure is fatal — every civ has a Player template.
	player, err := civdata.LoadPlayerTemplate(g.Layout, civInfo.Code, g.Resolver)
	if err != nil {
		return Output{}, err
	}
	// LoadTeamBonus may legitimately be missing for civs without one in
	// future R-versions, but in R28 every civ has one. Use errors.Is to
	// traverse the %w wrap chain inside aura.Load.
	teamBonus, err := aura.LoadTeamBonus(g.Layout, civInfo.Code)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Output{}, err
	}

	if _, err := g.loadStatusEffectsCached(); err != nil {
		return Output{}, err
	}

	return Output{
		Overview:  g.renderOverview(civInfo, civ, player, teamBonus, bonuses, civSpecific, notciv, units, buildings, heroAuras),
		Structree: g.renderStructree(civInfo.Code, res, heroAuras, catafalqueAuras),
	}, nil
}

// loadStatusEffectsCached returns the sorted slice of status effects and, as
// a side-effect, populates g.statusEffects (the code→effect map). Both the
// slice and the map are cached on the Generator after the first call;
// subsequent calls return them without touching disk. Important under --all,
// where the same Generator services 15 civs plus common.md.
func (g *Generator) loadStatusEffectsCached() ([]statusEffect, error) {
	if g.statusEffectsList != nil {
		return g.statusEffectsList, nil
	}
	list, err := loadStatusEffects(g.Layout.StatusEffects())
	if err != nil {
		return nil, err
	}
	g.statusEffects = make(map[string]statusEffect, len(list))
	for _, e := range list {
		g.statusEffects[e.Code] = e
	}
	g.statusEffectsList = list
	return list, nil
}

// RenderCommon returns the body of the shared common.md (without the
// skeleton wrapper). See internal/render/common.go for section layout.
func (g *Generator) RenderCommon() (string, error) {
	return g.renderCommonBody()
}

func formatStartEntities(entities []civdata.StartEntity) string {
	parts := []string{}
	for _, e := range entities {
		base := filepath.Base(e.Template)
		name := strings.TrimPrefix(base, "structures/")
		name = strings.TrimPrefix(name, "units/")
		count := e.Count
		if count == 0 {
			count = 1
		}
		parts = append(parts, name+" ×"+itoa(count))
	}
	return strings.Join(parts, ", ")
}

func itoa(n int) string {
	// kept local to avoid an extra strconv import for this small helper
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits [20]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		digits[i] = '-'
	}
	return string(digits[i:])
}

func escapeTable(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
