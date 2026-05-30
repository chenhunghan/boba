package box_test

import (
	"fmt"

	"github.com/chenhunghan/boba/box"
)

// A Box is a bordered region with optional labeled notches in its top
// edge. Render is pure: it returns a string of exactly width×height
// cells, ready to compose with lipgloss.JoinHorizontal / JoinVertical.
// Styling (BorderColor, FillColor, per-notch Style) is supplied by the
// caller — the zero value renders with the terminal's defaults.
func ExampleBox_Render() {
	b := box.Box{
		LeftNotches:  []box.Notch{{Text: "status"}},
		RightNotches: []box.Notch{{Text: "1", Badge: "q"}},
		Body:         "ready.",
	}

	out := b.Render(24, 4)

	lines := 0
	for _, r := range out {
		if r == '\n' {
			lines++
		}
	}
	fmt.Println(lines + 1) // height is exactly as requested
	// Output: 4
}
