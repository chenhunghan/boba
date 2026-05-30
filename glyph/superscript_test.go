package glyph

import "testing"

// TestSuperscript spot-checks the digit conversion across single
// digits, multi-digit, zero, and negative numbers.
func TestSuperscript(t *testing.T) {
	cases := map[int]string{
		0:   "⁰",
		1:   "¹",
		9:   "⁹",
		12:  "¹²",
		100: "¹⁰⁰",
		-3:  "⁻³",
		-42: "⁻⁴²",
	}
	for n, want := range cases {
		if got := Superscript(n); got != want {
			t.Errorf("Superscript(%d) = %q, want %q", n, got, want)
		}
	}
}
