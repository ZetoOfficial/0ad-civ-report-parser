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
