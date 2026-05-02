package i18n

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func ResourceName(key string) string {
	switch key {
	case "food":
		return "еды"
	case "wood":
		return "дерева"
	case "stone":
		return "камня"
	case "metal":
		return "металла"
	}
	return key
}

func DamageType(key string) string {
	switch key {
	case "Hack":
		return "Hack"
	case "Pierce":
		return "Pierce"
	case "Crush":
		return "Crush"
	case "Capture":
		return "Capture"
	case "Fire":
		return "Fire"
	}
	return key
}

func PhaseRequirement(req string) string {
	switch req {
	case "":
		return ""
	case "phase_village":
		return "Village"
	case "phase_town", "phase_town_generic":
		return "Town"
	case "phase_city", "phase_city_generic":
		return "City"
	}
	if strings.HasPrefix(req, "phase_town") {
		return "Town"
	}
	if strings.HasPrefix(req, "phase_city") {
		return "City"
	}
	return req
}

func FormatNumber(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	// Round to 6 decimals to eliminate IEEE 754 noise from arithmetic on
	// parsed floats (e.g. WalkSpeed 9 × RunMultiplier 1.2 → 10.8, not
	// 10.799999999999999). Six decimals preserve sub-percent precision
	// found in template data while collapsing binary-fraction tails.
	rounded := math.Round(v*1e6) / 1e6
	if rounded == float64(int64(rounded)) {
		return strconv.FormatInt(int64(rounded), 10)
	}
	return strconv.FormatFloat(rounded, 'f', -1, 64)
}

func FormatPercent(multiplier float64) string {
	delta := (multiplier - 1) * 100
	sign := "+"
	if delta < 0 {
		sign = "−"
		delta = -delta
	}
	// Round to 6 decimal places to eliminate IEEE 754 noise
	// (e.g. multiplier 1.1 → delta 10.000000000000009 → 10).
	rounded := math.Round(delta*1e6) / 1e6
	return fmt.Sprintf("%s%s%%", sign, FormatNumber(rounded))
}
