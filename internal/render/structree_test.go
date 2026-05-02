package render

import (
	"strings"
	"testing"

	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/civdata"
	"github.com/ZetoOfficial/0ad-civ-report-parser/internal/paths"
)

func TestPhaseRank(t *testing.T) {
	cases := []struct {
		label string
		want  int
	}{
		{"Village", 0},
		{"Town", 1},
		{"City", 2},
		{"—", 3},
		{"", 3},
		{"Unknown", 3},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			if got := phaseRank(c.label); got != c.want {
				t.Errorf("phaseRank(%q) = %d, want %d", c.label, got, c.want)
			}
		})
	}
}

func TestStructree_Germ_ResearchSortedByPhase(t *testing.T) {
	skipIfNoGamedata(t)
	out := generateFor(t, "germ")

	// Split on "#### Исследует" blocks; skip first chunk (pre-header text).
	blocks := strings.Split(out.Structree, "#### Исследует")
	if len(blocks) < 2 {
		t.Fatal("no #### Исследует blocks found in germ structree")
	}

	violated := false
	for i, block := range blocks[1:] {
		// Trim the block to its content before the next header or separator.
		cut := block
		for _, sep := range []string{"\n#", "\n---"} {
			if idx := strings.Index(cut, sep); idx != -1 {
				cut = cut[:idx]
			}
		}

		// Extract the phase column (4th pipe-delimited column) from each data row.
		var phases []string
		for _, line := range strings.Split(cut, "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "|") {
				continue
			}
			// Skip separator rows and the header row.
			if strings.HasPrefix(line, "|---") || strings.Contains(line, "| Технология |") {
				continue
			}
			cols := strings.Split(line, "|")
			// cols[0]="", cols[1]=name, cols[2]=cost, cols[3]=time, cols[4]=phase, cols[5]=effect, cols[6]=""
			if len(cols) < 6 {
				continue
			}
			phase := strings.TrimSpace(cols[4])
			if phase == "" {
				continue
			}
			phases = append(phases, phase)
		}

		// Verify that phase ranks are non-decreasing within the block.
		prev := -1
		for j, p := range phases {
			r := phaseRank(p)
			if r < prev {
				t.Errorf("block %d row %d: phase %q (rank %d) appears after rank %d", i, j, p, r, prev)
				violated = true
			}
			if r > prev {
				prev = r
			}
		}
	}
	if !violated {
		t.Logf("all #### Исследует blocks in germ structree are sorted by phase")
	}
}

func TestStructree_Germ_TwoWallSets(t *testing.T) {
	skipIfNoGamedata(t)
	out := generateFor(t, "germ")
	n := strings.Count(out.Structree, "### Стены")
	if n != 2 {
		t.Errorf("expected 2 wallset headers for germ, got %d", n)
	}
}

func TestStructree_Han_HasPairMarker(t *testing.T) {
	skipIfNoGamedata(t)
	out := generateFor(t, "han")
	if !strings.Contains(out.Structree, "◐") {
		t.Errorf("expected pair marker ◐ in han structree")
	}
}

func TestStructree_Athen_PhaseTownAthen(t *testing.T) {
	skipIfNoGamedata(t)
	out := generateFor(t, "athen")
	hasName := strings.Contains(out.Structree, "phase_town_athen") ||
		strings.Contains(out.Structree, "Kōmopolis")
	if !hasName {
		t.Errorf("athen structree missing phase_town_athen / Kōmopolis row")
	}
	if strings.Contains(out.Structree, "phase_town_generic") {
		t.Errorf("athen structree must not include phase_town_generic")
	}
	// Phase column for phase_town_athen must resolve to "Town" via Supersedes fallback,
	// not fall back to "—" (which happens when requirements.entity is used instead of
	// requirements.tech).
	if !strings.Contains(out.Structree, "Town") {
		t.Errorf("athen structree missing Town label for phase_town_athen (Supersedes fallback not working)")
	}
	// Regression: chainSuffix должен показывать generic "Town Phase" в "заменяет",
	// а НЕ civ-локальный "Town Phase (локально: Kōmopolis)".
	if strings.Contains(out.Structree, "заменяет: Town Phase (локально:") {
		t.Errorf("athen structree shows self-referential display in chainSuffix (заменяет должен быть generic)")
	}
}

func TestStructree_Germ_NoForeignTechs(t *testing.T) {
	skipIfNoGamedata(t)
	out := generateFor(t, "germ")
	forbidden := []string{"Roman Roads", "Hoplite Tradition", "Archery Tradition", "Hellenistic Metropolis"}
	for _, name := range forbidden {
		if strings.Contains(out.Structree, name) {
			t.Errorf("germ structree must NOT contain %q (civ-blocked tech)", name)
		}
	}
}

func TestOverview_Germ_HasCivSpecificTechs(t *testing.T) {
	skipIfNoGamedata(t)
	out := generateFor(t, "germ")
	expected := []string{"Migratory Resettlement", "Grove of Fetters"}
	for _, name := range expected {
		if !strings.Contains(out.Overview, name) {
			t.Errorf("germ overview must contain unique tech %q", name)
		}
	}
}

func TestStructree_Germ_NoPhaseTokenLeak(t *testing.T) {
	skipIfNoGamedata(t)
	out := generateFor(t, "germ")
	forbidden := []string{
		"phase_town_germ",
		"phase_city_germ",
	}
	for _, name := range forbidden {
		if strings.Contains(out.Structree, name) {
			t.Errorf("germ structree leaks raw phase token %q (should resolve to generic)", name)
		}
	}
}

func TestStructree_Athen_PhaseTokenResolvesCivVariant(t *testing.T) {
	skipIfNoGamedata(t)
	out := generateFor(t, "athen")
	// For athen, phase_town_athen.json EXISTS and should be displayed
	// by display name (Town Phase / Kōmopolis), not by raw token.
	if strings.Contains(out.Structree, "| phase_town_athen |") {
		t.Errorf("athen structree shows raw phase_town_athen token instead of display name")
	}
	if strings.Contains(out.Structree, "| phase_town_germ |") {
		t.Error("athen structree should not contain phase_town_germ token (other civ)")
	}
}

func TestStructree_Rome_HasArmyCamp(t *testing.T) {
	skipIfNoGamedata(t)
	out := generateFor(t, "rome")
	// Army Camp is reachable only via the Promotion chain
	// (swordsman_b → _a → _e → legionary → builds army_camp). It is also
	// referenced as the <Fort> piece of wallset_siege but must appear as
	// a standalone TOWN-phase building because it has its own Trainer.
	if !strings.Contains(out.Structree, "Army Camp") && !strings.Contains(out.Structree, "Castra") {
		t.Error("rome structree missing Army Camp / Castra (Promotion edge or wallset-piece filter regression)")
	}
}

func TestStructree_Germ_TownAndCityNotEmpty(t *testing.T) {
	skipIfNoGamedata(t)
	out := generateFor(t, "germ")
	// Bug fix: BuildingPhase used to read Identity/RequiredTechnology
	// (which doesn't exist in R28 data), so all buildings collapsed to
	// Village. Now reads Identity/Requirements/Techs.
	if strings.Contains(out.Structree, "## TOWN PHASE\n\n*В этой фазе нет уникальных построек.*") {
		t.Error("germ TOWN PHASE block is empty (BuildingPhase regression)")
	}
	if strings.Contains(out.Structree, "## CITY PHASE\n\n*В этой фазе нет уникальных построек.*") {
		t.Error("germ CITY PHASE block is empty (BuildingPhase regression)")
	}
	// Sanity: Great Hall must be in TOWN PHASE block, not VILLAGE.
	townStart := strings.Index(out.Structree, "## TOWN PHASE")
	cityStart := strings.Index(out.Structree, "## CITY PHASE")
	villageStart := strings.Index(out.Structree, "## VILLAGE PHASE")
	if townStart < 0 || cityStart < 0 || villageStart < 0 {
		t.Fatal("missing phase headers")
	}
	townSection := out.Structree[townStart:cityStart]
	if !strings.Contains(townSection, "Great Hall") {
		t.Error("Great Hall expected in TOWN PHASE block of germ structree")
	}
	villageSection := out.Structree[villageStart:townStart]
	if strings.Contains(villageSection, "Great Hall") {
		t.Error("Great Hall must NOT be in VILLAGE PHASE block")
	}
}

// generateFor builds a full Output for the named civ using real gamedata.
func generateFor(t *testing.T, civ string) Output {
	t.Helper()
	layout := paths.Layout{Root: gamedataRoot()}
	resolver := newResolver(t)
	g := NewGenerator(layout, resolver)
	info, ok := civdata.ResolveCivInput(civ)
	if !ok {
		t.Fatalf("%s resolution failed", civ)
	}
	out, err := g.Generate(info)
	if err != nil {
		t.Fatalf("generate %s: %v", civ, err)
	}
	return out
}
