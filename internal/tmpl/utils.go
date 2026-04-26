package tmpl

import "strings"

func ParseFloatTrim(s string) (float64, bool) {
	return parseFloat(strings.TrimSpace(s))
}

func ParseInt(s string) (int, bool) {
	v, ok := parseFloat(strings.TrimSpace(s))
	if !ok {
		return 0, false
	}
	return int(v), true
}
