package civdata

import "strings"

type CivCode struct {
	Code       string
	NameEN     string
	NameRU     string
	OutputFile string
	Aliases    []string
}

var Civilizations = []CivCode{
	{"athen", "Athenians", "Афиняне", "athenians_buildings_report.md", []string{"афин", "athens"}},
	{"brit", "Britons", "Бритты", "britons_buildings_report.md", []string{"брит", "britons"}},
	{"cart", "Carthaginians", "Карфагеняне", "carthaginians_buildings_report.md", []string{"карфаг"}},
	{"gaul", "Gauls", "Галлы", "gauls_buildings_report.md", []string{"галл"}},
	{"germ", "Germans", "Германцы", "germans_buildings_report.md", []string{"герман", "german"}},
	{"han", "Han Chinese", "Хань", "han_buildings_report.md", []string{"хан", "китай", "chinese"}},
	{"iber", "Iberians", "Иберы", "iberians_buildings_report.md", []string{"ибер", "iberians"}},
	{"kush", "Kushites", "Кушиты", "kushites_buildings_report.md", []string{"куш", "kushite"}},
	{"mace", "Macedonians", "Македоняне", "macedonians_buildings_report.md", []string{"македон", "macedon"}},
	{"maur", "Mauryas", "Маурьи", "mauryas_buildings_report.md", []string{"маурь", "маур", "индиец", "mauryas"}},
	{"pers", "Achaemenid Persians", "Персы", "persians_buildings_report.md", []string{"перс", "persian", "ахеменид"}},
	{"ptol", "Ptolemies", "Птолемеи", "ptolemies_buildings_report.md", []string{"птолеме", "ptolemy", "египт"}},
	{"rome", "Romans", "Римляне", "romans_buildings_report.md", []string{"рим", "romans"}},
	{"sele", "Seleucids", "Селевкиды", "seleucids_buildings_report.md", []string{"селевкид", "seleucid"}},
	{"spart", "Spartans", "Спартанцы", "spartans_buildings_report.md", []string{"спарт", "spartan"}},
}

func ResolveCivInput(input string) (CivCode, bool) {
	if input == "" {
		return CivCode{}, false
	}
	low := strings.ToLower(strings.TrimSpace(input))
	for _, c := range Civilizations {
		if c.Code == low {
			return c, true
		}
	}
	for _, c := range Civilizations {
		for _, alias := range c.Aliases {
			if strings.Contains(low, strings.ToLower(alias)) {
				return c, true
			}
		}
	}
	return CivCode{}, false
}

func ByCode(code string) (CivCode, bool) {
	for _, c := range Civilizations {
		if c.Code == code {
			return c, true
		}
	}
	return CivCode{}, false
}
