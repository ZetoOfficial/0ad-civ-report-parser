package skeleton

import (
	"strings"
	"testing"
)

func TestRender_Overview_Substitutions(t *testing.T) {
	d := Data{
		CivName:        "Spartans",
		CivCodeUpper:   "Spart",
		Date:           "2026-04-26",
		Lang:           "",
		IncludeHistory: false,
		IncludeIcons:   false,
		Body:           "BODY-MARKER",
	}
	out, err := Render("overview", d)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	musts := []string{
		"# Spartans (Spart) — Civilization Overview",
		"Сгенерировано 2026-04-26",
		"include_history=false",
		"lang=—",
		"BODY-MARKER",
	}
	for _, m := range musts {
		if !strings.Contains(out, m) {
			t.Errorf("Render(overview) missing %q in:\n%s", m, out)
		}
	}
}

func TestRender_Structree_LangSubstitution(t *testing.T) {
	d := Data{CivName: "Han Chinese", CivCodeUpper: "Han", Date: "2026-04-26", Lang: "ru", Body: "X"}
	out, err := Render("structree", d)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(out, "lang=ru") {
		t.Errorf("expected lang=ru in:\n%s", out)
	}
	if !strings.Contains(out, "# Han Chinese (Han) — Structure Tree") {
		t.Errorf("expected structree header; got:\n%s", out)
	}
}

func TestRender_Common_BodySlot(t *testing.T) {
	d := Data{Date: "2026-04-26", Body: "COMMON-MARKER"}
	out, err := Render("common", d)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(out, "# Common Reference") {
		t.Errorf("missing common header in:\n%s", out)
	}
	if !strings.Contains(out, "COMMON-MARKER") {
		t.Errorf("missing body in:\n%s", out)
	}
}

func TestRender_UnknownTemplate(t *testing.T) {
	_, err := Render("nope", Data{})
	if err == nil {
		t.Fatal("expected error for unknown template name")
	}
}
