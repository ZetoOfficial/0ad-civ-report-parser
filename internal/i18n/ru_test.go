package i18n

import "testing"

func TestFormatNumber(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{100, "100"},
		{9.5, "9.5"},
		{1.1, "1.1"},
		{0.05, "0.05"},
		{0.15, "0.15"},
		// IEEE 754 noise: 9 * 1.2 == 10.799999999999999 in float64.
		{9 * 1.2, "10.8"},
		// Walk speed 8 * 1.35 RunMultiplier == 10.799999999999999.
		{8 * 1.35, "10.8"},
		// Rounding produces an integer result.
		{9 * 0.1 * 10, "9"},
		{-1.5, "-1.5"},
	}
	for _, c := range cases {
		t.Run(c.want, func(t *testing.T) {
			got := FormatNumber(c.in)
			if got != c.want {
				t.Errorf("FormatNumber(%v) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
