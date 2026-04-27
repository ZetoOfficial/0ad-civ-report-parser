package i18n

import (
	"fmt"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
)

// TechDisplayName returns the human-readable tech name for a given civ.
// If specificName has an entry for the civ code, it is appended in
// parentheses ("Town Phase (локально: Astiteia)"). Falls back to
// GenericName, then Name.
func TechDisplayName(t *tech.Technology, civCode string) string {
	if t == nil {
		return ""
	}
	name := t.GenericName
	if name == "" {
		name = t.Name
	}
	if t.SpecificName == nil {
		return name
	}
	v, ok := t.SpecificName[civCode]
	if !ok {
		return name
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return name
	}
	return fmt.Sprintf("%s (локально: %s)", name, s)
}
