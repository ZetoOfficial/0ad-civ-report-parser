package civdata

import (
	"fmt"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

type PlayerTemplate struct {
	GenericName string
	History     string
	IconPath    string
}

// LoadPlayerTemplate resolves templates/special/players/<civ>.xml through
// the given resolver (so parent="template_player" inheritance is applied),
// then extracts Identity/{GenericName,History,Icon}.
//
// History fields in the source files store paragraph breaks as the
// literal two-character sequence "\n". We convert that to real markdown
// paragraph breaks ("\n\n") so the rendered overview reads naturally.
func LoadPlayerTemplate(layout paths.Layout, civCode string,
	resolver *tmpl.Resolver) (*PlayerTemplate, error) {
	path := layout.PlayerTemplate(civCode)
	el, err := resolver.Resolve(path)
	if err != nil {
		return nil, fmt.Errorf("resolve player template %s: %w", path, err)
	}
	return &PlayerTemplate{
		GenericName: el.GetText("Identity/GenericName"),
		History:     unescapeNewlines(el.GetText("Identity/History")),
		IconPath:    el.GetText("Identity/Icon"),
	}, nil
}

// unescapeNewlines converts the literal two-character sequence \n into a
// markdown paragraph break (real \n\n). It only touches that exact pattern;
// real newlines from the XML stay untouched.
func unescapeNewlines(s string) string {
	return strings.ReplaceAll(s, `\n`, "\n\n")
}
