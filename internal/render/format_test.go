package render

import (
	"reflect"
	"strings"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/tmpl"
)

// mkEl builds an Element with the given name and children.
func mkEl(name string, children ...*tmpl.Element) *tmpl.Element {
	return &tmpl.Element{Name: name, Children: children}
}

// mkText builds a leaf Element with the given name and text content.
func mkText(name, text string) *tmpl.Element {
	return &tmpl.Element{Name: name, Text: text}
}

func TestFormatAttackBonuses(t *testing.T) {
	tests := []struct {
		name   string
		input  *tmpl.Element
		expect string
	}{
		{
			name:   "nil input",
			input:  nil,
			expect: "",
		},
		{
			name:   "no bonuses node",
			input:  mkEl("Melee"),
			expect: "",
		},
		{
			name: "single bonus single class",
			input: mkEl("Melee",
				mkEl("Bonuses",
					mkEl("BonusCav",
						mkText("Classes", "Cavalry"),
						mkText("Multiplier", "2.5"),
					),
				),
			),
			expect: "×2.5 vs Cavalry",
		},
		{
			name: "single bonus multiple classes",
			input: mkEl("Melee",
				mkEl("Bonuses",
					mkEl("BonusCavEl",
						mkText("Classes", "Cavalry Elephant"),
						mkText("Multiplier", "3"),
					),
				),
			),
			expect: "×3 vs Cavalry+Elephant",
		},
		{
			name: "multiple bonuses",
			input: mkEl("Melee",
				mkEl("Bonuses",
					mkEl("BonusCav",
						mkText("Classes", "Cavalry"),
						mkText("Multiplier", "2.5"),
					),
					mkEl("BonusEl",
						mkText("Classes", "Elephant"),
						mkText("Multiplier", "2"),
					),
				),
			),
			expect: "×2.5 vs Cavalry, ×2 vs Elephant",
		},
		{
			name: "bonus missing Multiplier skipped",
			input: mkEl("Melee",
				mkEl("Bonuses",
					mkEl("BonusCav",
						mkText("Classes", "Cavalry"),
					),
				),
			),
			expect: "",
		},
		{
			name: "bonus missing Classes skipped",
			input: mkEl("Melee",
				mkEl("Bonuses",
					mkEl("BonusCav",
						mkText("Multiplier", "2"),
					),
				),
			),
			expect: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatAttackBonuses(tc.input)
			if got != tc.expect {
				t.Errorf("formatAttackBonuses() = %q, want %q", got, tc.expect)
			}
		})
	}
}

func TestFormatPreferredClasses(t *testing.T) {
	tests := []struct {
		name   string
		input  *tmpl.Element
		expect string
	}{
		{
			name:   "nil input",
			input:  nil,
			expect: "",
		},
		{
			name:   "missing PreferredClasses",
			input:  mkEl("Melee"),
			expect: "",
		},
		{
			name: "one class",
			input: mkEl("Melee",
				mkText("PreferredClasses", "Human"),
			),
			expect: "Human",
		},
		{
			name: "multiple classes",
			input: mkEl("Melee",
				mkText("PreferredClasses", "Human Cavalry"),
			),
			expect: "Human, Cavalry",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatPreferredClasses(tc.input)
			if got != tc.expect {
				t.Errorf("formatPreferredClasses() = %q, want %q", got, tc.expect)
			}
		})
	}
}

func TestFormatSplash(t *testing.T) {
	tests := []struct {
		name     string
		input    *tmpl.Element
		expect   string
		contains string // if non-empty, check substring instead of exact match
	}{
		{
			name:   "nil input",
			input:  nil,
			expect: "",
		},
		{
			name:   "no splash node",
			input:  mkEl("Ranged"),
			expect: "",
		},
		{
			name: "full canonical Circular friendly-fire false",
			input: mkEl("Ranged",
				mkEl("Splash",
					mkText("Shape", "Circular"),
					mkText("Range", "1.5"),
					mkText("FriendlyFire", "false"),
					mkEl("Damage",
						mkText("Crush", "120"),
					),
				),
			),
			expect: "Crush 120, круг R=1.5, не задевает союзников",
		},
		{
			name: "default friendly fire (missing FriendlyFire)",
			input: mkEl("Ranged",
				mkEl("Splash",
					mkText("Shape", "Circular"),
					mkText("Range", "2"),
					mkEl("Damage",
						mkText("Pierce", "50"),
					),
				),
			),
			contains: "задевает союзников",
		},
		{
			name: "linear shape",
			input: mkEl("Ranged",
				mkEl("Splash",
					mkText("Shape", "Linear"),
					mkText("Range", "3"),
					mkEl("Damage",
						mkText("Hack", "10"),
					),
				),
			),
			contains: "линия",
		},
		{
			name: "with own Bonuses",
			input: mkEl("Ranged",
				mkEl("Splash",
					mkText("Shape", "Circular"),
					mkText("Range", "2"),
					mkEl("Damage",
						mkText("Crush", "80"),
					),
					mkEl("Bonuses",
						mkEl("BonusClass",
							mkText("Classes", "Class"),
							mkText("Multiplier", "2"),
						),
					),
				),
			),
			contains: "×2 vs Class",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatSplash(tc.input)
			if tc.contains != "" {
				if !strings.Contains(got, tc.contains) {
					t.Errorf("formatSplash() = %q, want substring %q", got, tc.contains)
				}
				return
			}
			if got != tc.expect {
				t.Errorf("formatSplash() = %q, want %q", got, tc.expect)
			}
		})
	}
}

func TestFormatCaptureAttack(t *testing.T) {
	tests := []struct {
		name     string
		input    *tmpl.Element
		expect   string
		contains string
	}{
		{
			name:   "nil input",
			input:  nil,
			expect: "",
		},
		{
			name:   "no capture node",
			input:  mkEl("Attack"),
			expect: "",
		},
		{
			name: "missing Capture rate",
			input: mkEl("Attack",
				mkEl("Capture"),
			),
			expect: "",
		},
		{
			name: "full canonical",
			input: mkEl("Attack",
				mkEl("Capture",
					mkText("Capture", "2.5"),
					mkText("MaxRange", "4"),
					mkText("RepeatTime", "1000"),
					mkText("RestrictedClasses", "Field Palisade Wall"),
				),
			),
			expect: "захват 2.5 (4м, 1000мс); исключает: Field, Palisade, Wall",
		},
		{
			name: "no RestrictedClasses",
			input: mkEl("Attack",
				mkEl("Capture",
					mkText("Capture", "3"),
					mkText("MaxRange", "5"),
					mkText("RepeatTime", "750"),
				),
			),
			expect: "захват 3 (5м, 750мс)",
		},
		{
			name: "partial only MaxRange",
			input: mkEl("Attack",
				mkEl("Capture",
					mkText("Capture", "1"),
					mkText("MaxRange", "4"),
				),
			),
			contains: "(4м)",
		},
		{
			name: "partial only RepeatTime",
			input: mkEl("Attack",
				mkEl("Capture",
					mkText("Capture", "1"),
					mkText("RepeatTime", "500"),
				),
			),
			contains: "(500мс)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatCaptureAttack(tc.input)
			if tc.contains != "" {
				if !strings.Contains(got, tc.contains) {
					t.Errorf("formatCaptureAttack() = %q, want substring %q", got, tc.contains)
				}
				return
			}
			if got != tc.expect {
				t.Errorf("formatCaptureAttack() = %q, want %q", got, tc.expect)
			}
		})
	}
}

func TestFormatApplyStatuses(t *testing.T) {
	byCode := map[string]statusEffect{
		"Poisoned": {Code: "Poisoned", StatusName: "Poisoned"},
		"Burning":  {Code: "Burning", StatusName: "Burning"},
	}

	// Helper to build an ApplyStatus child element.
	makePoisoned := func(dmgPoison, interval, duration string) *tmpl.Element {
		children := []*tmpl.Element{}
		if dmgPoison != "" {
			children = append(children, mkEl("Damage", mkText("Poison", dmgPoison)))
		}
		if interval != "" {
			children = append(children, mkText("Interval", interval))
		}
		if duration != "" {
			children = append(children, mkText("Duration", duration))
		}
		return mkEl("Poisoned", children...)
	}

	makeBurning := func(dmgFire, interval, duration string) *tmpl.Element {
		children := []*tmpl.Element{}
		if dmgFire != "" {
			children = append(children, mkEl("Damage", mkText("Fire", dmgFire)))
		}
		if interval != "" {
			children = append(children, mkText("Interval", interval))
		}
		if duration != "" {
			children = append(children, mkText("Duration", duration))
		}
		return mkEl("Burning", children...)
	}

	tests := []struct {
		name   string
		modeEl *tmpl.Element
		want   []string
	}{
		{
			name:   "no apply status node",
			modeEl: mkEl("Ranged"),
			want:   nil,
		},
		{
			name: "Maur Maiden Archer canonical",
			modeEl: mkEl("Ranged",
				mkEl("ApplyStatus",
					makePoisoned("2", "600", "3600"),
				),
			),
			want: []string{"Poisoned: Poison 2/0.6с × 3.6с (см. common.md#poisoned)"},
		},
		{
			name: "Iber Champion Cavalry Burning",
			modeEl: mkEl("Ranged",
				mkEl("ApplyStatus",
					makeBurning("2", "3000", "9000"),
				),
			),
			want: []string{"Burning: Fire 2/3с × 9с (см. common.md#burning)"},
		},
		{
			name: "BlockChance no-op default",
			modeEl: mkEl("Ranged",
				mkEl("ApplyStatus",
					mkEl("Poisoned",
						mkText("BlockChance", "1"),
						mkText("Duration", "0"),
					),
				),
			),
			want: nil,
		},
		{
			name: "zero duration without BlockChance",
			modeEl: mkEl("Ranged",
				mkEl("ApplyStatus",
					mkEl("Poisoned",
						mkEl("Damage", mkText("Poison", "5")),
						mkText("Interval", "1000"),
						mkText("Duration", "0"),
					),
				),
			),
			want: nil,
		},
		{
			name: "multiple statuses",
			modeEl: mkEl("Ranged",
				mkEl("ApplyStatus",
					makePoisoned("2", "600", "3600"),
					makeBurning("3", "1000", "5000"),
				),
			),
			want: []string{
				"Poisoned: Poison 2/0.6с × 3.6с (см. common.md#poisoned)",
				"Burning: Fire 3/1с × 5с (см. common.md#burning)",
			},
		},
		{
			name: "missing Damage",
			modeEl: mkEl("Ranged",
				mkEl("ApplyStatus",
					mkEl("Poisoned",
						mkText("Interval", "1000"),
						mkText("Duration", "5000"),
					),
				),
			),
			want: []string{"Poisoned: каждые 1с × 5с (см. common.md#poisoned)"},
		},
		{
			name: "unknown code falls back to raw",
			modeEl: mkEl("Ranged",
				mkEl("ApplyStatus",
					mkEl("Frenzied",
						mkEl("Damage", mkText("Hack", "10")),
						mkText("Interval", "500"),
						mkText("Duration", "2000"),
					),
				),
			),
			want: []string{"Frenzied: Hack 10/0.5с × 2с (см. common.md#frenzied)"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatApplyStatuses(tc.modeEl, byCode)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("formatApplyStatuses() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestFormatApplyStatuses_NilInput covers the nil modeEl branch separately.
func TestFormatApplyStatuses_NilInput(t *testing.T) {
	got := formatApplyStatuses(nil, nil)
	if got != nil {
		t.Errorf("formatApplyStatuses(nil, nil) = %v, want nil", got)
	}
}


