package civdata

import "testing"

func TestResolveCivInput(t *testing.T) {
	cases := []struct {
		in   string
		code string
		ok   bool
	}{
		{"spart", "spart", true},
		{"спартанцы", "spart", true},
		{"German", "germ", true},
		{"Germans", "germ", true},
		{"Афиняне", "athen", true},
		{"Han", "han", true},
		{"карфагеняне", "cart", true},
		{"unknown", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got, ok := ResolveCivInput(tc.in)
		if ok != tc.ok {
			t.Errorf("ResolveCivInput(%q) ok=%v; want %v", tc.in, ok, tc.ok)
			continue
		}
		if ok && got.Code != tc.code {
			t.Errorf("ResolveCivInput(%q) = %q; want %q", tc.in, got.Code, tc.code)
		}
	}
}

func TestCivCode_FileNames(t *testing.T) {
	c, ok := ResolveCivInput("germ")
	if !ok {
		t.Fatal("germ not resolved")
	}
	if c.BaseName != "germans" {
		t.Errorf("BaseName = %q; want %q", c.BaseName, "germans")
	}
	if got := c.OverviewFile(); got != "germans_overview.md" {
		t.Errorf("OverviewFile() = %q", got)
	}
	if got := c.StructreeFile(); got != "germans_structree.md" {
		t.Errorf("StructreeFile() = %q", got)
	}
}

func TestCivCode_AllHaveBaseName(t *testing.T) {
	for _, c := range Civilizations {
		if c.BaseName == "" {
			t.Errorf("civ %q: BaseName empty", c.Code)
		}
	}
}
