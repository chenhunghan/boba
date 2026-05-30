package glyph

import (
	"strconv"
	"strings"
)

// Superscript returns the Unicode-superscript form of n's decimal
// digits. Negative n yields a leading "⁻" minus sign followed by
// the digits. Useful for keyboard-shortcut hints in TUI labels —
// a raised glyph is visually distinct from the surrounding text
// without needing a separate color or font weight.
//
//	Superscript(1)   → "¹"
//	Superscript(12)  → "¹²"
//	Superscript(-3)  → "⁻³"
func Superscript(n int) string {
	digits := []rune("⁰¹²³⁴⁵⁶⁷⁸⁹")
	s := strconv.Itoa(n)
	var b strings.Builder
	for _, r := range s {
		if r == '-' {
			b.WriteRune('⁻')
			continue
		}
		b.WriteRune(digits[r-'0'])
	}
	return b.String()
}
