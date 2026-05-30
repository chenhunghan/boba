package glyph

import "testing"

// TestSubscript spot-checks the digit conversion across single
// digits, multi-digit, zero, and negative numbers.
func TestSubscript(t *testing.T) {
	cases := map[int]string{
		0:   "₀",
		1:   "₁",
		9:   "₉",
		12:  "₁₂",
		100: "₁₀₀",
		-3:  "₋₃",
		-42: "₋₄₂",
	}
	for n, want := range cases {
		if got := Subscript(n); got != want {
			t.Errorf("Subscript(%d) = %q, want %q", n, got, want)
		}
	}
}
