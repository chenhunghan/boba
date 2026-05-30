package glyph

import (
	"strconv"
	"strings"
)

// Subscript returns the Unicode-subscript form of n's decimal digits.
// Negative n yields a leading "₋" minus sign followed by the digits.
// Symmetric with Superscript.
//
//	Subscript(1)   → "₁"
//	Subscript(12)  → "₁₂"
//	Subscript(-3)  → "₋₃"
func Subscript(n int) string {
	digits := []rune("₀₁₂₃₄₅₆₇₈₉")
	s := strconv.Itoa(n)
	var b strings.Builder
	for _, r := range s {
		if r == '-' {
			b.WriteRune('₋')
			continue
		}
		b.WriteRune(digits[r-'0'])
	}
	return b.String()
}
