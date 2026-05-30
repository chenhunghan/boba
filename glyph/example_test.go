package glyph_test

import (
	"fmt"

	"github.com/chenhunghan/boba/glyph"
)

func ExampleSuperscript() {
	fmt.Printf("x%s\n", glyph.Superscript(2))
	// Output: x²
}

func ExampleSubscript() {
	fmt.Printf("H%sO\n", glyph.Subscript(2))
	// Output: H₂O
}
