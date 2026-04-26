package render

import (
	"path/filepath"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/aura"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

type Generator struct {
	Layout   paths.Layout
	Resolver *tmpl.Resolver
	Catalog  *tech.Catalog
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
	buildings, err := civdata.Buildings(g.Layout.StructuresOf(civInfo.Code), civInfo.Code, g.Resolver)
	if err != nil {
		return Output{}, err
	}
	units, err := civdata.Units(g.Layout.UnitsOf(civInfo.Code), civInfo.Code, g.Resolver)
	if err != nil {
		return Output{}, err
	}
	bonuses, err := g.Catalog.AllCivBonuses(civInfo.Code)
	if err != nil {
		return Output{}, err
	}
	notciv, err := g.Catalog.AllNotCiv(civInfo.Code)
	if err != nil {
		return Output{}, err
	}
	heroAuras, _ := aura.ListInDir(g.Layout.HeroAuras(), civInfo.Code+"_hero_")
	catafalqueAuras, _ := aura.ListInDir(g.Layout.CatafalqueAuras(), civInfo.Code+"_")

	return Output{
		Overview:  g.renderOverview(civInfo, civ, bonuses, notciv),
		Structree: g.renderStructree(civInfo.Code, buildings, units, heroAuras, catafalqueAuras),
	}, nil
}

// RenderCommon returns the body of the shared common.md.
// In epic 1 this is a placeholder — populated in epic 2.
func (g *Generator) RenderCommon() (string, error) {
	return renderCommonBody(), nil
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
