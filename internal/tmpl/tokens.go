package tmpl

import "strings"

func SubstCiv(text, civ string) string {
	out := strings.ReplaceAll(text, "{civ}", civ)
	out = strings.ReplaceAll(out, "{native}", civ)
	return out
}

func SubstCivAll(toks []string, civ string) []string {
	out := make([]string, 0, len(toks))
	for _, t := range toks {
		out = append(out, SubstCiv(t, civ))
	}
	return out
}
