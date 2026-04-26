package i18n

import (
	"fmt"
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
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func FormatPercent(multiplier float64) string {
	delta := (multiplier - 1) * 100
	sign := "+"
	if delta < 0 {
		sign = "−"
		delta = -delta
	}
	return fmt.Sprintf("%s%s%%", sign, FormatNumber(delta))
}
