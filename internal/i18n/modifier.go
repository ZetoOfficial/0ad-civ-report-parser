package i18n

import (
	"fmt"
	"strings"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tech"
)

func DescribeModification(m tech.Modification) string {
	target := translatePath(m.Value)
	var body string
	switch {
	case m.Multiply != 0:
		body = fmt.Sprintf("%s %s", target, FormatPercent(m.Multiply))
	case m.Add != 0:
		sign := "+"
		val := m.Add
		if val < 0 {
			sign = "−"
			val = -val
		}
		body = fmt.Sprintf("%s %s%s", target, sign, FormatNumber(val))
	case m.Replace != nil:
		body = fmt.Sprintf("%s = %v", target, m.Replace)
	default:
		body = target
	}
	if affects := m.AffectsList(); len(affects) > 0 {
		body += fmt.Sprintf(" (только %s)", strings.Join(affects, "+"))
	}
	return body
}

func DescribeModifications(mods []tech.Modification) string {
	if len(mods) == 0 {
		return ""
	}
	parts := make([]string, 0, len(mods))
	for _, m := range mods {
		parts = append(parts, DescribeModification(m))
	}
	return strings.Join(parts, "; ")
}

var pathTranslations = map[string]string{
	"Health/Max":                         "ОЗ",
	"Health/RegenRate":                   "регенерация ОЗ",
	"Attack/Melee/Damage/Hack":           "рубящий урон ближнего боя",
	"Attack/Melee/Damage/Pierce":         "колющий урон ближнего боя",
	"Attack/Melee/Damage/Crush":          "дробящий урон ближнего боя",
	"Attack/Melee/RepeatTime":            "перезарядка ближнего боя",
	"Attack/Melee/MaxRange":              "дальность ближнего боя",
	"Attack/Ranged/Damage/Hack":          "рубящий урон стрельбы",
	"Attack/Ranged/Damage/Pierce":        "колющий урон стрельбы",
	"Attack/Ranged/Damage/Crush":         "дробящий урон стрельбы",
	"Attack/Ranged/MaxRange":             "дальность стрельбы",
	"Attack/Ranged/RepeatTime":           "перезарядка стрельбы",
	"Attack/Ranged/Projectile/Spread":    "разброс снаряда",
	"Attack/Capture/Capture":             "захват",
	"Resistance/Entity/Damage/Hack":      "защита от рубящего",
	"Resistance/Entity/Damage/Pierce":    "защита от колющего",
	"Resistance/Entity/Damage/Crush":     "защита от дробящего",
	"UnitMotion/WalkSpeed":               "скорость ходьбы",
	"UnitMotion/RunMultiplier":           "множитель бега",
	"Vision/Range":                       "обзор",
	"Cost/BuildTime":                     "время постройки",
	"Cost/Population":                    "стоимость населения",
	"Cost/PopulationBonus":               "бонус населения",
	"Cost/Resources/wood":                "стоимость дерева",
	"Cost/Resources/stone":               "стоимость камня",
	"Cost/Resources/metal":               "стоимость металла",
	"Cost/Resources/food":                "стоимость еды",
	"ResourceGatherer/BaseSpeed":         "скорость сбора",
	"ResourceGatherer/Rates/food.fish":   "сбор рыбы",
	"ResourceGatherer/Rates/food.fruit":  "сбор фруктов",
	"ResourceGatherer/Rates/food.grain":  "сбор зерна",
	"ResourceGatherer/Rates/food.meat":   "сбор мяса",
	"ResourceGatherer/Rates/wood.tree":   "рубка деревьев",
	"ResourceGatherer/Rates/wood.ruins":  "сбор древесины из руин",
	"ResourceGatherer/Rates/stone.rock":  "добыча камня",
	"ResourceGatherer/Rates/stone.ruins": "добыча камня из руин",
	"ResourceGatherer/Rates/metal.ore":   "добыча металла",
	"ResourceGatherer/Capacities/food":   "вместимость еды",
	"ResourceGatherer/Capacities/wood":   "вместимость дерева",
	"ResourceGatherer/Capacities/stone":  "вместимость камня",
	"ResourceGatherer/Capacities/metal":  "вместимость металла",
	"GarrisonHolder/Max":                 "вместимость гарнизона",
	"BuildingAI/DefaultArrowCount":       "стрелы (базово)",
	"BuildingAI/GarrisonArrowMultiplier": "стрелы за гарнизонного юнита",
	"TerritoryInfluence/Radius":          "радиус влияния территории",
	"Capturable/CapturePoints":           "очки захвата",
	"ProductionQueue/TimeMultiplier":     "скорость производства",
	"Promotion/RequiredXp":               "требуемый опыт ранга",
	"Loot/xp":                            "опыт за убийство",
	"Loot/food":                          "доход еды за убийство",
	"Loot/wood":                          "доход дерева за убийство",
	"Loot/stone":                         "доход камня за убийство",
	"Loot/metal":                         "доход металла за убийство",
	"Heal/Range":                         "дальность лечения",
	"Heal/Health":                        "лечение ОЗ",
	"Heal/HP":                            "лечение HP",
	"Trader/GainMultiplier":              "множитель торговой прибыли",
	"Pack/Time":                          "время сборки/разборки",
}

func translatePath(value string) string {
	if v, ok := pathTranslations[value]; ok {
		return v
	}
	return value
}
