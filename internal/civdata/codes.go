package civdata

import "strings"

type CivCode struct {
	Code     string
	NameEN   string
	NameRU   string
	BaseName string
	Aliases  []string
}

func (c CivCode) OverviewFile() string  { return c.BaseName + "_overview.md" }
func (c CivCode) StructreeFile() string { return c.BaseName + "_structree.md" }

var Civilizations = []CivCode{
	{"athen", "Athenians", "Афиняне", "athenians", []string{"афин", "athens"}},
	{"brit", "Britons", "Бритты", "britons", []string{"брит", "britons"}},
	{"cart", "Carthaginians", "Карфагеняне", "carthaginians", []string{"карфаг"}},
	{"gaul", "Gauls", "Галлы", "gauls", []string{"галл"}},
	{"germ", "Germans", "Германцы", "germans", []string{"герман", "german"}},
	{"han", "Han Chinese", "Хань", "han", []string{"хан", "китай", "chinese"}},
	{"iber", "Iberians", "Иберы", "iberians", []string{"ибер", "iberians"}},
	{"kush", "Kushites", "Кушиты", "kushites", []string{"куш", "kushite"}},
	{"mace", "Macedonians", "Македоняне", "macedonians", []string{"македон", "macedon"}},
	{"maur", "Mauryas", "Маурьи", "mauryas", []string{"маурь", "маур", "индиец", "mauryas"}},
	{"pers", "Achaemenid Persians", "Персы", "persians", []string{"перс", "persian", "ахеменид"}},
	{"ptol", "Ptolemies", "Птолемеи", "ptolemies", []string{"птолеме", "ptolemy", "египт"}},
	{"rome", "Romans", "Римляне", "romans", []string{"рим", "romans"}},
	{"sele", "Seleucids", "Селевкиды", "seleucids", []string{"селевкид", "seleucid"}},
	{"spart", "Spartans", "Спартанцы", "spartans", []string{"спарт", "spartan"}},
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
